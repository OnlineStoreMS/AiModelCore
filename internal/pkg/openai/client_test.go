package openai

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

func TestCallTraffic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()
	ctx, traffic := WithTraffic(context.Background())
	_, raw, err := New().Call(ctx, http.MethodPost, srv.URL, "k", map[string]string{"a": "b"})
	if err != nil {
		t.Fatal(err)
	}
	if traffic.RequestBytes != int64(len(`{"a":"b"}`)) || traffic.ResponseBytes != int64(len(raw)) {
		t.Fatalf("traffic %+v raw %d", traffic, len(raw))
	}
}
