package mysqldump

const (
	binary = "mariadb-dump"
	// binary = "mysqldump"
)

type Options struct {
	Flags          *Flags
	AdditionalArgs []string
}

type Flags struct {
	BindAddress                string   `flag:"--bind-address="`
	CharacterSetsDir           string   `flag:"--character-sets-dir="`
	CompressAlgorithms         string   `flag:"--compression-algorithms="`
	Debug                      string   `flag:"--debug="`
	DefaultAuth                string   `flag:"--default-auth="`
	DefaultCharacterSet        string   `flag:"--default-character-set="`
	DefaultsExtraFile          string   `flag:"--defaults-extra-file="`
	DefaultsFile               string   `flag:"--defaults-file="`
	DefaultsGroupSuffix        string   `flag:"--defaults-group-suffix="`
	FieldsEnclosedBy           string   `flag:"--fields-enclosed-by="`
	FieldsEscapedBy            string   `flag:"--fields-escaped-by="`
	FieldsOptionallyEnclosedBy string   `flag:"--fields-optionally-enclosed-by="`
	FieldsTerminatedBy         string   `flag:"--fields-terminated-by="`
	Host                       string   `flag:"--host=" validate:"min=1"`
	IgnoreError                string   `flag:"--ignore-error="`
	IgnoreTable                string   `flag:"--ignore-table="`
	LinesTerminatedBy          string   `flag:"--lines-terminated-by="`
	LogError                   string   `flag:"--log-error="`
	LoginPath                  string   `flag:"--login-path="`
	MasterData                 string   `flag:"--master-data="`
	MaxAllowedPacket           string   `flag:"--max-allowed-packet="`
	NetBufferLength            string   `flag:"--net-buffer-length="`
	Password                   string   `flag:"--password="`
	PluginDir                  string   `flag:"--plugin-dir="`
	Protocol                   string   `flag:"--protocol="`
	ResultFile                 string   `flag:"--result-file=" validate:"min=1"`
	ServerPublicKeyPath        string   `flag:"--server-public-key-path="`
	SharedMemoryBaseName       string   `flag:"--shared-memory-base-name="`
	ShowCreateSkipSecondary    string   `flag:"--show-create-skip-secondary-engine="`
	Socket                     string   `flag:"--socket="`
	SslCa                      string   `flag:"--ssl-ca="`
	SslCaPath                  string   `flag:"--ssl-capath="`
	SslCert                    string   `flag:"--ssl-cert="`
	SslCipher                  string   `flag:"--ssl-cipher="`
	SslCrl                     string   `flag:"--ssl-crl="`
	SslCrlPath                 string   `flag:"--ssl-crlpath="`
	SslKey                     string   `flag:"--ssl-key="`
	SkipSsl                    bool     `flag:"--skip-ssl"`
	Tab                        string   `flag:"--tab="`
	TLSCipherSuites            string   `flag:"--tls-ciphersuites="`
	TLSVersion                 string   `flag:"--tls-version="`
	User                       string   `flag:"--user="`
	Where                      string   `flag:"--where"`
	Databases                  []string `flag:"--databases"`
	Tables                     []string `flag:"--tables"`
	Port                       int      `flag:"--port="`
	ZstdCompressionLevel       int      `flag:"--zstd-compression-level"`
	AddDropDatabase            bool     `flag:"--add-drop-database"`
	AddDropTable               bool     `flag:"--add-drop-table"`
	AddDropTrigger             bool     `flag:"--add-drop-trigger"`
	AddLocks                   bool     `flag:"--add-locks"`
	AllDatabases               bool     `flag:"--all-databases"`
	AllowKeywords              bool     `flag:"--allow-keywords"`
	ApplySlaveStatements       bool     `flag:"--apply-slave-statements"`
	ColumnStatistics           bool     `flag:"--column-statistics"`
	Comments                   bool     `flag:"--comments"`
	Compact                    bool     `flag:"--compact"`
	Compatible                 bool     `flag:"--compatible"`
	CompleteInsert             bool     `flag:"--complete-insert"`
	Compress                   bool     `flag:"--compress"`
	CreateOptions              bool     `flag:"--create-options"`
	DebugCheck                 bool     `flag:"--debug-check"`
	DebugInfo                  bool     `flag:"--debug-info"`
	DeleteMasterLogs           bool     `flag:"--delete-master-logs"`
	DisableKeys                bool     `flag:"--disable-keys"`
	DumpDate                   bool     `flag:"--dump-date"`
	DumpSlave                  bool     `flag:"--dump-slave"`
	EnableCleartextPlugin      bool     `flag:"--enable-cleartext-plugin"`
	Events                     bool     `flag:"--events"`
	ExtendedInsert             bool     `flag:"--extended-insert"`
	FlushLogs                  bool     `flag:"--flush-logs"`
	FlushPrivileges            bool     `flag:"--flush-privileges"`
	Force                      bool     `flag:"--force"`
	GetServerPublicKey         bool     `flag:"--get-server-public-key"`
	HexBlog                    bool     `flag:"--hex-blob"`
	IncludeMasterHostPort      bool     `flag:"--include-master-host-port"`
	InsertIgnore               bool     `flag:"--insert-ignore"`
	LockAllTables              bool     `flag:"--lock-all-tables"`
	LockTables                 bool     `flag:"--lock-tables"`
	NetworkTimeout             bool     `flag:"--network-timeout"`
	NoAutocommit               bool     `flag:"--no-autocommit"`
	NoCreateDB                 bool     `flag:"--no-create-db"`
	NoCreateInfo               bool     `flag:"--no-create-info"`
	NoData                     bool     `flag:"--no-data"`
	NoDefaults                 bool     `flag:"--no-defaults"`
	NoSetNames                 bool     `flag:"--no-set-names"`
	NoTablespaces              bool     `flag:"--no-tablespaces"`
	Opt                        bool     `flag:"--opt"`
	OrderByPrimary             bool     `flag:"--order-by-primary"`
	PrintDefaults              bool     `flag:"--print-defaults"`
	Quick                      bool     `flag:"--quick"`
	QuoteNames                 bool     `flag:"--quote-names"`
	Replace                    bool     `flag:"--replace"`
	Routines                   bool     `flag:"--routines"`
	SetCharset                 bool     `flag:"--set-charset"`
	SetGtidPurged              bool     `flag:"--set-gtid-purged"`
	SingleTransaction          bool     `flag:"--single-transaction"`
	SkipAddDropTable           bool     `flag:"--skip-add-drop-table"`
	SkipAddLocks               bool     `flag:"--skip-add-locks"`
	SkipComments               bool     `flag:"--skip-comments"`
	SkipCompact                bool     `flag:"--skip-compact"`
	SkipDisableKeys            bool     `flag:"--skip-disable-keys"`
	SkipExtendedInsert         bool     `flag:"--skip-extended-insert"`
	SkipLockTables             bool     `flag:"--skip-lock-tables"`
	SkipOpt                    bool     `flag:"--skip-opt"`
	SkipQuick                  bool     `flag:"--skip-quick"`
	SkipQuoteNames             bool     `flag:"--skip-quote-names"`
	SkipSetCharset             bool     `flag:"--skip-set-charset"`
	SkipTriggers               bool     `flag:"--skip-triggers"`
	SkipTzUtc                  bool     `flag:"--skip-tz-utc"`
	SslFipsMode                bool     `flag:"--ssl-fips-mode"`
	Triggers                   bool     `flag:"--triggers"`
	TzUtc                      bool     `flag:"--tz-utc"`
	XML                        bool     `flag:"--xml"`
}
