package pgrestore

const (
	binary = "psql"
)

type Options struct {
	Flags          *Flags
	Command        string
	SourceFile     string
	AdditionalArgs []string
}

type Flags struct {
	Command             string `flag:"--command="`
	Dbname              string `flag:"--dbname=" validate:"min=1"`
	EchoAll             bool   `flag:"--echo-all"`
	EchoHidden          bool   `flag:"--echo-hidden"`
	EchoQueries         bool   `flag:"--echo-queries"`
	Expanded            bool   `flag:"--expanded"`
	FieldSeparator      string `flag:"--field-separator="`
	FieldSeparatorZero  bool   `flag:"--field-separator-zero"`
	File                string `flag:"--file="`
	Host                string `flag:"--host=" validate:"min=1"`
	HTML                bool   `flag:"--html"`
	List                bool   `flag:"--list"`
	LogFile             string `flag:"--log-file="`
	NoAllign            bool   `flag:"--no-align"`
	NoPassword          bool   `flag:"--no-password"`
	NoPsqlrc            bool   `flag:"--no-psqlrc"`
	NoReafline          bool   `flag:"--no-readline"`
	Output              string `flag:"--output="`
	Password            bool   `flag:"--password"`
	Port                int    `flag:"--port="`
	Pset                string `flag:"--pset="`
	Quiet               bool   `flag:"--quiet"`
	RecordSeparator     string `flag:"--record-separator="`
	RecordSeparatorZero bool   `flag:"--record-separator-zero"`
	SingleLine          bool   `flag:"--single-line"`
	SingleStep          bool   `flag:"--single-step"`
	SingleTransaction   bool   `flag:"--single-transaction"`
	Set                 string `flag:"--set="`
	TableAttr           string `flag:"--table-attr="`
	TuplesOnly          bool   `flag:"--tuples-only"`
	User                string `flag:"--username="`
	Variable            string `flag:"--variable="`
	Version             bool   `flsg:"--version"`
}
