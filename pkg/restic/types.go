package restic

// Global options for restic
type GlobalOptions struct {
	Flags *GlobalFlags
}

// Global restic flags
type GlobalFlags struct {
	CaCert         string `flag:"--cacert"`
	CacheDir       string `flag:"--cache-dir"`
	KeyHint        string `flag:"--key-hint"`
	PasswordFile   string `flag:"--password-file"`
	Repo           string `flag:"--repo"`
	RepositoryFile string `flag:"--repository-file"`
	TLSClientCert  string `flag:"--tls-client-cert"`
	LimitDownload  int    `flag:"--limit-download"`
	LimitUpload    int    `flag:"--limit-upload"`
	CleanupCache   bool   `flag:"--cleanup-cache"`
	NoCache        bool   `flag:"--no-cache"`
	NoLock         bool   `flag:"--no-lock"`
}

// BackupResult for cmd "restic backup"
type BackupResult struct {
	SnapshotID       string
	ParentSnapshotID string
}

// BackupOptions for cmd: "restic backup"
type BackupOptions struct {
	Flags *BackupFlags
	Paths []string
}

// BackupFlags for cmd: "restic backup"
type BackupFlags struct {
	FilesFromVerbatim string   `flag:"--files-from-verbatim file"`
	ExcludeFile       string   `flag:"--exclude-file"`
	Host              string   `flag:"--host"`
	IexcludeFile      string   `flag:"--iexclude-file"`
	IexCludePattern   string   `flag:"--iexclude-pattern"`
	Parent            string   `flag:"--parent"`
	StdinFilename     string   `flag:"--stdin-filename"`
	Time              string   `flag:"--time"`
	Exclude           []string `flag:"-e"`
	FilesFromFile     []string `flag:"--files-from file"`
	FilesFromRaw      []string `flag:"--files-from-raw file"`
	Tags              []string `flag:"--tag"`
	ExcludeLargerThan int      `flag:"exclude-larger-than"`
	ExcludeCaches     bool     `flag:"--exclude-caches"`
	Force             bool     `flag:"-f"`
	IgnoreInode       bool     `flag:"--ignore-inode"`
	OneFileSystem     bool     `flag:"-x"`
	Stdin             bool     `flag:"--stdin"`
	WithAtime         bool     `flag:"--with-atime"`
}

// StatsOptions for cmd: "restic stats"
type StatsOptions struct {
	Flags *StatsFlags
	IDs   []string
}

// StatsFlags for cmd: "restic stats"
type StatsFlags struct {
	Host string `flag:"--host"`
	Mode string `flag:"--mode"`
}

// Stats for "restic stats" json-logging
//
//nolint:tagliatelle // upstream type
type Stats struct {
	TotalSize      uint64 `json:"total_size"`
	TotalFileCount uint64 `json:"total_file_count"`
	TotalBlobCount uint64 `json:"total_blob_count,omitempty"`
}

// SnapshotOptions for cmd: "restic snapshots"
type SnapshotOptions struct {
	Flags *SnapshotFlags
	IDs   []string
}

// Response wraps summary and status responses from "restic backup"
type Response struct {
	Responses []interface{}
}

// SnapshotFlags for cmd: "restic snapshots"
type SnapshotFlags struct {
	Host  string   `flag:"-H"`
	Paths []string `flag:"--path"`
	Tags  []string `flag:"--tag"`
}

// Snapshot type for the (json-)result of "restic snapshots"
type Snapshot struct {
	ID       *string  `json:"id"`
	Time     string   `json:"time"`
	Tree     string   `json:"tree"`
	Tags     []string `json:"tags"`
	Paths    []string `json:"paths"`
	Hostname string   `json:"hostname"`
	Username string   `json:"username"`
	UID      *int     `json:"uid"`
	GID      *int     `json:"gid"`
}

// CheckFlags for cmd: "restic check"
type CheckFlags struct {
	CheckUnused bool `flag:"--check-unused"`
	ReadData    bool `flag:"--read-data"`
}

// ForgetOptions for cmd: "restic forget"
type ForgetOptions struct {
	Flags *ForgetFlags
	IDs   []string
}

// ForgetFlags for cmd: "restic forget"
type ForgetFlags struct {
	KeepLast    int      `flag:"-l"`
	KeepHourly  int      `flag:"-H"`
	KeepDaily   int      `flag:"-d"`
	KeepWeekly  int      `flag:"-w"`
	KeepMonthly int      `flag:"-m"`
	KeepYearly  int      `flag:"-y"`
	KeepTags    []string `flag:"--keep-tag"`
	KeepWithin  string   `flag:"--keep-within"`
	Host        string   `flag:"--host"`
	Tags        []string `flag:"--tag"`
	Paths       []string `flag:"--path"`
	GroupBy     string   `flag:"-g"`
	DryRun      bool     `flag:"-n"`
	Prune       bool     `flag:"--prune"`
	Compact     bool     `flag:"--compact"`
}

// ForgetResponse for "restic forget" json-logging
type ForgetResponse struct {
	Tags []*ForgetTag
}

// ForgetTag for "restic forget" json-logging
type ForgetTag struct {
	Tags    []string       `json:"tags"`
	Host    string         `json:"host"`
	Paths   []string       `json:"paths"`
	Keep    []Snapshot     `json:"keep"`
	Remove  []Snapshot     `json:"remove"`
	Reasons []ForgetReason `json:"reasons"`
}

// ForgetReason for "restic forget" json-logging
type ForgetReason struct {
	Snapshot Snapshot `json:"snapshot"`
	Matches  []string `json:"matches"`
	Counters Counters `Json:"counters"`
}

// Counters for "restic forget" json-logging
type Counters struct {
	Last    int `json:"last"`
	Hourly  int `json:"hourly"`
	Daily   int `json:"daily"`
	Weekly  int `json:"weekly"`
	Monthly int `json:"monthly"`
	Yearly  int `json:"yearly"`
}

// ForgetSnapshot for "restic forget" json-logging
type ForgetSnapshot struct {
	Time     string   `json:"time"`
	Tree     string   `json:"tree"`
	Paths    []string `json:"paths"`
	Hostname string   `json:"hostname"`
	Username string   `json:"username"`
	UID      *int     `json:"uid"`
	GID      *int     `json:"gid"`
}

// RestoreOptions for cmd: "restic restore"
type RestoreOptions struct {
	Flags *RestoreFlags
	ID    string
}

// RestoreFlags for cmd: "restic restore"
type RestoreFlags struct {
	Exclude []string `flag:"-e"`
	Host    string   `flag:"-H"`
	Include []string `flag:"-i"`
	Verify  bool     `flag:"--verify"`
	Path    string   `flag:"--path"`
	Tags    string   `flag:"--tag"`
	Target  string   `flag:"-t"`
}

// DumpOptions for cmd: "restic dump"
type DumpOptions struct {
	ID   string
	File string
}

// TagOptions for cmd: "restic tag"
type TagOptions struct {
	Flags *TagFlags
	IDs   []string
}

// TagFlags for cmd: "restic tag"
type TagFlags struct {
	Add    []string `flag:"--add"`
	Host   string   `flag:"-H"`
	Path   string   `flag:"--path"`
	Remove []string `flag:"--remove"`
	Set    []string `flag:"--set"`
	Tag    string   `flag:"--tag"`
}

// FindOptions for cmd "restic find"
type FindOptions struct {
	Flags   *FindFlags
	Pattern string
}

// FindFlags for cmd "restic find"
type FindFlags struct {
	SnapshotID string `flag:"-s"`
	Host       string `flag:"-H"`
	IgnoreCase bool   `flag:"-i"`
	Long       bool   `flag:"-l"`
	Newest     string `flag:"-N"`
	Oldest     string `flag:"-O"`
	Path       string `flag:"--path"`
	Tag        string `flag:"--tag"`
}

// FindResult contains the result of "restic find"
type FindResult struct {
	Matches    []FindMatch `json:"matches"`
	Hits       int         `json:"hits"`
	SnapshotID *string     `json:"snapshot"`
}

// FindMatch represents one match of "restic find"
//
//nolint:tagliatelle // upstream type
type FindMatch struct {
	Path        string `json:"path"`
	Permissions string `json:"permissions"`
	Type        string `json:"type"`
	Mode        *int   `json:"mode"`
	MTime       string `json:"mtime"`
	ATime       string `json:"atime"`
	CTime       string `json:"ctime"`
	UID         *int   `json:"uid"`
	GID         *int   `json:"gid"`
	User        string `json:"user"`
	DeviceID    int    `json:"device_id"`
	Size        int    `json:"size"`
	Links       int    `json:"links"`
}

// LsOptions for cmd "restic ls"
type LsOptions struct {
	Flags       *LsFlags
	SnapshotIDs []string
}

// LsFlags for cmd "restic ls"
type LsFlags struct {
	Host string `flag:"-H"`
	Long bool   `flag:"-l"`
	Path string `flag:"--path"`
	Tag  string `flag:"--tag"`
}

// LsResult for cmd "restic ls"
type LsResult struct {
	SnapshotID string
	Paths      []string
	Time       string
	Files      []LsFile
	Size       uint64
}

// LsMessage for "resitc ls" json-logging
//
//nolint:tagliatelle // upstream restic type
type LsMessage struct {
	Time       string   `json:"time"`
	Tree       string   `json:"tree"`
	Paths      []string `json:"paths"`
	Hostname   string   `json:"hostname"`
	Username   string   `json:"username"`
	UID        *int     `json:"uid"`
	GID        *int     `json:"gid"`
	ID         *string  `json:"id"`
	ShortID    *string  `json:"short_id"`
	StructType string   `json:"struct_type"`
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Path       string   `json:"path"`
	Size       uint64   `json:"size"`
	Mode       *int     `json:"mode"`
	MTime      string   `json:"mtime"`
	CTime      string   `json:"ctime"`
	ATime      string   `json:"atime"`
}

// LsFile for cmd "restic ls"
type LsFile struct {
	Permissions int
	User        int
	Group       int
	Size        uint64
	Time        string
	Path        string
}

// UnlockOptions for cmd "restic unlock"
type UnlockOptions struct {
	Flags *UnlockFlags
}

// UnlockFlags for cmd "restic unlock"
type UnlockFlags struct {
	RemoveAll bool `flag:"--remove-all"`
}
