package tar

type Flags struct {
	Create         bool     `flag:"-c"`
	Extract        bool     `flag:"-x"`
	Overwrite      bool     `flag:"--overwrite"`
	NoOverwriteDir bool     `flag:"--no-overwrite-dir"`
	Warning        []string `flag:"--warning"`
	Exclude        []string `flag:"--exclude"`
	Target         string   `flag:"-C"`
	File           string   `flag:"-f"`
}

type Options struct {
	Flags *Flags
	Paths []string
}
