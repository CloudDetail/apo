// Copyright 2015 Prometheus Team
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package notification

import (
	"bytes"
	sprig "github.com/go-task/slim-sprig"
	"io"
	"text/template"
)

type Template struct {
	tmpl *template.Template
}

func FromDefault() (*Template, error) {
	tmpl := template.New("").
		Option("missingkey=zero").
		Funcs(defaultFuncs).
		Funcs(sprig.TxtFuncMap())

	f, err := Assets.Open("/templates/default.tmpl")
	if err != nil {
		return nil, err
	}

	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	if _, err := tmpl.Parse(string(b)); err != nil {
		return nil, err
	}

	return &Template{tmpl: tmpl}, nil
}

func (t *Template) ExecuteTextString(text string, data interface{}) (string, error) {
	if text == "" {
		return "", nil
	}
	tmpl, err := t.tmpl.Clone()
	if err != nil {
		return "", err
	}
	tmpl, err = tmpl.New("").Option("missingkey=zero").Parse(text)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	return buf.String(), err
}
