package cmd

import (
	"os"
	"path"
	"strings"

	"github.com/mittwald/brudi/pkg/config"

	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	cfgFiles            []string
	useRestic           bool
	useResticForget     bool
	cleanup             bool
	listResticSnapshots bool

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

	rootCmd.PersistentFlags().BoolVar(&listResticSnapshots, "restic-snapshots", false, "List snapshots in restic repository afterwards")

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

	home, err := homedir.Dir()
	if err != nil {
		log.WithError(err).Fatal("unable to determine homedir for current user")
	}

	logFields := log.WithField("cfgFiles", cfgFiles)
	// check if default config exists and prepend it to list of configs
	_, err = os.Stat(path.Join(home, ".brudi.yaml"))
	if os.IsNotExist(err) {
		logFields.Warn("default config does not exist")
	} else {
		cfgFiles = append([]string{path.Join(home, ".brudi.yaml")}, cfgFiles...)
	}

	configFiles := config.ReadPaths(cfgFiles...)
	templatedConfigs := config.RawConfigs(configFiles)
	renderedConfigs := config.RenderConfigs(templatedConfigs)
	config.MergeConfigs(renderedConfigs)

	log.WithField("config", cfgFiles).Info("configs loaded")
}
