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
	Raw        bool   `flag:"--raw"`
	NoRaw      bool   `flag:"--no-raw"`
	Csv        bool   `flag:"--csv"`
	Stat       bool   `flag:"--stat"`
	ResultFile string `flag:"--rdb"`
}
