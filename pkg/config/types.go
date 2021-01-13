package config

const (
	KeyOptionsFlags          = "options.flags"
	KeyOptionsAdditionalArgs = "options.additionalArgs"
)

type ExtraResticFlags struct {
	ResticList    bool
	ResticCheck   bool
	ResticPrune   bool
	ResticRebuild bool
	ResticTags    bool
}
