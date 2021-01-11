package cli

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"syscall"
	"time"

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
func StructToCLI(optionStruct interface{}) []string { // nolint: gocyclo
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
		commandLine = append([]string{"nice", fmt.Sprintf("-n%d", *cmd.Nice)}, commandLine...)
	}

	if cmd.IONice != nil {
		commandLine = append([]string{"ionice", fmt.Sprintf("-c%d", *cmd.IONice)}, commandLine...)
	}

	return commandLine
}

// RunWithTimeout executes the given binary within a max execution time
func RunWithTimeout(runContext context.Context, cmd CommandType, timeout time.Duration) ([]byte, error) {
	var ctx context.Context

	if runContext != nil {
		ctx = runContext
	} else {
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
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
		out, err = exec.CommandContext(ctx, commandLine[0], commandLine[1:]...).CombinedOutput()
		if ctx.Err() != nil {
			return out, fmt.Errorf("failed to execute command: timed out or canceled")
		}
	} else {
		out, err = exec.Command(commandLine[0], commandLine[1:]...).CombinedOutput()
	}
	if err != nil {
		return out, fmt.Errorf("failed to execute command: %s", err)
	}

	log.WithField("command", strings.Join(commandLine, " ")).Debug("successfully executed command")
	return out, nil
}

// RunPipedWithTimeout executes "RunPiped" within a max execution time
func RunPipedWithTimeout(
	runContext context.Context,
	cmd1, cmd2 CommandType,
	timeout time.Duration, pids *PipedCommandsPids) ([]byte, error) {
	var ctx context.Context

	if runContext != nil {
		ctx = runContext
	} else {
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return RunPiped(ctx, cmd1, cmd2, pids)
}

// RunPiped executes cmd1 and pipes its output to cmd2. Will return the output of cmd2
func RunPiped(ctx context.Context, cmd1, cmd2 CommandType, pids *PipedCommandsPids) ([]byte, error) {
	var errors []string
	var err error
	var cmd1Exec, cmd2Exec *exec.Cmd
	var out bytes.Buffer
	cmdLine1 := ParseCommandLine(cmd1)
	cmdLine2 := ParseCommandLine(cmd2)
	log.WithField("command",
		fmt.Sprintf(
			"%s | %s",
			strings.Join(cmdLine1, " "),
			strings.Join(cmdLine2, " "),
		),
	).Debug("executing command")

	if ctx != nil {
		cmd1Exec = exec.CommandContext(ctx, cmdLine1[0], cmdLine1[1:]...)
		cmd2Exec = exec.CommandContext(ctx, cmdLine2[0], cmdLine2[1:]...)
	} else {
		cmd1Exec = exec.Command(cmdLine1[0], cmdLine1[1:]...)
		cmd2Exec = exec.Command(cmdLine2[0], cmdLine2[1:]...)
	}

	cmd2Exec.Stdin, err = cmd1Exec.StdoutPipe()
	if err != nil {
		return nil, err
	}

	cmd1Exec.Stderr = &out
	cmd2Exec.Stdout = &out
	cmd2Exec.Stderr = &out
	err = cmd2Exec.Start()
	if err != nil {
		errors = append(errors, err.Error())
	}

	err = cmd1Exec.Start()
	if err != nil {
		errors = append(errors, err.Error())
	}

	if pids != nil {
		if cmd1Exec.Process.Pid != 0 {
			pids.Pid1 = cmd1Exec.Process.Pid
		}
		if cmd2Exec.Process.Pid != 0 {
			pids.Pid2 = cmd2Exec.Process.Pid
		}
	}

	err = cmd1Exec.Wait()
	if err != nil {
		msg, ok := err.(*exec.ExitError)
		if !ok || !(cmd1.Binary == "tar" && msg.Sys().(syscall.WaitStatus).ExitStatus() == 1) { // ignore tar exit-code of 1
			errors = append(errors, err.Error())
		}
	}

	err = cmd2Exec.Wait()
	if err != nil {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		return out.Bytes(), fmt.Errorf(strings.Join(errors, "\n"))
	}

	log.WithField("command",
		fmt.Sprintf(
			"%s | %s",
			strings.Join(cmdLine1, " "),
			strings.Join(cmdLine2, " "),
		),
	).Debug("successfully executed command")

	return out.Bytes(), nil
}

// GzipFile compresses a file with gzip and returns the path of the created archive
func GzipFile(fileName string) (string, error) {
	var err error

	//open input file
	var inFile *os.File
	inFile, err = os.Open(fileName)
	if err != nil {
		return "", errors.WithStack(err)
	}

	// read file content
	var content []byte
	reader := bufio.NewReader(inFile)
	content, err = ioutil.ReadAll(reader)
	if err != nil {
		return "", errors.WithStack(err)
	}

	// open output file
	var outFile *os.File
	outName := fmt.Sprintf("%s.gz", fileName)
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
	var archiveWriter *gzip.Writer
	archiveWriter = gzip.NewWriter(outFile)
	_, err = archiveWriter.Write(content)
	if err != nil {
		return "", errors.WithStack(err)
	}
	err = archiveWriter.Close()
	if err != nil {
		log.WithError(err).Error("failed to close archive reader")
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
	outFile, err = os.Create(outName)
	if err != nil {
		return "", err
	}
	defer func() {
		outErr := outFile.Close()
		if outErr != nil {
			log.WithError(outErr).Errorf("failed to close output file %s", outName)
		}
	}()

	// write unzipped file to file system
	_, err = io.Copy(outFile, archiveReader)
	if err != nil {
		return "", errors.WithStack(err)
	}
	extractedName := archiveReader.Name
	return extractedName, nil
}
