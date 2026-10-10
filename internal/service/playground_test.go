package service

import (
	"strings"
	"testing"
	"time"
)

func TestChatPayloadDeepSeekThinking(t *testing.T) {
	body := chatPayload("deepseek", "deepseek-flash", "你是助手", "你好", "", nil)
	thinking, ok := body["thinking"].(map[string]string)
	if !ok || thinking["type"] != "enabled" {
		t.Fatalf("thinking = %#v", body["thinking"])
	}
	if body["reasoning_effort"] != "high" {
		t.Fatalf("reasoning_effort = %#v", body["reasoning_effort"])
	}
	if body["model"] != "deepseek-flash" {
		t.Fatalf("model = %#v", body["model"])
	}
}

func TestChatPayloadVisionOmitsThinkingForGrok(t *testing.T) {
	body := chatPayload("xai", "grok-4.7", "", "看图", "https://example.com/a.png", nil)
	if _, ok := body["thinking"]; ok {
		t.Fatal("grok payload should not force thinking")
	}
	messages := body["messages"].([]any)
	content := messages[0].(map[string]any)["content"].([]any)
	if len(content) != 2 {
		t.Fatalf("content parts = %d", len(content))
	}
}

func TestExtractChatKeepsReasoning(t *testing.T) {
	raw := []byte(`{"choices":[{"message":{"content":"答案","reasoning_content":"先想一下"}}]}`)
	text, reasoning := extractChat(raw)
	if text != "答案" || reasoning != "先想一下" {
		t.Fatalf("text=%q reasoning=%q", text, reasoning)
	}
}

func TestPresentVideoPending(t *testing.T) {
	out := presentVideo("https://api.x.ai/v1/videos/abc", 200, []byte(`{"request_id":"abc","status":"pending","progress":40}`), time.Now(), "abc")
	if !out.Pending || out.RequestID != "abc" || out.Progress != 40 || out.VideoURL != "" {
		t.Fatalf("%+v", out)
	}
}

func TestRedactHidesImageBase64(t *testing.T) {
	raw := []byte(`{"data":[{"b64_json":"` + strings.Repeat("A", 500) + `"}]}`)
	out := string(redactRaw(raw))
	if strings.Contains(out, strings.Repeat("A", 500)) {
		t.Fatal("base64 was not redacted")
	}
	if !strings.Contains(out, "省略") {
		t.Fatal(out)
	}
}
