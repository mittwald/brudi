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
	SslCAFile                        string `flag:"--sslCAFile="`
	SslPEMKeyFile                    string `flag:"--sslPEMKeyFile="`
	SslPEMKeyPassword                string `flag:"--sslPEMKeyPassword="`
	SslCRLFile                       string `flag:"--sslCRLFile="`
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
	OplogFile                        string `flag:"--oplogFile="`
	Dir                              string `flag:"--dir="`
	WriteConcern                     string `flag:"--writeConcern="`
	Out                              string `flag:"--out="`
	Archive                          string `flag:"--archive="`
	ExcludeCollection                string `flag:"--excludeCollection="`
	ExcludeCollectionsWithPrefix     string `flag:"--excludeCollectionsWithPrefix="`
	Port                             int    `flag:"--port="`
	OplogLimit                       int    `flag:"--oplogLimit="`
	NumInsertionWorkersPerCollection int    `flag:"--numInsertionWorkersPerCollection="`
	NumParallelCollections           int    `flag:"--numParallelCollections="`
	Ssl                              bool   `flag:"--ssl"`
	SslFIPSMode                      bool   `flag:"	--sslFIPSMode"`
	TlSInsecure                      bool   `flag:"--tlsInsecure"`
	ForceTableScan                   bool   `flag:"--forceTableScan"`
	Objcheck                         bool   `flag:"--objcheck"`
	OplogReplay                      bool   `flag:"--oplogReplay"`
	RestoreOnUsersAndRoles           bool   `flag:"--restoreDbUsersAndRoles"`
	Drop                             bool   `flag:"--drop"`
	DryRun                           bool   `flag:"--dryRun"`
	NoIndexRestore                   bool   `flag:"--noIndexRestore"`
	ConvertLegacyIndexes             bool   `flag:"--convertLegacyIndexes"`
	NoOptionRestore                  bool   `flag:"--noOptionsRestore"`
	KeepIndexVersion                 bool   `flag:"--keepIndexVersion"`
	MaintainInsertionOrder           bool   `flag:"--maintainInsertionOrder"`
	StopOnError                      bool   `flag:"--stopOnError"`
	BypassDocumentValidation         bool   `flag:"--bypassDocumentValidation"`
	PreserveUUID                     bool   `flag:"--preserveUUID"`
	FixDottedHashIndex               bool   `flag:"--fixDottedHashIndex"`
	Gzip                             bool   `flag:"--gzip"`
}
