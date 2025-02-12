package mysqlrestore

const (
	binary = "mysql"
)

type Options struct {
	Flags          *Flags
	AdditionalArgs []string
	SourceFile     string
}

type Flags struct {
	BindAddress            string `flag:"--bind-address="`
	CharacterSetsDir       string `flag:"--character-sets-dir="`
	CompressionAlgorithms  string `flag:"--compression-algorithms="`
	Database               string `flag:"--database="`
	Debug                  string `flag:"--debug="`
	DefaultAuth            string `flag:"--default-auth="`
	DefaultCharacterSet    string `flag:"--default-character-set="`
	DefaultsExtraFile      string `flag:"--defaults-extra-file="`
	DefaultsFile           string `flag:"--defaults-file="`
	DefaultsGroupSuffix    string `flag:"--defaults-group-suffix="`
	Delimiter              string `flag:"--delimiter="`
	DNSSrvName             string `flag:"--dns-srv-name"`
	Execute                string `flag:"--execute="`
	Host                   string `flag:"--host="                    validate:"min=1"`
	InitCommand            string `flag:"--init-command="`
	LoadDataLocalDir       string `flag:"--load-data-local-dir="`
	LoginPath              string `flag:"--login-path="`
	MaxAllowedPacket       string `flag:"--max-allowed-packet="`
	NetworkNamespace       string `flag:"--network-namespace"`
	Pager                  string `flag:"--pager"`
	Password               string `flag:"--password="`
	PluginDir              string `flag:"--plugin-dir"`
	Promtp                 string `flag:"--prompt="`
	Protocol               string `flag:"--protocol="`
	ServerPublicKeyPath    string `flag:"--server-public-key-path="`
	SharedMemoryBaseName   string `flag:"--shared-memory-base-name="`
	Socket                 string `flag:"--socket="`
	Ssl                    *int   `flag:"--ssl="`
	SslCa                  string `flag:"--ssl-ca="`
	SslCaPath              string `flag:"--ssl-capath="`
	SslCert                string `flag:"--ssl-cert="`
	SslCipher              string `flag:"--ssl-cipher="`
	SslCrl                 string `flag:"--ssl-crl="`
	SslCrlPath             string `flag:"--ssl-crlpath="`
	SslFipsMode            string `flag:"--ssl-fips-mode"`
	SslKey                 string `flag:"--ssl-key="`
	Tee                    string `flag:"--tee="`
	TLSCipherSuites        string `flag:"--tls-ciphersuites="`
	TLSVersion             string `flag:"--tls-version="`
	User                   string `flag:"--user="`
	ConnectTimeout         int    `flag:"--connect-timeout"`
	LocalInfile            int    `flag:"--local-infile"`
	MaxJoinSize            int    `flag:"--max-join-size="`
	NetBufferLength        int    `flag:"--net-buffer-length="`
	Port                   int    `flag:"--port="`
	SelcetLimit            int    `flag:"--select-limit"`
	ZstdCompressionLevel   int    `flag:"--zstd-compression-level="`
	AutoRehash             bool   `flag:"--auto-rehash"`
	AutoVerticalInput      bool   `flag:"--auto-vertical-output"`
	Batch                  bool   `flag:"--batch"`
	BinaryAsHex            bool   `flag:"--binary-as-hex"`
	BinaryMode             bool   `flag:"--binary-mode"`
	ColumnNames            bool   `flag:"--column-names"`
	ColumnTypeInfo         bool   `flag:"column-type-info"`
	Comments               bool   `flag:"--comments"`
	Compress               bool   `flag:"--compress"`
	ConnectExpiredPassword bool   `flag:"--connect-expired-password"`
	DebugCheck             bool   `flag:"--debug-check"`
	DebugInfo              bool   `flag:"--debug-info"`
	DisableNamedCommands   bool   `flag:"--disable-named-commands"`
	EnableCleartextPlugin  bool   `flag:"--enable-cleartext-plugin"`
	Force                  bool   `flaf:"--force"`
	GetServerPublicKey     bool   `flag:"--get-server-public-key"`
	Help                   bool   `flag:"--help"`
	HistIgnore             bool   `flag:"--histignore"`
	HTML                   bool   `flag:"--html"`
	IAmADummy              bool   `flag:"--i-am-a-dummy"`
	IgnoreSpaces           bool   `flag:"--ignore-spaces"`
	LineNumbers            bool   `flag:"--line-numbers"`
	NamedCommands          bool   `flag:"--named-commands"`
	NoAutoRehash           bool   `flag:"--no-auto-rehash"`
	NoBeep                 bool   `flag:"--no-beep"`
	NoDefaults             bool   `flag:"--no-defaults"`
	OneDatabase            bool   `flag:"--one-database"`
	Pipe                   bool   `flag:"--pipe"`
	PrintDefaults          bool   `flag:"--print-defaults"`
	Quick                  bool   `flag:"--quick"`
	Raw                    bool   `flag:"--raw"`
	Reconnect              bool   `flag:"--reconnect"`
	SafeUpdates            bool   `flag:"--safe-updates"`
	ShowWarning            bool   `flag:"--show-warnings"`
	SigintIgnore           bool   `flag:"--sigint-ignore"`
	Silent                 bool   `flag:"--silent"`
	SkipAutoRehash         bool   `flag:"--skip-auto-rehash"`
	SkipColumnNames        bool   `flags:"--skip-column-names"`
	SkipLineNumbers        bool   `flags:"--skip-line-numbers"`
	SkipNamedCommands      bool   `flags:"--skip-named-commands"`
	SkipPager              bool   `flag:"--skip-pager"`
	SkipReconnect          bool   `flag:"--skip-reconnect"`
	Syslog                 bool   `flag:"--syslog"`
	Table                  bool   `flag:"--table"`
	Unbuffered             bool   `flag:"--unbuffered"`
	Verbose                bool   `flag:"--verbose"`
	Version                bool   `flag:"--version"`
	Vertical               bool   `flag:"--vertical"`
	Wait                   bool   `flag:"--wait"`
	XML                    bool   `flag:"--xml"`
}
