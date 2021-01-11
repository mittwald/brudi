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
	BinaryUpgrade         bool   `flag:"--binary-upgrade"`
	Blobs                 bool   `flag:"--blobs"`
	Clean                 bool   `flag:"--clean"`
	ColumnInserts         bool   `flag:"--column-inserts"`
	Compress              int    `flag:"--compress="`
	Create                bool   `flag:"--create"`
	DataOnly              bool   `flag:"--data-only"`
	DBName                string `flag:"--dbname="`
	DisableTriggers       bool   `flag:"--disable-triggers"`
	ExitOnError           bool   `flag:"--exit-on-error"`
	File                  string `flag:"--file="`
	Format                string `flag:"--format="`
	Function              string `flag:"--function="`
	Host                  string `flag:"--host="`
	IgnoreVersion         bool   `flag:"--ignore-version"`
	Index                 string `flag:"--index="`
	Inserts               bool   `flag:"--inserts"`
	Jobs                  int    `flag:"--jobs="`
	List                  bool   `flag:"--list"`
	ListFile              string `flag:"--use-List="`
	LoadViaPartitionRoot  bool   `flag:"--load-via-partition-root"`
	LockWaitTimeout       string `flag:"--lock-wait-timeout="`
	NoACL                 bool   `flag:"--no-acl"`
	NoComments            bool   `flag:"--no-comments"`
	NoDataForFailedTables bool   `flag:"--no-data-for-failed-tables"`
	NoOwner               bool   `flag:"--no-owner"`
	NoPassword            bool   `flag:"--no-password"`
	NoPrivileges          bool   `flag:"--no-privileges"`
	NoReconnect           bool   `flag:"--no-reconnect"`
	NoSecurityLabels      bool   `flag:"--no-security-labels"`
	NoTablespaces         bool   `flag:"--no-tablespaces"`
	// unfortunately pg_dump has no cli-option to specify the password
	// therefore we have to workaround by setting the corresponding password env-var
	Password                   string `flag:"-" env:"PGPASSWORD"`
	Port                       int    `flag:"--port="`
	QuoteAllIdentifiers        bool   `flag:"--quote-all-identifiers"`
	Role                       string `flag:"--role="`
	Schema                     string `flag:"--schema="`
	SchemaOnly                 bool   `flag:"--schema-only"`
	Section                    string `flag:"--section="`
	Superuser                  string `flag:"--superuser="`
	Table                      string `flag:"--table="`
	Trigger                    string `flag:"--trigger="`
	Username                   string `flag:"--username="`
	UseSetSessionAuthorization bool   `flag:"--use-set-session-authorization"`
	Verbose                    bool   `flag:"--verbose"`
}
