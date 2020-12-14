package config

import (
	"bytes"
	"github.com/mitchellh/go-homedir"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

func ReadConfigFiles(cfgFiles []string) [][]byte {
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

func TemplateConfigs(configContent [][]byte) []*template.Template {
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

func MergeConfigs(renderedConfigs []*bytes.Buffer) {
	for _, conf := range renderedConfigs {
		if err := viper.MergeConfig(conf); err != nil {
			log.WithError(err).Fatalf("failed while reading config '%s'", conf)
		}
	}
}
