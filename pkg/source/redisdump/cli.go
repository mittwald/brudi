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
	Cluster          string `flag:"--cluster="`
	Delimiter        string `flag:"-d"`
	Eval             string `flag:"--eval="`
	Host             string `flag:"-h" validate:"min=1"`
	LruTest          string `flag:"--lru-test="`
	Password         string `flag:"-a"`
	Pattern          string `flag:"--pattern="`
	Rdb              string `flag:"--rdb" validate:"min=1"`
	Socket           string `flag:"-s"`
	URI              string `flag:"-u"`
	DatabaseNumber   int    `flag:"-n"`
	Interval         int    `flag:"-i"`
	IntrinsicLatency int    `flag:"--intrinsic-latency="`
	MemkeysSamples   int    `flag:"--memkeys-samples="`
	PipeTimeout      int    `flag:"--pipe-timeout="`
	Port             int    `flag:"-p"`
	Repeat           int    `flag:"-r"`
	Bigkeys          bool   `flag:"--bigKeys"`
	ClusterMode      bool   `flag:"-c"`
	Csv              bool   `flag:"--csv"`
	Hotkeys          bool   `flag:"--hotkeys"`
	Latency          bool   `flag:"--latency"`
	LatencyDist      bool   `flag:"latency-dist"`
	LatencyHistory   bool   `flag:"--latency-history"`
	Ldb              bool   `flag:"--ldb"`
	LdbSyncMode      bool   `flag:"--ldb-sync-mode"`
	Memkeys          bool   `flag:"--memkeys"`
	NoAuthWarning    bool   `flag:"--no-auth-warning"`
	NoRaw            bool   `flag:"--no-raw"`
	Pipe             bool   `flag:"--pipe"`
	Raw              bool   `flag:"--raw"`
	ReadStdin        bool   `flag:"-x"`
	Replica          bool   `flag:"--replica"`
	Scan             bool   `flag:"--scan"`
	Stat             bool   `flag:"--stat"`
	Verbose          bool   `flag:"--verbose"`
}
