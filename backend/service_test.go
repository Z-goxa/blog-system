package main

import (
	"testing"
	"time"
)

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string // 只校验前缀，因为尾部有时间戳
	}{
		{"普通英文标题", "Hello World", "hello-world-"},
		{"混合大小写", "Go Programming 101", "go-programming-101-"},
		{"中文+英文", "你好 Golang", "golang-"},
		{"纯特殊字符", "!!! ???", "tag-"},
		{"开头和结尾有空格/连字符", "- Spaces Around -", "spaces-around-"},
		{"连续符号", "a---b___c", "a-b-c-"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateSlug(tt.input)
			if len(got) <= len(tt.want) {
				t.Errorf("generateSlug(%q) = %q, want prefix %q", tt.input, got, tt.want)
			}
			if got[:len(tt.want)] != tt.want {
				t.Errorf("generateSlug(%q) = %q, want prefix %q", tt.input, got, tt.want)
			}
			// 确认以 -XXXXX 结尾（时间戳后缀）
			suffix := got[len(tt.want):]
			if len(suffix) < 1 {
				t.Errorf("缺少时间戳后缀")
			}
		})
	}
}

func TestGenerateSlugEmpty(t *testing.T) {
	t.Run("空字符串", func(t *testing.T) {
		got := generateSlug("")
		if len(got) <= len("tag-") {
			t.Errorf("空输入应使用默认前缀，got %q", got)
		}
		if got[:4] != "tag-" {
			t.Errorf("空输入应以前缀 tag- 开头，got %q", got)
		}
	})
}

func TestPasswordHashAndVerify(t *testing.T) {
	original := "MySecret@123!测试"

	hash, err := hashPassword(original)
	if err != nil {
		t.Fatalf("hashPassword 失败: %v", err)
	}
	if hash == "" {
		t.Fatal("hash 不应为空")
	}
	if hash == original {
		t.Fatal("hash 不应等于明文")
	}

	// 正确密码应通过
	if err := verifyPassword(hash, original); err != nil {
		t.Errorf("verifyPassword 正确密码失败: %v", err)
	}

	// 错误密码应失败
	if err := verifyPassword(hash, "wrong-password"); err == nil {
		t.Error("verifyPassword 错误密码应返回错误")
	}

	// 空密码
	if err := verifyPassword(hash, ""); err == nil {
		t.Error("verifyPassword 空密码应返回错误")
	}

	// 同一密码多次 hash 结果不同（bcrypt 自动加盐）
	hash2, _ := hashPassword(original)
	if hash == hash2 {
		t.Error("bcrypt 两次 hash 不应相同")
	}
}

func TestValidateToken(t *testing.T) {
	user := &User{
		ID:       1,
		Username: "testuser",
		Role:     "admin",
	}

	token, err := generateToken(user)
	if err != nil {
		t.Fatalf("generateToken 失败: %v", err)
	}
	if token == "" {
		t.Fatal("token 不应为空")
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken 失败: %v", err)
	}
	if claims.UserID != 1 {
		t.Errorf("UserID 应为 1, got %d", claims.UserID)
	}
	if claims.Username != "testuser" {
		t.Errorf("Username 应为 testuser, got %q", claims.Username)
	}
	if claims.Role != "admin" {
		t.Errorf("Role 应为 admin, got %q", claims.Role)
	}
	if !claims.ExpiresAt.After(time.Now()) {
		t.Error("token 应未过期")
	}

	// 无效 token
	if _, err := ValidateToken("invalid_token_string"); err == nil {
		t.Error("无效 token 应返回错误")
	}
}

func TestRenderMarkdown(t *testing.T) {
	tests := []struct {
		name string
		md   string
		check func(t *testing.T, html string)
	}{
		{
			name: "普通段落",
			md:   "Hello World",
			check: func(t *testing.T, html string) {
				if html == "" {
					t.Fatal("html 不应为空")
				}
			},
		},
		{
			name: "标题",
			md:   "# Title\n\nParagraph",
			check: func(t *testing.T, html string) {
				if !contains(html, "<h1>") {
					t.Errorf("应包含 <h1>，实际: %s", html)
				}
				if !contains(html, "Title") {
					t.Errorf("应包含 Title")
				}
			},
		},
		{
			name: "代码块",
			md:   "```go\nfmt.Println(\"hi\")\n```",
			check: func(t *testing.T, html string) {
				if !contains(html, "<pre") {
					t.Errorf("代码块应包含 <pre>，实际: %s", html)
				}
				if !contains(html, "fmt.Println") {
					t.Errorf("代码块应保留代码内容")
				}
			},
		},
		{
			name: "加粗和斜体",
			md:   "**bold** and *italic*",
			check: func(t *testing.T, html string) {
				if !contains(html, "<strong>") {
					t.Errorf("应包含 <strong>")
				}
				if !contains(html, "<em>") {
					t.Errorf("应包含 <em>")
				}
			},
		},
		{
			name: "无序列表",
			md:   "- Item 1\n- Item 2",
			check: func(t *testing.T, html string) {
				if !contains(html, "<li>") {
					t.Errorf("应包含 <li>")
				}
				if !contains(html, "Item 1") {
					t.Errorf("应包含 Item 1")
				}
			},
		},
		{
			name: "空输入",
			md:   "",
			check: func(t *testing.T, html string) {
				if html != "" {
					t.Errorf("空输入应返回空字符串，got %q", html)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderMarkdown(tt.md)
			tt.check(t, got)
		})
	}
}

func TestRenderMarkdownTable(t *testing.T) {
	md := "| A | B |\n|---|---|\n| 1 | 2 |"
	html := renderMarkdown(md)
	if !contains(html, "<table") {
		t.Errorf("表格应包含 <table>，实际: %s", html)
	}
	if !contains(html, "<th>") {
		t.Errorf("表格应包含 <th>")
	}
}

// contains 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}