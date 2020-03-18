package pgdump

const (
	binary = "pg_dump"
)

type Options struct {
	Flags          *Flags
	AdditionalArgs []string
}

type Flags struct {
	File                       string `flag:"--file="`
	Format                     string `flag:"--format="`
	Jobs                       int    `flag:"--jobs="`
	Verbose                    bool   `flag:"--verbose"`
	Compress                   int    `flag:"--compress="`
	LockWaitTimeout            string `flag:"--lock-wait-timeout="`
	NoSync                     bool   `flag:"--no-sync"`
	DataOnly                   bool   `flag:"--data-only"`
	Blobs                      bool   `flag:"--blobs"`
	Clean                      bool   `flag:"--clean"`
	Create                     bool   `flag:"--create"`
	Encoding                   string `flag:"--encoding="`
	Schema                     string `flag:"--schema="`
	ExcludeSchema              string `flag:"--exclude-schema="`
	NoOwner                    bool   `flag:"--no-owner"`
	SchemaOnly                 bool   `flag:"--schema-only"`
	Superuser                  string `flag:"--superuser="`
	Table                      string `flag:"--table="`
	ExcludeTable               string `flag:"--exclude-table="`
	NoPrivileges               bool   `flag:"--no-privileges"`
	BinaryUpgrade              bool   `flag:"--binary-upgrade"`
	ColumnInserts              bool   `flag:"--column-inserts"`
	DisableDollarQuoting       bool   `flag:"--disable-dollar-quoting"`
	DisableTriggers            bool   `flag:"--disable-triggers"`
	EnableRowSecurity          bool   `flag:"--enable-row-security"`
	ExcludeTableData           string `flag:"--exclude-table-data="`
	ExtraFloatDigits           int    `flag:"--extra-float-digits="`
	IfExists                   bool   `flag:"--if-exists"`
	Inserts                    bool   `flag:"--inserts"`
	LoadViaPartitionRoot       bool   `flag:"--load-via-partition-root"`
	NoComments                 bool   `flag:"--no-comments"`
	NoPublications             bool   `flag:"--no-publications"`
	NoSecurityLabels           bool   `flag:"--no-security-labels"`
	NoSubscriptions            bool   `flag:"--no-subscriptions"`
	NoSynchronizedSnapshots    bool   `flag:"--no-synchronized-snapshots"`
	NoTablespaces              bool   `flag:"--no-tablespaces"`
	NoUnloggedTableData        bool   `flag:"--no-unlogged-table-data"`
	OnConflictDoNothing        bool   `flag:"--on-conflict-do-nothing"`
	QuoteAllIdentifiers        bool   `flag:"--quote-all-identifiers"`
	RowsPerInsert              int    `flag:"--rows-per-insert="`
	Section                    string `flag:"--section="`
	SerializableDeferrable     bool   `flag:"--serializable-deferrable"`
	Snapshots                  string `flag:"--snapshot="`
	StrictNames                bool   `flag:"--strict-names"`
	UseSetSessionAuthorization bool   `flag:"--use-set-session-authorization"`
	DBName                     string `flag:"--dbname="`
	Host                       string `flag:"--host="`
	Port                       int    `flag:"--port="`
	Username                   string `flag:"--username="`
	NoPassword                 bool   `flag:"--no-password"`
	// unfortunately pg_dump has no cli-option to specify the password
	// therefore we have to workaround by setting the corresponding password env-var
	Password string `flag:"-" env:"PGPASSWORD"`
	Role     string `flag:"--role="`
}
