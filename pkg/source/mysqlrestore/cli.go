package mysqlrestore

const (
	binary = "mysql"
)

type Options struct {
	Flags          *Flags
	AdditionalArgs []string
	Command        string
	SourceFile     string
}

type Flags struct {
	AutoRehash             bool   `flag:"--auto-rehash"`
	AutoVerticalInput      bool   `flag:"--auto-vertical-output"`
	Batch                  bool   `flag:"--batch"`
	BinaryAsHex            bool   `flag:"--binary-as-hex"`
	BinaryMode             bool   `flag:"--binary-mode"`
	BindAddress            string `flag:"--bind-address="`
	CharacterSetsDir       string `flag:"--character-sets-dir="`
	ColumnNames            bool   `flag:"--column-names"`
	ColumnTypeInfo         bool   `flag:"column-type-info"`
	Comments               bool   `flag:"--comments"`
	Compress               bool   `flag:"--compress"`
	CompressionAlgorithms  string `flag:"--compression-algorithms="`
	ConnectExpiredPassword bool   `flag:"--connect-expired-password"`
	ConnectTimeout         int    `flag:"--connect-timeout"`
	Database               string `flag:"--database=" validate:"min=1"`
	Debug                  string `flag:"--debug="`
	DebugCheck             bool   `flag:"--debug-check"`
	DebugInfo              bool   `flag:"--debug-info"`
	DefaultAuth            string `flag:"--default-auth="`
	DefaultCharacterSet    string `flag:"--default-character-set="`
	DefaultsExtraFile      string `flag:"--defaults-extra-file="`
	DefaultsFile           string `flag:"--defaults-file="`
	DefaultsGroupSuffix    string `flag:"--defaults-group-suffix="`
	Delimiter              string `flag:"--delimiter="`
	DisableNamedCommands   bool   `flag:"--disable-named-commands"`
	DNSSrvName             string `flag:"--dns-srv-name"`
	EnableCleartextPlugin  bool   `flag:"--enable-cleartext-plugin"`
	Execute                bool   `flag:"--execute"`
	Force                  bool   `flaf:"--force"`
	GetServerPublicKey     bool   `flag:"--get-server-public-key"`
	Help                   bool   `flag:"--help"`
	HistIgnore             bool   `flag:"--histignore"`
	Host                   string `flag:"--host=" validate:"min=1"`
	HTML                   bool   `flag:"--html"`
	IAmADummy              bool   `flag:"--i-am-a-dummy"`
	IgnoreSpaces           bool   `flag:"--ignore-spaces"`
	InitCommand            string `flag:"--init-command="`
	LineNumbers            bool   `flag:"--line-numbers"`
	LoadDataLocalDir       string `flag:"--load-data-local-dir="`
	LocalInfile            int    `flag:"--local-infile"`
	LoginPath              string `flag:"--login-path="`
	MaxAllowedPacket       string `flag:"--max-allowed-packet="`
	MaxJoinSize            int    `flag:"--max-join-size="`
	NamedCommands          bool   `flag:"--named-commands"`
	NetBufferLength        int    `flag:"--net-buffer-length="`
	NetworkNamespace       string `flag:"--network-namespace"`
	NoAutoRehash           bool   `flag:"--no-auto-rehash"`
	NoBeep                 bool   `flag:"--no-beep"`
	NoDefaults             bool   `flag:"--no-defaults"`
	OneDatabase            bool   `flag:"--one-database"`
	Pager                  string `flag:"--pager"`
	Password               string `flag:"--password="`
	Pipe                   bool   `flag:"--pipe"`
	PluginDir              string `flag:"--plugin-dir"`
	Port                   int    `flag:"--port="`
	PrintDefaults          bool   `flag:"--print-defaults"`
	Promtp                 string `flag:"--prompt="`
	Protocol               string `flag:"--protocol="`
	Quick                  bool   `flag:"--quick"`
	Raw                    bool   `flag:"--raw"`
	Reconnect              bool   `flag:"--reconnect"`
	SafeUpdates            bool   `flag:"--safe-updates"`
	SelcetLimit            int    `flag:"--select-limit"`
	ServerPublicKeyPath    string `flag:"--server-public-key-path="`
	SharedMemoryBaseName   string `flag:"--shared-memory-base-name="`
	ShowWarning            bool   `flag:"--show-warnings"`
	SigintIgnore           bool   `flag:"--sigint-ignore"`
	Silent                 bool   `flag:"--silent"`
	SkipAutoRehash         bool   `flag:"--skip-auto-rehash"`
	SkipColumnNames        bool   `flags:"--skip-column-names"`
	SkipLineNumbers        bool   `flags:"--skip-line-numbers"`
	SkipNamedCommands      bool   `flags:"--skip-named-commands"`
	SkipPager              bool   `flag:"--skip-pager"`
	SkipReconnect          bool   `flag:"--skip-reconnect"`
	Socket                 string `flag:"--socket="`
	SslCa                  string `flag:"--ssl-ca="`
	SslCaPath              string `flag:"--ssl-capath="`
	SslCert                string `flag:"--ssl-cert="`
	SslCipher              string `flag:"--ssl-cipher="`
	SslCrl                 string `flag:"--ssl-crl="`
	SslCrlPath             string `flag:"--ssl-crlpath="`
	SslFipsMode            string `flag:"--ssl-fips-mode"`
	SslKey                 string `flag:"--ssl-key="`
	Syslog                 bool   `flag:"--syslog"`
	Table                  bool   `flag:"--table"`
	Tee                    string `flag:"--tee="`
	TLSCipherSuites        string `flag:"--tls-ciphersuites="`
	TLSVersion             string `flag:"--tls-version="`
	Unbuffered             bool   `flag:"--unbuffered"`
	User                   string `flag:"--user="`
	Verbose                bool   `flag:"--verbose"`
	Version                bool   `flag:"--version"`
	Vertical               bool   `flag:"--vertical"`
	Wait                   bool   `flag:"--wait"`
	XML                    bool   `flag:"--xml"`
	ZstdCompressionLevel   int    `flag:"--zstd-compression-level="`
}
