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

type Response struct {
	ResponseSummary []SummaryResponse
	ResponseStatus  []StatusResponse
}

type SummaryResponse struct {
	MessageType         string  `json:"message_type"`
	FilesNew            int     `json:"flies_new"`
	FilesChanged        int     `json:"files_changed"`
	FilesUnmodified     int     `json:"files_unmodified"`
	DirsNew             int     `json:"dirs_new"`
	DirsChanged         int     `json:"dirs_changed"`
	DirsUnmodified      int     `json:"dirs_unmodified"`
	DataBlobs           int     `json:"data_blobs"`
	TreeBlobs           int     `json:"tree_blobs"`
	DataAdded           int     `json:"data_added"`
	TotalFilesProcessed int     `json:"total_files_processed"`
	TotalBytesProcessed int     `json:"total_bytes_processed"`
	TotalDuration       float32 `json:"total_duration"`
	SnapshotID          string  `json:"snapshot_id"`
	ParentSnapshotID    string  `json:"parent"`
}

type StatusResponse struct {
	MessageType  string   `json:"message_type"`
	PercentDone  int      `json:"percent_done"`
	TotalFiles   int      `json:"total_files"`
	FilesDone    int      `json:"files:done"`
	TotalBytes   int      `json:"total_bytes"`
	BytesDone    int      `json:"bytes_done"`
	CurrentFiles []string `json:"current_files"`
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
	Tags     []string `json:"tags"`
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

type ForgetResponse struct {
	tags []ForgetTag
}

type ForgetTag struct {
	Tags   []string       `json:"tags"`
	Host   string         `json:"host"`
	Paths  []string       `json:"paths"`
	Keep   []Snapshot     `json:"keep"`
	Remove []Snapshot     `json:"remove"`
	Resons []ForgetReason `json:"reasons"`
}

type ForgetReason struct {
	Snapshot Snapshot `json:"snapshot"`
	Matches  []string `json:"matches"`
	Counters Counters `Json:"counters"`
}

type Counters struct {
	Last    int `json:"last"`
	Hourly  int `json:"hourly"`
	Daily   int `json:"daily"`
	Weekly  int `json:"weekly"`
	Monthly int `json:"monthly"`
	Yearly  int `json:"yearly"`
}

type ForgetSnapshot struct {
	Time     string   `json:"time"`
	Tree     string   `json:"tree"`
	Paths    []string `json:"paths"`
	Hostname string   `json:"hostname"`
	Username string   `json:"username"`
	UID      int      `json:"uid"`
	GID      int      `json:"gid"`
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

type LsMessage struct {
	Time      string   `json:"time"`
	Tree      string   `json:"tree"`
	Paths     []string `json:"paths"`
	Hostname  string   `json:"hostname"`
	Username  string   `json:"username"`
	UID       int      `json:"uid"`
	GID       int      `json:"gid"`
	ID        string   `json:"id"`
	ShortID   string   `json:"short-id"`
	StrucType string   `json:"struct-type"`
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	Path      string   `json:"path"`
	Size      uint64   `json:"size"`
	Mode      int      `json:"mode"`
	MTime     string   `json:"mtime"`
	CTime     string   `json:"ctime"`
	ATime     string   `json:"atime"`
}

// LsFile for cmd "restic ls"
type LsFile struct {
	Permissions int    `json:"permissions"`
	User        int    `json:"user"`
	Group       int    `json:"group"`
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
