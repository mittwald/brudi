package mysqlrestore

const (
	binary = "mysql"
)

type Options struct {
	Flags          *Flags
	SourceFile     string
	AdditionalArgs []string
}

type Flags struct {
	Host     string `flag:"--host=" validate:"min=1"`
	Password string `flag:"--password="`
	Port     int    `flag:"--port="`
	User     string `flag:"--user="`

	AutoRehash bool `flag:"--auto-rehash"`
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
	//--database
	//--debug
	//--debug-check
	//--debug-info
	//--default-auth
	//--default-character-set
	//--defaults-extra-file
	//--defaults-file
	//--defaults-group-suffix
	//--delimiter
	//--dns-srv-name
	//--enable-cleartext-plugin
	//--execute
	//--force
	//--get-server-public-key
	//--help
	//--histignore
	//--host
	//--html
	//--ignore-spaces
	//--init-command
	//--line-numbers
	//--load-data-local-dir
	//--local-infile
	//--login-path
	//--max-allowed-packet
	//--max-join-size
	//--named-commands
	//--net-buffer-length
	//--network-namespace
	//--no-auto-rehash
	//--no-beep
	//--no-defaults
	//--one-database
	//--pager
	//--password
	//--pipe
	//--plugin-dir
	//--port
	//--print-defaults
	//--prompt
	//--protocol
	//--quick
	//--raw
	//--reconnect
	//--safe-updates, --i-am-a-dummy
	//--select-limit
	//--server-public-key-path
	//--shared-memory-base-name
	//--show-warnings
	//--sigint-ignore
	//--silent
	//--skip-auto-rehash
	//--skip-column-names
	//--skip-line-numbers
	//--skip-named-commands
	//--skip-pager
	//--skip-reconnect
	//--socket
	//--ssl-ca
	//--ssl-capath
	//--ssl-cert
	//--ssl-cipher
	//--ssl-crl
	//--ssl-crlpath
	//--ssl-fips-mode
	//--ssl-key
	//--ssl-mode
	//--syslog
	//--table
	//--tee
	//--tls-ciphersuites
	//--tls-version
	//--unbuffered
	//--user
	//--verbose
	//--version
	//--vertical
	//--wait
	//--xml
	//--zstd-compression-level

}
