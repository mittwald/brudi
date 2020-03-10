package mysqldump

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/mittwald/brudi/internal"
)

const (
	Kind = "mysqldump"
)

var myCnfTmpl = `[client]
{{ range $key, $value := . }}
{{- if isBool $value }}
{{- $key }}
{{- else }}
{{- $key }}={{ $value }}
{{- end }}
{{ end -}}
`

type Config struct {
	ClientOptions   map[string]interface{}
	clientMyCnfPath string
	Out             string
}

type CliFlags struct {
	DefaultsFile string `flag:"--defaults-file="`
}

func (c *Config) InitFromViper() error {
	err := viper.UnmarshalKey(Kind, &c)
	if err != nil {
		return errors.WithStack(err)
	}

	if c.Out == "" {
		return errors.WithStack(fmt.Errorf("no dump output path provided"))
	}

	return c.generateClientMyCnf()
}

func (c *Config) generateClientMyCnf() error {
	tt := template.New("").Funcs(internal.TemplateFunctions())

	tpl, err := tt.Parse(myCnfTmpl)
	if err != nil {
		return err
	}

	var file *os.File
	file, err = ioutil.TempFile(os.TempDir(), "mycnf")
	if err != nil {
		return errors.WithStack(err)
	}

	c.clientMyCnfPath = file.Name()

	return tpl.Execute(file, c.ClientOptions)
}
