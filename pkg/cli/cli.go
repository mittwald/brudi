package cli

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const flagTag = "flag"
const gzipType = "application/x-gzip"

// includeFlag returns an string slice of [<flag>, <val>], or [<val>]
func includeFlag(flag, val string) []string {
	var cmd []string
	if flag != "" {
		if strings.HasSuffix(flag, "=") {
			cmd = append(cmd, flag+val)
		} else {
			cmd = append(cmd, flag, val)
		}
	} else {
		cmd = append(cmd, val)
	}
	return cmd
}

// StructToCLI builds and returns a commandline string-array from a struct.
//
//	Example structs:
//	----------------------------------------------------
//		type Example struct {
//			StringOption string   `flag:"--string"`
//			SliceOption  []string `flag:"--slice"`
//			IntOption    int      `flag:"--int"`
//			BoolOption1  bool     `flag:"--bool1"`
//			BoolOption2  bool     `flag:"--bool2"`
//			StructOption *ChildExample
//		}
//
//		type ChildExample struct {
//			StringOption string `flag:"--child-string"`
//			BoolOption   bool   `flag:"--child-bool"`
//		}
//
//	Example Data:
//	----------------------------------------------------
//		child := ChildExample{
//			StringOption: "child string",
//			BoolOption: true,
//		}
//
//		example := Example{
//			StringOption: "string1"
//			SliceOption: []string{
//				"slice1",
//				"slice2",
//			},
//			IntOption: 42,
//			BoolOption1: true,
//			BoolOption2: false,
//			StructOption: &child,
//		}
//
//		cli := parseTypeCmd(&example)
//
//	Example Result:
//	----------------------------------------------------
//		cli = [
//				"--string", "string1",
//				"--slice", "slice1",
//				"--slice", "slice2",
//				"--int", "42",
//				"--bool1",
//				"--child-string",
//				"child string",
//				"--child-bool"
//			]
//
//	Notice:
//	----------------------------------------------------
//		Zero values (0, "", nil, false) and "-" will be ignored
//
//nolint:gocognit,cyclop // refactor this at some point
func StructToCLI(optionStruct interface{}) []string {
	if optionStruct == reflect.Zero(reflect.TypeOf(optionStruct)).Interface() {
		return nil
	}
	var cmd []string

	structElem := reflect.ValueOf(optionStruct).Elem()
	for i := 0; i < structElem.NumField(); i++ {
		field := structElem.Field(i)
		fieldVal := field.Interface()
		flag := structElem.Type().Field(i).Tag.Get(flagTag)

		if flag == "-" {
			continue
		}

		switch t := fieldVal.(type) {
		case int:
			if t == 0 {
				break
			}
			cmd = append(cmd, includeFlag(flag, fmt.Sprint(t))...)
		case uint:
			if t == 0 {
				break
			}
			cmd = append(cmd, includeFlag(flag, fmt.Sprint(t))...)
		case int32:
			if t == 0 {
				break
			}
			cmd = append(cmd, includeFlag(flag, fmt.Sprint(t))...)
		case uint32:
			if t == 0 {
				break
			}
			cmd = append(cmd, includeFlag(flag, fmt.Sprint(t))...)
		case string:
			if t == "" {
				break
			}
			cmd = append(cmd, includeFlag(flag, t)...)
		case []string:
			for _, itm := range t {
				cmd = append(cmd, includeFlag(flag, itm)...)
			}
		case bool:
			if !t {
				break
			}
			cmd = append(cmd, flag)
		case interface{}:
			iCmd := StructToCLI(t)
			if len(iCmd) == 0 {
				break
			}
			cmd = append(cmd, iCmd...)
		}
	}

	return cmd
}

func ParseCommandLine(cmd CommandType) []string {
	commandLine := cmd.Args

	if cmd.Command != "" {
		commandLine = append([]string{cmd.Command}, commandLine...)
	}

	if cmd.Binary != "" {
		commandLine = append([]string{cmd.Binary}, commandLine...)
	}

	if cmd.Nice != nil {
		commandLine = append(
			[]string{
				"nice",
				fmt.Sprintf("-n%d", *cmd.Nice),
			}, commandLine...,
		)
	}

	if cmd.IONice != nil {
		commandLine = append(
			[]string{
				"ionice",
				fmt.Sprintf("-c%d", *cmd.IONice),
			}, commandLine...,
		)
	}

	return commandLine
}

// RunWithTimeout executes the given binary within a max execution time
func RunWithTimeout(runContext context.Context, cmd CommandType, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(runContext, timeout)
	defer cancel()

	return Run(ctx, cmd)
}

// Run executes the given binary
func Run(ctx context.Context, cmd CommandType) ([]byte, error) {
	var out []byte
	var err error
	commandLine := ParseCommandLine(cmd)
	log.WithField("command", strings.Join(commandLine, " ")).Debug("executing command")
	if ctx != nil {
		out, err = exec.CommandContext(ctx, commandLine[0], commandLine[1:]...).CombinedOutput() //nolint: gosec
		if ctx.Err() != nil {
			return out, fmt.Errorf("failed to execute command: timed out or canceled")
		}
	} else {
		out, err = exec.Command(commandLine[0], commandLine[1:]...).CombinedOutput() //nolint: gosec
	}
	if err != nil {
		return out, fmt.Errorf("failed to execute command: %s", err)
	}

	log.WithField("command", strings.Join(commandLine, " ")).Debug("successfully executed command")
	return out, nil
}

// GzipFile compresses a file with gzip and returns the path of the created archive
func GzipFile(fileName string) (string, error) {
	var err error

	// open input file
	var inFile *os.File
	inFile, err = os.Open(fileName)
	if err != nil {
		return "", errors.WithStack(err)
	}

	// read file content
	var content []byte
	reader := bufio.NewReader(inFile)
	content, err = io.ReadAll(reader)
	if err != nil {
		return "", errors.WithStack(err)
	}

	// open output file
	var outFile *os.File
	outName := fmt.Sprintf("%s%s", fileName, GzipSuffix)
	outFile, err = os.Create(outName)
	if err != nil {
		return "", errors.WithStack(err)
	}
	defer func() {
		outErr := outFile.Close()
		if outErr != nil {
			log.WithError(outErr).Errorf("failed to close output file %s", outName)
		}
	}()

	// write compressed content to file
	archiveWriter := gzip.NewWriter(outFile)
	archiveWriter.Name = fileName
	_, err = archiveWriter.Write(content)
	if err != nil {
		return "", errors.WithStack(err)
	}
	err = archiveWriter.Close()
	if err != nil {
		log.WithError(err).Error("failed to close archive reader")
	}

	// remove uncompressed source backup
	err = os.Remove(fileName)
	if err != nil {
		log.WithError(err).Error("failed to remove uncompressed backup file")
	}

	return outName, nil
}

// CheckAndGunzipFile checks if a file is gzipped and extracts it in that case...
// ... it also returns the name of the unzipped file
func CheckAndGunzipFile(fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", errors.WithStack(err)
	}
	defer func() {
		fileErr := file.Close()
		if fileErr != nil {
			log.WithError(fileErr).Errorf("failed to close source file %s", fileName)
		}
	}()

	// read first 512 bytes for http.DetectContentType
	headerBytes := make([]byte, 512)
	_, err = file.Read(headerBytes)
	if err != nil {
		return "", errors.WithStack(err)
	}

	// check if file is gzipped
	fileType := http.DetectContentType(headerBytes)
	if fileType != gzipType {
		return fileName, nil
	}
	// open gzipped file
	archive, archErr := os.Open(fileName)
	if archErr != nil {
		return "", errors.WithStack(err)
	}
	defer func() {
		archDeferredErr := archive.Close()
		if archDeferredErr != nil {
			log.WithError(archDeferredErr).Errorf("failed to close archive file %s", fileName)
		}
	}()

	// unzip gzipped file
	var archiveReader *gzip.Reader
	archiveReader, err = gzip.NewReader(archive)
	if err != nil {
		return "", errors.WithStack(err)
	}
	defer func() {
		readerErr := archiveReader.Close()
		if readerErr != nil {
			log.WithError(readerErr).Error("failed to close archive reader")
		}
	}()

	// open output file
	var outFile *os.File
	outName := archiveReader.Name
	// if archive header isn't set properly attempt to salvage by using filename without '.gz'
	if outName == "" {
		outName = strings.TrimRight(fileName, GzipSuffix)
	}
	outFile, err = os.Create(outName)
	if err != nil {
		return "", errors.WithStack(err)
	}
	defer func() {
		outErr := outFile.Close()
		if outErr != nil {
			log.WithError(outErr).Errorf("failed to close output file %s", outName)
		}
	}()

	// write unzipped file to file system
	_, err = io.Copy(outFile, archiveReader) //nolint: gosec // we work with potentially large backups
	if err != nil {
		return "", errors.WithStack(err)
	}
	extractedName := archiveReader.Name
	return extractedName, nil
}
