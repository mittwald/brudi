package mysqldump

const (
	binary = "mysqldump"
)

type Flags struct {
	DefaultsFile string `flag:"--defaults-file="`
}
