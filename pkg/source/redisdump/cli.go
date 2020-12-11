package redisdump

const (
	binary = "redis-cli"
)

type Options struct {
	Flags          *Flags
	AdditionalArgs []string
}

type Flags struct {
	Host             string `flag:"-h"`
	Port             int    `flag:"-p"`
	ResultFile       string `flag:"--rdb"`
	Password         string `flag:"-a"`
	Socket           string `flag:"-s"`
	URI              string `flag:"-u"`
	Repeat           int    `flag:"-r"`
	Interval         int    `flag:"-i"`
	DatabaseNumber   int    `flag:"-n"`
	ReadStdin        bool   `flag:"-x"`
	Delimiter        string `flag:"-d"`
	ClusterMode      bool   `flag:"-c"`
	Raw              bool   `flag:"--raw"`
	NoRaw            bool   `flag:"--no-raw"`
	Csv              bool   `flag:"--csv"`
	Stat             bool   `flag:"--stat"`
	Latency          bool   `flag:"--latency"`
	LatencyHistory   bool   `flag:"--latency-history"`
	LatencyDist      bool   `flag:"latency-dist"`
	LruTest          string `flag:"--lru-test="`
	Replica          bool   `flag:"--replica"`
	Pipe             bool   `flag:"--pipe"`
	PipeTimeout      int    `flag:"--pipe-timeout="`
	Bigkeys          bool   `flag:"--bigKeys"`
	Memkeys          bool   `flag:"--memkeys"`
	MemkeysSamples   int    `flag:"--memkeys-samples="`
	Hotkeys          bool   `flag:"--hotkeys"`
	Scan             bool   `flag:"--scan"`
	Pattern          string `flag:"--pattern="`
	IntrinsicLatency int    `flag:"--intrinsic-latency="`
	Eval             string `flag:"--eval="`
	Ldb              bool   `flag:"--ldb"`
	LdbSyncMode      bool   `flag:"--ldb-sync-mode"`
	Cluster          string `flag:"--cluster="`
	Verbose          bool   `flag:"--verbose"`
	NoAuthWarning    bool   `flag:"--no-auth-warning"`
}
