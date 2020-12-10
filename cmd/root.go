package cmd

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	cfgFile         string
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

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is ${HOME}/.brudi.yaml)")
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func initConfig() {
	if cfgFile == "" {
		home, err := homedir.Dir()
		if err != nil {
			log.WithError(err).Fatal("unable to determine homedir for current user")
		}

		cfgFile = path.Join(home, ".brudi.yaml")
	}

	logFields := log.WithField("cfgFile", cfgFile)

	info, err := os.Stat(cfgFile)
	if os.IsNotExist(err) {
		logFields.Warn("config does not exist")
		return
	} else if info.IsDir() {
		logFields.Warn("config is a directory")
		return
	}

	var cfgContent []byte
	cfgContent, err = ioutil.ReadFile(cfgFile)
	if err != nil {
		log.WithError(err).Fatal("failed while reading config")
	}

	var tpl *template.Template
	tpl, err = template.New("").Parse(string(cfgContent))
	if err != nil {
		log.WithError(err).Fatal()
	}

	type templateData struct {
		Env map[string]string
	}

	data := templateData{
		Env: make(map[string]string),
	}

	for _, e := range os.Environ() {
		e := strings.SplitN(e, "=", 2)
		if len(e) > 1 {
			data.Env[e[0]] = e[1]
		}
	}

	renderedCfg := new(bytes.Buffer)
	err = tpl.Execute(renderedCfg, &data)
	if err != nil {
		log.WithError(err).Fatal()
	}

	viper.SetConfigType("yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadConfig(renderedCfg); err != nil {
		log.WithError(err).Fatal("failed while reading config")
	}
	fmt.Println(viper.AllKeys())
	log.WithField("config", cfgFile).Info("config loaded")
}
