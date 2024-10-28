// Copyright 2018 The Prometheus Authors
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

//go:build !builtinassets

package notification

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/shurcooL/httpfs/filter"
	"github.com/shurcooL/httpfs/union"
)

// Assets contains the project's assets.
var Assets = func() http.FileSystem {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	wd = strings.Replace(wd, "\\", "/", -1)
	var assetsPrefix string
	switch path.Base(wd) {
	case "backend":
		// When running prometheus-webhook-dingtalk (without built-in assets) from the repo root.
		assetsPrefix = "./pkg/services/alerts/"
	case "notification":
		// When generating statically compiled-in assets.
		assetsPrefix = "../"
	case "alerts":
		assetsPrefix = "./"
	}

	templates := filter.Keep(
		http.Dir(path.Join(assetsPrefix, "notification")),
		func(path string, fi os.FileInfo) bool {
			return path == "/" || path == "/default.tmpl"
		},
	)
	fmt.Println(templates)
	return union.New(map[string]http.FileSystem{
		"/templates": templates,
	})
}()
