// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package permission

import (
	"strings"
	"testing"
)

func TestCheckRouterMatch(t *testing.T) {
	tests := []struct {
		name         string
		checkRouter  string
		matchRouter  string
		want         bool
		errorMessage string
	}{
		{
			name:         "完全匹配的静态路由",
			checkRouter:  "/user/profile",
			matchRouter:  "/user/profile",
			want:         true,
			errorMessage: "静态路由应完全匹配",
		},
		{
			name:         "带动态参数的相同结构",
			checkRouter:  "/user/1",
			matchRouter:  "/user/:id",
			want:         true,
			errorMessage: "相同动态参数结构应匹配",
		},
		{
			name:         "路径长度不同",
			checkRouter:  "/user/123/profile",
			matchRouter:  "/user/:id",
			want:         false,
			errorMessage: "路径层级深度不同应不匹配",
		},
		{
			name:         "静态部分不匹配",
			checkRouter:  "/admin/profile",
			matchRouter:  "/user/profile",
			want:         false,
			errorMessage: "静态路径部分不匹配应返回失败",
		},
		{
			name:         "动态参数位置不匹配",
			checkRouter:  "/user/1/contacts",
			matchRouter:  "/user/profile/:id",
			want:         false,
			errorMessage: "动态参数位置不同应不匹配",
		},

		{
			name:         "根路径匹配",
			checkRouter:  "/",
			matchRouter:  "/",
			want:         true,
			errorMessage: "根路径应正确匹配",
		},
		{
			name:         "空路径处理",
			checkRouter:  "",
			matchRouter:  "",
			want:         true,
			errorMessage: "空路径应视为匹配",
		},
		{
			name:         "混合大小写路径",
			checkRouter:  "/User/Profile",
			matchRouter:  "/user/profile",
			want:         false,
			errorMessage: "应区分路径大小写",
		},

		{
			name:         "多级动态参数",
			checkRouter:  "/api/1/users/2",
			matchRouter:  "/api/:version/users/:id",
			want:         true,
			errorMessage: "多级动态参数结构应匹配",
		},
		{
			name:         "嵌套动态参数",
			checkRouter:  "/blog/2003/2/1",
			matchRouter:  "/blog/:year/:month/:slug",
			want:         true,
			errorMessage: "嵌套动态参数应匹配",
		},

		{
			name:         "含特殊字符的静态路径",
			checkRouter:  "/user/100%",
			matchRouter:  "/user/100%",
			want:         true,
			errorMessage: "特殊字符路径应精确匹配",
		},
		{
			name:         "含点的路径段",
			checkRouter:  "/file/v1.2.3",
			matchRouter:  "/file/v1.2.3",
			want:         true,
			errorMessage: "含点路径应精确匹配",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkRouterMatch(tt.checkRouter, tt.matchRouter)

			if got != tt.want {
				partsCheck := strings.Split(tt.checkRouter, "/")
				partsMatch := strings.Split(tt.matchRouter, "/")

				t.Errorf(
					"%s\nCheck路由: %s\nMatch路由: %s\n路径分解:\nCheck: %#v\nMatch: %#v\n预期: %v\n实际: %v",
					tt.errorMessage,
					tt.checkRouter,
					tt.matchRouter,
					partsCheck,
					partsMatch,
					tt.want,
					got,
				)
			}
		})
	}
}
