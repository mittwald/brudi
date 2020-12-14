package config

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// ReadPaths checks if config files exist, creates a list of unique configs and read files from disk
func ReadPaths(cfgFiles ...string) [][]byte {
	logFields := log.WithField("cfgFiles", cfgFiles)

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
			logFields.Fatalf("config '%s' does not exist", file)
		} else if info.IsDir() {
			logFields.Fatalf("config '%s' is a directory", file)
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
	return cfgContent
}

// RawConfigs creates templates for provided configs
func RawConfigs(configContent [][]byte) []*template.Template {
	var tpl []*template.Template

	for _, content := range configContent {
		tpltemp, err := template.New("").Parse(string(content))
		if err != nil {
			log.WithError(err).Fatalf("failed while templating config '%s'", content)
		}
		tpl = append(tpl, tpltemp)
	}
	return tpl
}

// RenderConfigs fills templated configs with environment variables
func RenderConfigs(templates []*template.Template) []*bytes.Buffer {
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
	for _, template := range templates {
		renderedCfg := new(bytes.Buffer)
		err := template.Execute(renderedCfg, &data)
		if err != nil {
			log.WithError(err).Fatalf("failed while rendering template '%s'", template.Name())
		}
		cfgsRendered = append(cfgsRendered, renderedCfg)
	}
	return cfgsRendered
}

// MergeConfigs merges configs into viper
func MergeConfigs(renderedConfigs []*bytes.Buffer) {
	for _, conf := range renderedConfigs {
		if err := viper.MergeConfig(conf); err != nil {
			log.WithError(err).Fatalf("failed while reading config '%s'", conf)
		}
	}
}
