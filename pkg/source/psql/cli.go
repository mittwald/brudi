package psql

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
	Dbname              string `flag:"--dbname="`
	FieldSeparator      string `flag:"--field-separator="`
	File                string `flag:"--file="`
	Host                string `flag:"--host="                 validate:"min=1"`
	LogFile             string `flag:"--log-file="`
	Output              string `flag:"--output="`
	Password            string `env:"PGPASSWORD"               flag:"-"`
	Pset                string `flag:"--pset="`
	RecordSeparator     string `flag:"--record-separator="`
	Set                 string `flag:"--set="`
	TableAttr           string `flag:"--table-attr="`
	User                string `flag:"--username="`
	Variable            string `flag:"--variable="`
	Port                int    `flag:"--port="`
	EchoAll             bool   `flag:"--echo-all"`
	EchoHidden          bool   `flag:"--echo-hidden"`
	EchoQueries         bool   `flag:"--echo-queries"`
	Expanded            bool   `flag:"--expanded"`
	FieldSeparatorZero  bool   `flag:"--field-separator-zero"`
	HTML                bool   `flag:"--html"`
	List                bool   `flag:"--list"`
	NoAllign            bool   `flag:"--no-align"`
	NoPassword          bool   `flag:"--no-password"`
	NoPsqlrc            bool   `flag:"--no-psqlrc"`
	NoReafline          bool   `flag:"--no-readline"`
	Quiet               bool   `flag:"--quiet"`
	RecordSeparatorZero bool   `flag:"--record-separator-zero"`
	SingleLine          bool   `flag:"--single-line"`
	SingleStep          bool   `flag:"--single-step"`
	SingleTransaction   bool   `flag:"--single-transaction"`
	TuplesOnly          bool   `flag:"--tuples-only"`
	Version             bool   `flsg:"--version"`
}
