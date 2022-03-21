package tmplhelper

import (
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/rakyll/statik/fs"
)

func LoadAll(ns string, items ...string) (*template.Template, error) {
	statikfs, err := fs.NewWithNamespace(ns)
	if err != nil {
		return nil, err
	}
	t := template.New("__root__")
	for _, fileName := range items {
		name := filepath.Base(fileName)[:len(fileName)-len(filepath.Ext(fileName))]
		file, err := statikfs.Open("/" + fileName)
		if err != nil {
			return nil, err
		}
		content, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, err
		}
		file.Close()
		_, err = t.New(name).Parse(string(content))
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}
