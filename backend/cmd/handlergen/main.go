// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"go/token"
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

var handlerName string

const handlerTemplate = `// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package %s

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

 // TODO move to model/request package
type %sRequest struct {
}

 // TODO move to model/response package
type %sResponse struct {
}

%s
// @Summary%s
// @Description%s
%s
// @Accept application/x-www-form-urlencoded
// @Produce json
 // TODO The following request parameter types and response types must be changed according to actual requirements.
 // @Param Request body request.%sRequest true "Request information"
// @Param Authorization header string false "Bearer accessToken"
// @Success 200 {object} response.%sResponse
// @Failure 400 {object} code.Failure
%s
func (h *handler) %s() core.HandlerFunc {
	return func(c core.Context) {
		req := new(%sRequest)
 // TODO Adjust the API based on the request parameter type
		if err := c.ShouldBindQuery(req); err != nil {
			c.AbortWithError(core.Error(
				http.StatusBadRequest,
				code.ParamBindError,
				code.Text(code.ParamBindError)).WithError(err),
			)
			return
		}

 // TODO replace with Service call
		resp := new(%sResponse)
		c.Payload(resp)
	}
}
`

func init() {
	handler := flag.String("handler", "", "请输入需要生成的 handler 名称\n")
	flag.Parse()

	handlerName = strings.ToLower(*handler)
}

func main() {
	fs := token.NewFileSet()
	filePath := fmt.Sprintf("./pkg/api/%s", handlerName)
	parsedFile, err := decorator.ParseFile(fs, filePath+"/handler.go", nil, 0)
	if err != nil {
		log.Fatalf("parsing package: %s: %s\n", filePath, err)
	}

	dst.Inspect(parsedFile, func(n dst.Node) bool {
		decl, ok := n.(*dst.GenDecl)
		if !ok || decl.Tok != token.TYPE {
			return true
		}

		for _, spec := range decl.Specs {
			typeSpec, _ok := spec.(*dst.TypeSpec)
			if !_ok {
				continue
			}

			var interfaceType *dst.InterfaceType
			if interfaceType, ok = typeSpec.Type.(*dst.InterfaceType); !ok {
				continue
			}

			for _, v := range interfaceType.Methods.List {
				if len(v.Names) > 0 {
					if v.Names[0].String() == "i" {
						continue
					}

					filepath := "./pkg/api/" + handlerName
					filename := fmt.Sprintf("%s/func_%s.go", filepath, strings.ToLower(v.Names[0].String()))
					if _, err := os.Stat(filename); !os.IsNotExist(err) {
						continue
					}

					funcFile, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0766)
					if err != nil {
						fmt.Printf("create and open func file error %v\n", err.Error())
						continue
					}

					if funcFile == nil {
						fmt.Printf("func file is nil \n")
						continue
					}

					fmt.Println("  └── file : ", filename)

					comments := v.Decorations().Start.All()
					methodName := v.Names[0].String()
					methodKey := Lcfirst(methodName)
					methodDesc := strings.Split(comments[0], methodName)[1]
					funcContent := fmt.Sprintf(handlerTemplate,
						handlerName,
						methodKey,
						methodKey,
						comments[0], // first line comment
						methodDesc,
						methodDesc,
						comments[1], // Tags
						methodKey,
						methodKey,
						comments[2], // Router
						methodName,
						methodKey,
						methodKey,
					)
					funcFile.WriteString(funcContent)
					funcFile.Close()
				}
			}
		}
		return true
	})
}

func Lcfirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}
