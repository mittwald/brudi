package cli

import (
	"io"
	"sync"
)

const GzipSuffix = ".gz"
const DoStdinBackupKey = "doPipingBackup"

type CommandType struct {
	Binary      string
	Command     string
	Args        []string
	Pipe        io.Reader  // TODO: Remove when --stdin-command was added to restic
	PipeReady   *sync.Cond // TODO: Remove when --stdin-command was added to restic
	ReadingDone chan bool  // TODO: Remove when --stdin-command was added to restic
	Nice        *int       // https://linux.die.net/man/1/nice
	IONice      *int       // https://linux.die.net/man/1/ionice
}

type PipedCommandsPids struct {
	Pid1 int
	Pid2 int
}
