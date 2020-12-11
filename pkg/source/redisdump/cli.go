package redisdump

const (
	binary = "redis-cli"
)

type Options struct {
	Flags          *Flags
	AdditionalArgs []string
	Command        string
}

type Flags struct {
	Bigkeys          bool   `flag:"--bigKeys"`
	Cluster          string `flag:"--cluster="`
	ClusterMode      bool   `flag:"-c"`
	Csv              bool   `flag:"--csv"`
	DatabaseNumber   int    `flag:"-n"`
	Delimiter        string `flag:"-d"`
	Eval             string `flag:"--eval="`
	Host             string `flag:"-h"`
	Hotkeys          bool   `flag:"--hotkeys"`
	Interval         int    `flag:"-i"`
	IntrinsicLatency int    `flag:"--intrinsic-latency="`
	Latency          bool   `flag:"--latency"`
	LatencyDist      bool   `flag:"latency-dist"`
	LatencyHistory   bool   `flag:"--latency-history"`
	Ldb              bool   `flag:"--ldb"`
	LdbSyncMode      bool   `flag:"--ldb-sync-mode"`
	LruTest          string `flag:"--lru-test="`
	Memkeys          bool   `flag:"--memkeys"`
	MemkeysSamples   int    `flag:"--memkeys-samples="`
	NoAuthWarning    bool   `flag:"--no-auth-warning"`
	NoRaw            bool   `flag:"--no-raw"`
	Password         string `flag:"-a"`
	Pattern          string `flag:"--pattern="`
	Pipe             bool   `flag:"--pipe"`
	PipeTimeout      int    `flag:"--pipe-timeout="`
	Port             int    `flag:"-p"`
	Raw              bool   `flag:"--raw"`
	Rdb              string `flag:"--rdb"validate:"min=1"`
	ReadStdin        bool   `flag:"-x"`
	Repeat           int    `flag:"-r"`
	Replica          bool   `flag:"--replica"`
	Scan             bool   `flag:"--scan"`
	Socket           string `flag:"-s"`
	Stat             bool   `flag:"--stat"`
	URI              string `flag:"-u"`
	Verbose          bool   `flag:"--verbose"`
}
