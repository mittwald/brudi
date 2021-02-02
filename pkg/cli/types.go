package cli

const GzipSuffix = ".gz"

type CommandType struct {
	Binary  string
	Command string
	Args    []string
	Nice    *int // https://linux.die.net/man/1/nice
	IONice  *int // https://linux.die.net/man/1/ionice
}

type PipedCommandsPids struct {
	Pid1 int
	Pid2 int
}
