package pgrestore

const (
	binary = "pg_restore"
)

type Options struct {
	Flags          *Flags
	Command        string
	SourceFile     string
	PGRestore      bool
	AdditionalArgs []string
}

type Flags struct {
	DBName          string `flag:"--dbname="`
	File            string `flag:"--file="`
	Format          string `flag:"--format="`
	Function        string `flag:"--function="`
	Host            string `flag:"--host="`
	Index           string `flag:"--index="`
	ListFile        string `flag:"--use-List="`
	LockWaitTimeout string `flag:"--lock-wait-timeout="`
	// unfortunately pg_restore has no cli-option to specify the password
	// therefore we have to workaround by setting the corresponding password env-var
	Password                   string `env:"PGPASSWORD"                       flag:"-"`
	Role                       string `flag:"--role="`
	Schema                     string `flag:"--schema="`
	Section                    string `flag:"--section="`
	Superuser                  string `flag:"--superuser="`
	Table                      string `flag:"--table="`
	Trigger                    string `flag:"--trigger="`
	Username                   string `flag:"--username="`
	Compress                   int    `flag:"--compress="`
	Jobs                       int    `flag:"--jobs="`
	Port                       int    `flag:"--port="`
	BinaryUpgrade              bool   `flag:"--binary-upgrade"`
	Blobs                      bool   `flag:"--blobs"`
	Clean                      bool   `flag:"--clean"`
	ColumnInserts              bool   `flag:"--column-inserts"`
	Create                     bool   `flag:"--create"`
	DataOnly                   bool   `flag:"--data-only"`
	DisableTriggers            bool   `flag:"--disable-triggers"`
	ExitOnError                bool   `flag:"--exit-on-error"`
	IgnoreVersion              bool   `flag:"--ignore-version"`
	Inserts                    bool   `flag:"--inserts"`
	List                       bool   `flag:"--list"`
	LoadViaPartitionRoot       bool   `flag:"--load-via-partition-root"`
	NoACL                      bool   `flag:"--no-acl"`
	NoComments                 bool   `flag:"--no-comments"`
	NoDataForFailedTables      bool   `flag:"--no-data-for-failed-tables"`
	NoOwner                    bool   `flag:"--no-owner"`
	NoPassword                 bool   `flag:"--no-password"`
	NoPrivileges               bool   `flag:"--no-privileges"`
	NoReconnect                bool   `flag:"--no-reconnect"`
	NoSecurityLabels           bool   `flag:"--no-security-labels"`
	NoTablespaces              bool   `flag:"--no-tablespaces"`
	QuoteAllIdentifiers        bool   `flag:"--quote-all-identifiers"`
	SchemaOnly                 bool   `flag:"--schema-only"`
	UseSetSessionAuthorization bool   `flag:"--use-set-session-authorization"`
	Verbose                    bool   `flag:"--verbose"`
}
