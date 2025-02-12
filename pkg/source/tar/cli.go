package tar

const (
	binary = "tar"
)

type Options struct {
	Flags          *Flags
	AdditionalArgs []string
	Paths          []string
}

type Flags struct {
	Target          string   `flag:"-C"`
	File            string   `flag:"-f"                  validate:"min=1"`
	Warning         []string `flag:"--warning"`
	Exclude         []string `flag:"--exclude"`
	StripComponents int      `flag:"--strip-components="`
	Create          bool     `flag:"-c"`
	Gzip            bool     `flag:"-z"`
	Extract         bool     `flag:"-x"`
	Overwrite       bool     `flag:"--overwrite"`
	NoOverwriteDir  bool     `flag:"--no-overwrite-dir"`
}
