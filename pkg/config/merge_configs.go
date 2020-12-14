package config

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/spf13/viper"

	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
)

func MergeConfigs(cfgFiles []string) {
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

	exists := make(map[string]struct{})
	var cfgUniques []string
	for _, val := range cfgFiles {
		if _, ok := exists[val]; !ok {
			exists[val] = struct{}{}
			cfgUniques = append(cfgUniques, val)
		} else {
			logFields.Warnf("config '%s' has been specified more than once, ignoring additional instances", val)
		}
	}
	cfgFiles = cfgUniques

	for _, file := range cfgFiles {
		info, err := os.Stat(file)
		if os.IsNotExist(err) {
			logFields.Warnf("config '%s' does not exist", file)
			return
		} else if info.IsDir() {
			logFields.Warnf("config '%s' is a directory", file)
			return
		}
	}

	var cfgContent [][]byte
	for _, file := range cfgFiles {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			log.WithError(err).Fatalf("failed while reading config file '%s'", file)
		}
		cfgContent = append(cfgContent, content)
	}

	var tpl []*template.Template

	for _, content := range cfgContent {
		tpltemp, err := template.New("").Parse(string(content))
		if err != nil {
			log.WithError(err).Fatalf("failed while templating config '%s'", content)
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

	var cfgsRendered []*bytes.Buffer
	for _, template := range tpl {
		renderedCfg := new(bytes.Buffer)
		err := template.Execute(renderedCfg, &data)
		if err != nil {
			log.WithError(err).Fatalf("failed while rendering template '%s'", template.Name())
		}
		cfgsRendered = append(cfgsRendered, renderedCfg)
	}

	// Merge configs into one
	for _, conf := range cfgsRendered {
		if err := viper.MergeConfig(conf); err != nil {
			log.WithError(err).Fatalf("failed while reading config '%s'", conf)
		}
	}

	log.WithField("config", cfgFiles).Info("configs loaded")
}
