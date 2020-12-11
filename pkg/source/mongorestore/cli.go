package mongorestore

const (
	binary = "mongorestore"
)

type Options struct {
	Flags          *Flags
	AdditionalArgs []string
}

type Flags struct {
	URI                              string `flag:"--uri="`
	Host                             string `flag:"--host=" validate:"min=1"`
	Port                             int    `flag:"--port="`
	Ssl                              bool   `flag:"--ssl"`
	SslCAFile                        string `flag:"--sslCAFile="`
	SslPEMKeyFile                    string `flag:"--sslPEMKeyFile="`
	SslPEMKeyPassword                string `flag:"--sslPEMKeyPassword="`
	SslCRLFile                       string `flag:"--sslCRLFile="`
	SslFIPSMode                      bool   `flag:"	--sslFIPSMode"`
	TlSInsecure                      bool   `flag:"--tlsInsecure"`
	Username                         string `flag:"--username="`
	Password                         string `flag:"--password="`
	AuthenticationDatabase           string `flag:"--authenticationDatabase="`
	AuthenticationMechanism          string `flag:"--authenticationMechanism="`
	GssapiServiceName                string `flag:"--gssapiServiceName="`
	GssapiHostName                   string `flag:"--gssapiHostName="`
	Database                         string `flag:"--db="`
	Collection                       string `flag:"--collection="`
	NsExclude                        string `flag:"--nsExclude="`
	NsInclude                        string `flag:"--nsInclude="`
	NsFrom                           string `flag:"--nsFrom="`
	NsTo                             string `flag:"--nsTo="`
	ReadPreference                   string `flag:"--readPreference="`
	ForceTableScan                   bool   `flag:"--forceTableScan"`
	Objcheck                         bool   `flag:"--objcheck"`
	OplogReplay                      bool   `flag:"--oplogReplay"`
	OplogLimit                       int    `flag:"--oplogLimit="`
	OplogFile                        string `flag:"--oplogFile="`
	RestoreOnUsersAndRoles           bool   `flag:"--restoreDbUsersAndRoles"`
	Dir                              string `flag:"--dir="`
	Drop                             bool   `flag:"--drop"`
	DryRun                           bool   `flag:"--dryRun"`
	WriteConcern                     string `flag:"--writeConcern="`
	NoIndexRestore                   bool   `flag:"--noIndexRestore"`
	ConvertLegacyIndexes             bool   `flag:"--convertLegacyIndexes"`
	NoOptionRestore                  bool   `flag:"--noOptionsRestore"`
	KeepIndexVersion                 bool   `flag:"--keepIndexVersion"`
	MaintainInsertionOrder           bool   `flag:"--maintainInsertionOrder"`
	NumInsertionWorkersPerCollection int    `flag:"--numInsertionWorkersPerCollection="`
	StopOnError                      bool   `flag:"--stopOnError"`
	BypassDocumentValidation         bool   `flag:"--bypassDocumentValidation"`
	PreserveUUID                     bool   `flag:"--preserveUUID"`
	FixDottedHashIndex               bool   `flag:"--fixDottedHashIndex"`
	Gzip                             bool   `flag:"--gzip"`
	Out                              string `flag:"--out="`
	Archive                          string `flag:"--archive="`
	ExcludeCollection                string `flag:"--excludeCollection="`
	ExcludeCollectionsWithPrefix     string `flag:"--excludeCollectionsWithPrefix="`
	NumParallelCollections           int    `flag:"--numParallelCollections="`
}
