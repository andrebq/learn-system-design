package tmplhelper

import (
	"bytes"
	"io/ioutil"
	"text/template"
)

type (
	RenderConfig struct {
		Data       interface{}
		Name       string
		OutputFile string
	}
)

func RenderAll(rootTmpl *template.Template, data interface{}, targets ...RenderConfig) error {
	for _, v := range targets {
		if v.Data == nil {
			v.Data = data
		}
		err := Render(rootTmpl, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func Render(rootTmpl *template.Template, target RenderConfig) error {
	buf := bytes.Buffer{}
	err := rootTmpl.ExecuteTemplate(&buf, target.Name, target.Data)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(target.OutputFile, buf.Bytes(), 0644)
	if err != nil {
		return err
	}
	return nil
}
