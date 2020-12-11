package redisdump

const (
	binary = "redis-cli"
)

type Options struct {
	Flags          *Flags
	AdditionalArgs []string
}

type Flags struct {
	Host       string `flag:"-h"`
	Port       int    `flag:"-p"`
	ResultFile string `flag:"--rdb"`
	Password   string `flag:"-a"`
}
