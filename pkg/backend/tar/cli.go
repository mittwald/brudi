package tar

const (
	binary = "tar"
)

type Flags struct {
	Create          bool     `flag:"-c"`
	Gzip            bool     `flag:"-z"`
	Extract         bool     `flag:"-x"`
	StripComponents int      `flag:"--strip-components="`
	Overwrite       bool     `flag:"--overwrite"`
	NoOverwriteDir  bool     `flag:"--no-overwrite-dir"`
	Warning         []string `flag:"--warning"`
	Exclude         []string `flag:"--exclude"`
	Target          string   `flag:"-C"`
	File            string   `flag:"-f"`
}

type Options struct {
	Flags *Flags
	Paths []string
}
