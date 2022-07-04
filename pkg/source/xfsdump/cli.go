package xfsdump

const (
	binary = "xfsdump"
)

type Options struct {
	Flags          *Flags
	AdditionalArgs []string
	TargetFS       string // filesystem to be dumped
}

type Flags struct {
	IgnoreFilesWithOfflineCopies bool     `flag:"-a"`
	Exclude                      bool     `flag:"-e"`
	UseMinimalTapeProtocol       bool     `flag:"-m"`
	Overwrite                    bool     `flag:"-o"`
	DestinationIsQIC             bool     `flag:"-q"`
	DontDumpExtendedAttributes   bool     `flag:"-A"`
	PreEraseMedia                bool     `flag:"-E"`
	DontPromptOperator           bool     `flag:"-F"`
	ShowInventory                bool     `flag:"-I"`
	InhibitInventoryUpdate       bool     `flag:"-J"`
	ResumeInterruptedSession     bool     `flag:"-R"`
	InhibitDialogueTimeouts      bool     `flag:"-T"`
	BlockSizeInBytes             int      `flag:"-b"`
	FileSize                     int      `flag:"-d"`
	Level                        int      `flag:"-l"`
	ProgressReportInterval       int      `flag:"-p"`
	MaxIncludedFileSize          int      `flag:"-z"`
	BufferRingLength             int      `flag:"-Y"`
	AlertProgramName             string   `flag:"-c"`
	Destination                  string   `flag:"-f"`
	OnlyFromPath                 string   `flag:"-s"`
	DumpTimeFromFIle             string   `flag:"-t"`
	BaseOnSessionID              string   `flag:"-B"`
	SessionLabel                 string   `flag:"-L"`
	MediaObjectLabel             string   `flag:"-M"`
	OptionsFile                  string   `flag:"-O"`
	Verbosity                    []string `flag:"-v"`
}
