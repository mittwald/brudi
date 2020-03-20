package restic

// Global options for restic
type GlobalOptions struct {
	Flags *GlobalFlags
}

// Global restic flags
type GlobalFlags struct {
	CaCert        string `flag:"--cacert"`
	CacheDir      string `flag:"--cache-dir"`
	CleanupCache  bool   `flag:"--cleanup-cache"`
	KeyHint       string `flag:"--key-hint"`
	LimitDownload int    `flag:"--limit-download"`
	LimitUpload   int    `flag:"--limit-upload"`
	NoCache       bool   `flag:"--no-cache"`
	NoLock        bool   `flag:"--no-lock"`
	PasswordFile  string `flag:"--password-file"`
	Repo          string `flag:"--repo"`
	TLSClientCert string `flag:"--tls-client-cert"`
	//--json                     set output mode to JSON for commands that support it
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
	Exclude       []string `flag:"-e"`
	ExcludeCaches bool     `flag:"--exclude-caches"`
	Force         bool     `flag:"-f"`
	Host          string   `flag:"--host"`
	OneFileSystem bool     `flag:"-x"`
	Parent        string   `flag:"--parent"`
	Tags          []string `flag:"--tag"`
	Time          string   `flag:"--time"`
	Stdin         bool     `flag:"--stdin"`
	StdinFilename string   `flag:"--stdin-filename"`
}

// StatsOptions for cmd: "restic stats"
type StatsOptions struct {
	Flags *StatsFlags
	IDs   []string
}

type StatsFlags struct {
	Host string `flag:"--host"`
	Mode string `flag:"--mode"`
}

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

// SnapshotFlags for cmd: "restic snapshots"
type SnapshotFlags struct {
	Host  string   `flag:"-H"`
	Paths []string `flag:"--path"`
	Tags  []string `flag:"--tag"`
}

// Snapshot type for the (json-)result of "restic snapshots"
type Snapshot struct {
	ID       string   `json:"id"`
	Time     string   `json:"time"`
	Tree     string   `json:"tree"`
	Tags     []string `json:"tags,omitempty"`
	Paths    []string `json:"paths"`
	Hostname string   `json:"hostname"`
	Username string   `json:"username"`
	UID      int      `json:"uid"`
	GID      int      `json:"gid"`
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
	SnapshotID string      `json:"snapshot"`
}

// FindMatch represents one match of "restic find"
type FindMatch struct {
	Path        string `json:"path"`
	Permissions string `json:"permissions"`
	Type        string `json:"type"`
	Mode        int    `json:"mode"`
	MTime       string `json:"mtime"`
	ATime       string `json:"atime"`
	CTime       string `json:"ctime"`
	UID         int    `json:"uid"`
	GID         int    `json:"gid"`
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

// LsFile for cmd "restic ls"
type LsFile struct {
	Permissions string `json:"permissions"`
	User        string `json:"user"`
	Group       string `json:"group"`
	Size        uint64 `json:"size"`
	Time        string `json:"time"`
	Path        string `json:"path"`
}

// UnlockOptions for cmd "restic unlock"
type UnlockOptions struct {
	Flags *UnlockFlags
}

// UnlockFlags for cmd "restic unlock"
type UnlockFlags struct {
	RemoveAll bool `flag:"--remove-all"`
}
