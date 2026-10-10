package openai

import "testing"

func TestEndpoint(t *testing.T) {
	cases := map[string]string{
		"https://api.deepseek.com/v1":                             Endpoint("https://api.deepseek.com/v1", "chat/completions"),
		"https://api.deepseek.com":                                Endpoint("https://api.deepseek.com", "chat/completions"),
		"https://open.bigmodel.cn/api/paas/v4":                    Endpoint("https://open.bigmodel.cn/api/paas/v4/", "models"),
		"https://generativelanguage.googleapis.com/v1beta/openai": Endpoint("https://generativelanguage.googleapis.com/v1beta/openai", "chat/completions"),
	}
	want := map[string]string{
		"https://api.deepseek.com/v1":                             "https://api.deepseek.com/v1/chat/completions",
		"https://api.deepseek.com":                                "https://api.deepseek.com/v1/chat/completions",
		"https://open.bigmodel.cn/api/paas/v4":                    "https://open.bigmodel.cn/api/paas/v4/models",
		"https://generativelanguage.googleapis.com/v1beta/openai": "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions",
	}
	for in, got := range cases {
		if got != want[in] {
			t.Fatalf("%s => %s, want %s", in, got, want[in])
		}
	}
}
