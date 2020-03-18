package cli

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"reflect"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

const flagTag = "flag"

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
		log.WithError(err).WithField("output", string(out)).Error("failed to execute command")
		return out, err
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
