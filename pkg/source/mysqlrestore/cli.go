package mysqlrestore

const (
	binary = "mysql"
)

type Options struct {
	Flags          *Flags
	Command        string
	SourceFile     string
	AdditionalArgs []string
}

type Flags struct {
	Database              string `flag:"--database" valiudate:"min=1"`
	Debug                 string `flag:"--debug="`
	DebugCheck            bool   `flag:"--debug-check"`
	DebugInfo             bool   `flag:"--debug-info"`
	DefaultAuth           string `flag:"--default-auth="`
	DefaultCharacterSet   string `flag:"--default-character-set="`
	DefaultsExtraFile     string `flag:"--defaults-extra-file="`
	DefaultsFile          string `flag:"--defaults-file="`
	DefaultsGroupSuffix   string `flag:"--defaults-group-suffix="`
	EnableCleartextPlugin bool   `flag:"--enable-cleartext-plugin"`
	Force                 bool   `flaf:"--force"`
	GetServerPublicKey    bool   `flag:"--get-server-public-key"`
	Host                  string `flag:"--host=" validate:"min=1"`
	LoginPath             string `flag:"--login-path="`
	MaxAllowedPacket      string `flag:"--max-allowed-packet="`
	Password              string `flag:"--password="`
	Port                  int    `flag:"--port="`
	ServerPublicKeyPath   string `flag:"--server-public-key-path="`
	SharedMemoryBaseName  string `flag:"--shared-memory-base-name="`
	Socket                string `flag:"--socket="`
	SslCa                 string `flag:"--ssl-ca="`
	SslCaPath             string `flag:"--ssl-capath="`
	SslCert               string `flag:"--ssl-cert="`
	SslCipher             string `flag:"--ssl-cipher="`
	SslCrl                string `flag:"--ssl-crl="`
	SslCrlPath            string `flag:"--ssl-crlpath="`
	SslFipsMode           string `flag:"--ssl-fips-mode"`
	SslKey                string `flag:"--ssl-key="`
	TLSCipherSuites       string `flag:"--tls-ciphersuites="`
	TLSVersion            string `flag:"--tls-version="`
	User                  string `flag:"--user="`
	XML                   bool   `flag:"--xml"`
	ZstdCompressionLevel  int    `flag:"--zstd-compression-level"`

	//AutoRehash bool `flag:"--auto-rehash"`
	//AutoVerticalInput --auto-vertical-output
	//--batch
	//--binary-as-hex
	//--binary-mode
	//--bind-address
	//--character-sets-dir
	//--column-names
	//--column-type-info
	//--comments
	//--compress
	//--compression-algorithms
	//--connect-expired-password
	//--connect-timeout
	//--delimiter
	//--dns-srv-name
	//--execute
	//--help
	//--histignore
	//--html
	//--ignore-spaces
	//--init-command
	//--line-numbers
	//--load-data-local-dir
	//--local-infile
	//--max-join-size
	//--named-commands
	//--net-buffer-length
	//--network-namespace
	//--no-auto-rehash
	//--no-beep
	//--no-defaults
	//--one-database
	//--pager
	//--pipe
	//--plugin-dir
	//--print-defaults
	//--prompt
	//--protocol
	//--quick
	//--raw
	//--reconnect
	//--safe-updates, --i-am-a-dummy
	//--select-limit
	//--show-warnings
	//--sigint-ignore
	//--silent
	//--skip-auto-rehash
	//--skip-column-names
	//--skip-line-numbers
	//--skip-named-commands
	//--skip-pager
	//--skip-reconnect
	//--ssl-mode
	//--syslog
	//--table
	//--tee
	//--unbuffered
	//--verbose
	//--version
	//--vertical
	//--wait

}
