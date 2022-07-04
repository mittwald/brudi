package xfsrestore

const (
	binary = "xfsrestore"
)

type Options struct {
	Flags          *Flags
	AdditionalArgs []string
	DestFS         string // filesystem to be dumped
}

type Flags struct {
	Housekeeping                  bool     `flag:"-a"`
	PreventOverride               bool     `flag:"-e"`
	InteractiveOperation          bool     `flag:"-i"`
	UseMinimalTapeProtocol        bool     `flag:"-m"`
	SourceIsQIC                   bool     `flag:"-q"`
	CumulativeMode                bool     `flag:"-r"`
	DisplayContents               bool     `flag:"-t"`
	DontRestoreExtendedAttributes bool     `flag:"-A"`
	MatchOwnershipToDumpRoot      bool     `flag:"-B"`
	RestoreDMAPI                  bool     `flag:"-D"`
	DontOverwriteNever            bool     `flag:"-E"`
	InhibitInteractivePrompts     bool     `flag:"-F"`
	ShowInventory                 bool     `flag:"-I"`
	InhibitInventoryUpdate        bool     `flag:"-J"`
	ForceCompletion               bool     `flag:"-Q"`
	ResumeInterruptedSession      bool     `flag:"-R"`
	InhibitDialogueTimeouts       bool     `flag:"-T"`
	BlockSizeInBytes              int      `flag:"-b"`
	ProgressReportInterval        int      `flag:"-p"`
	BufferRingLength              int      `flag:"-Y"`
	AlertProgramName              string   `flag:"-c"`
	Source                        string   `flag:"-f"`
	RestoreOnlyNeverTHan          string   `flag:"-n"`
	Subtree                       string   `flag:"-s"`
	Verbosity                     []string `flag:"-v"`
	SessionLabel                  string   `flag:"-L"`
	OptionsFile                   string   `flag:"-O"`
	SessionUUID                   string   `flag:"-S"`
	Exclude                       string   `flag:"-X"`
}
