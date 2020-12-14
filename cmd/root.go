package cmd

import (
	"strings"

	"github.com/mittwald/brudi/pkg/config/mergeConfigs"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	cfgFiles        []string
	useRestic       bool
	useResticForget bool
	cleanup         bool

	rootCmd = &cobra.Command{
		Use:   "brudi",
		Short: "Easy backup creation",
		Long: `Easy, incremental and encrypted backup creation for different backends (file, mongoDB, mysql, etc.)
After creating your desired tar- or dump-file, brudi backs up the result with restic - if you want to`,
		Version: tag,
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolVar(&useRestic, "restic", false, "backup result with 'restic backup'")

	rootCmd.PersistentFlags().BoolVar(&useResticForget, "restic-forget", false, "executes 'restic forget' after backing up things with restic")

	rootCmd.PersistentFlags().BoolVar(&cleanup, "cleanup", false, "cleanup backup files afterwards")

	rootCmd.PersistentFlags().StringSliceVarP(&cfgFiles, "config", "c", []string{}, "config file (default is ${HOME}/.brudi.yaml)")
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func initConfig() {

	viper.SetConfigType("yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Merge configs into one
	cfgsRendered := mergeConfigs(configFiles)
	for _, conf := range cfgsRendered {
		if err := viper.MergeConfig(conf); err != nil {
			log.WithError(err).Fatalf("failed while reading config '%s'", conf)
		}
	}

	log.WithField("config", cfgFiles).Info("configs loaded")
}
