package mongodump

const (
	binary = "mongodump"
)

type Options struct {
	Flags          *Flags
	AdditionalArgs []string
}

type Flags struct {
	URI                          string `flag:"--uri="`
	Host                         string `flag:"--host=" validate:"min=1"`
	SslCAFile                    string `flag:"--sslCAFile="`
	SslPEMKeyFile                string `flag:"--sslPEMKeyFile="`
	SslPEMKeyPassword            string `flag:"--sslPEMKeyPassword="`
	SslCRLFile                   string `flag:"--sslCRLFile="`
	Username                     string `flag:"--username="`
	Password                     string `flag:"--password="`
	AuthenticationDatabase       string `flag:"--authenticationDatabase="`
	AuthenticationMechanism      string `flag:"--authenticationMechanism="`
	GssapiServiceName            string `flag:"--gssapiServiceName="`
	GssapiHostName               string `flag:"--gssapiHostName="`
	Database                     string `flag:"--db="`
	Collection                   string `flag:"--collection="`
	Query                        string `flag:"--query="`
	QueryFile                    string `flag:"--queryFile="`
	ReadPreference               string `flag:"--readPreference="`
	Out                          string `flag:"--out="`
	Archive                      string `flag:"--archive="`
	ExcludeCollection            string `flag:"--excludeCollection="`
	ExcludeCollectionsWithPrefix string `flag:"--excludeCollectionsWithPrefix="`
	Port                         int    `flag:"--port="`
	NumParallelCollections       int    `flag:"--numParallelCollections="`
	IPv6                         bool   `flag:"--ipv6"`
	Ssl                          bool   `flag:"--ssl"`
	SslAllowInvalidCertificates  bool   `flag:"--sslAllowInvalidCertificates"`
	SslAllowInvalidHostnames     bool   `flag:"--sslAllowInvalidHostnames"`
	ForceTableScan               bool   `flag:"--forceTableScan"`
	Gzip                         bool   `flag:"--gzip"`
	Oplog                        bool   `flag:"--oplog"`
	DumpDBUsersAndRoles          bool   `flag:"--dumpDbUsersAndRoles"`
	ViewsAsCollections           bool   `flag:"--viewsAsCollections"`
}
