package cmd

import (
	"bytes"
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
	cfgFile         []string
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

	rootCmd.PersistentFlags().StringSliceVarP(&cfgFile, "config", "c", []string{}, "config file (default is ${HOME}/.brudi.yaml)")
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func initConfig() {
	if len(cfgFile) == 0 {
		home, err := homedir.Dir()
		if err != nil {
			log.WithError(err).Fatal("unable to determine homedir for current user")
		}
		cfgFile[0] = path.Join(home, ".brudi.yaml")
	}

	logFields := log.WithField("cfgFiles", cfgFile)

	for _, file := range cfgFile {
		info, err := os.Stat(file)
		if os.IsNotExist(err) {
			logFields.Warn("config does not exist")
			return
		} else if info.IsDir() {
			logFields.Warn("config is a directory")
			return
		}
	}

	var cfgContent [][]byte
	for _, file := range cfgFile {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			log.WithError(err).Fatalf("failed while reading config file %s", file)
			continue
		}
		cfgContent = append(cfgContent, content)
	}

	var tpl []*template.Template

	for _, content := range cfgContent {
		tpltemp, err := template.New("").Parse(string(content))
		if err != nil {
			log.WithError(err).Fatal()
			continue
		}
		tpl = append(tpl, tpltemp)
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

	var renderedCFGs []*bytes.Buffer
	for _, template := range tpl {
		renderedCfg := new(bytes.Buffer)
		err := template.Execute(renderedCfg, &data)
		if err != nil {
			log.WithError(err).Fatal()
			continue
		}
		renderedCFGs = append(renderedCFGs, renderedCfg)
	}

	viper.SetConfigType("yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Merge configs into one
	for _, conf := range renderedCFGs {
		if err := viper.MergeConfig(conf); err != nil {
			log.WithError(err).Fatal("failed while reading config %s", conf)
		}
	}

	log.WithField("config", cfgFile).Info("config loaded")
}
