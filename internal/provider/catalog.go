package provider

type SuggestedModel struct {
	ModelID      string   `json:"modelId"`
	DisplayName  string   `json:"displayName"`
	Capabilities []string `json:"capabilities"`
}

type Item struct {
	Code           string           `json:"code"`
	Name           string           `json:"name"`
	Vendor         string           `json:"vendor"`
	DefaultBaseURL string           `json:"defaultBaseUrl"`
	Protocol       string           `json:"protocol"`
	Capabilities   []string         `json:"capabilities"`
	Models         []SuggestedModel `json:"models"`
	Notes          string           `json:"notes"`
}

func Catalog() []Item {
	return []Item{
		{
			Code: "deepseek", Name: "DeepSeek", Vendor: "深度求索",
			DefaultBaseURL: "https://api.deepseek.com/v1", Protocol: "openai",
			Capabilities: []string{"chat", "vision"},
			Models: []SuggestedModel{
				{ModelID: "deepseek-flash", DisplayName: "DeepSeek V4.1 Flash", Capabilities: []string{"chat", "vision"}},
				{ModelID: "deepseek-v4-pro", DisplayName: "DeepSeek V4 Pro", Capabilities: []string{"chat"}},
			},
			Notes: "模型 ID 以 DeepSeek 文档为准。deepseek-flash 对应 V4.1-Flash，支持识图；deepseek-v4-pro 不支持识图。deepseek-chat 与 deepseek-reasoner 已于 2026-07-24 停用。",
		},
		{
			Code: "openai", Name: "OpenAI", Vendor: "OpenAI",
			DefaultBaseURL: "https://api.openai.com/v1", Protocol: "openai",
			Capabilities: []string{"chat", "vision", "image", "embedding"},
			Models: []SuggestedModel{
				{ModelID: "gpt-4o", DisplayName: "GPT-4o", Capabilities: []string{"chat", "vision"}},
				{ModelID: "gpt-4o-mini", DisplayName: "GPT-4o mini", Capabilities: []string{"chat", "vision"}},
				{ModelID: "gpt-image-1", DisplayName: "GPT Image", Capabilities: []string{"image"}},
			},
			Notes: "对话、识图与图片生成/编辑走官方兼容路径，供商品图片修改调用。",
		},
		{
			Code: "dashscope", Name: "阿里云百炼", Vendor: "阿里云",
			DefaultBaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1", Protocol: "openai",
			Capabilities: []string{"chat", "vision", "image", "embedding"},
			Models: []SuggestedModel{
				{ModelID: "qwen-plus", DisplayName: "通义千问 Plus", Capabilities: []string{"chat"}},
				{ModelID: "qwen-vl-max", DisplayName: "通义千问 VL Max", Capabilities: []string{"chat", "vision"}},
				{ModelID: "qwen-image", DisplayName: "通义万相", Capabilities: []string{"image"}},
			},
			Notes: "百炼 OpenAI 兼容模式。图片编辑能力以模型实际支持为准。",
		},
		{
			Code: "xai", Name: "Grok", Vendor: "xAI",
			DefaultBaseURL: "https://api.x.ai/v1", Protocol: "openai",
			Capabilities: []string{"chat", "vision", "image", "video", "voice"},
			Models: []SuggestedModel{
				{ModelID: "grok-4.7", DisplayName: "Grok 4.7", Capabilities: []string{"chat", "vision"}},
				{ModelID: "grok-4.6", DisplayName: "Grok 4.6", Capabilities: []string{"chat", "vision"}},
				{ModelID: "grok-4.5", DisplayName: "Grok 4.5", Capabilities: []string{"chat", "vision"}},
				{ModelID: "grok-4.3", DisplayName: "Grok 4.3", Capabilities: []string{"chat", "vision"}},
				{ModelID: "grok-4.20-0309-reasoning", DisplayName: "Grok 4.20 Reasoning", Capabilities: []string{"chat"}},
				{ModelID: "grok-4.20-0309-non-reasoning", DisplayName: "Grok 4.20", Capabilities: []string{"chat"}},
				{ModelID: "grok-build-0.1", DisplayName: "Grok Build 0.1", Capabilities: []string{"chat"}},
				{ModelID: "grok-4.20-multi-agent-0309", DisplayName: "Grok 4.20 Multi-Agent", Capabilities: []string{"chat"}},
				{ModelID: "grok-imagine-image-2.0", DisplayName: "Grok Imagine Image 2.0", Capabilities: []string{"image"}},
				{ModelID: "grok-imagine-image", DisplayName: "Grok Imagine Image", Capabilities: []string{"image"}},
				{ModelID: "grok-imagine-image-quality", DisplayName: "Grok Imagine Image Quality", Capabilities: []string{"image"}},
				{ModelID: "grok-imagine-video-1.5", DisplayName: "Grok Imagine Video 1.5", Capabilities: []string{"video"}},
				{ModelID: "grok-imagine-video-1.5-lite", DisplayName: "Grok Imagine Video 1.5 Lite", Capabilities: []string{"video"}},
				{ModelID: "grok-imagine-video", DisplayName: "Grok Imagine Video", Capabilities: []string{"video"}},
				{ModelID: "grok-voice-think-fast-2.0", DisplayName: "Grok Voice Think Fast 2.0", Capabilities: []string{"voice"}},
				{ModelID: "grok-voice-transcribe-2.0", DisplayName: "Grok Voice Transcribe 2.0", Capabilities: []string{"voice"}},
			},
			Notes: "模型 ID 以 xAI 文档为准。文生图走 /images/generations。视频和语音已登记，内部调用尚未接通。",
		},
		{
			Code: "moonshot", Name: "Kimi", Vendor: "月之暗面",
			DefaultBaseURL: "https://api.moonshot.cn/v1", Protocol: "openai",
			Capabilities: []string{"chat", "vision"},
			Models: []SuggestedModel{
				{ModelID: "moonshot-v1-auto", DisplayName: "Kimi Auto", Capabilities: []string{"chat"}},
				{ModelID: "moonshot-v1-128k-vision-preview", DisplayName: "Kimi Vision", Capabilities: []string{"chat", "vision"}},
			},
			Notes: "Moonshot OpenAI 兼容接口。",
		},
		{
			Code: "zhipu", Name: "智谱", Vendor: "智谱 AI",
			DefaultBaseURL: "https://open.bigmodel.cn/api/paas/v4", Protocol: "openai",
			Capabilities: []string{"chat", "vision", "image", "embedding"},
			Models: []SuggestedModel{
				{ModelID: "glm-4-plus", DisplayName: "GLM-4 Plus", Capabilities: []string{"chat"}},
				{ModelID: "glm-4v-plus", DisplayName: "GLM-4V Plus", Capabilities: []string{"chat", "vision"}},
			},
			Notes: "BigModel v4，路径已带版本号。",
		},
		{
			Code: "qianfan", Name: "百度千帆", Vendor: "百度",
			DefaultBaseURL: "https://qianfan.baidubce.com/v2", Protocol: "openai",
			Capabilities: []string{"chat", "vision", "image", "embedding"},
			Models: []SuggestedModel{
				{ModelID: "ernie-4.5-turbo-128k", DisplayName: "ERNIE 4.5 Turbo", Capabilities: []string{"chat"}},
			},
			Notes: "千帆 v2 兼容接口。",
		},
		{
			Code: "volcengine", Name: "火山方舟", Vendor: "字节跳动",
			DefaultBaseURL: "https://ark.cn-beijing.volces.com/api/v3", Protocol: "openai",
			Capabilities: []string{"chat", "vision", "image"},
			Models: []SuggestedModel{
				{ModelID: "doubao-seed-1-6-250615", DisplayName: "豆包 Seed", Capabilities: []string{"chat", "vision"}},
			},
			Notes: "方舟推理接入点也可把接入点 ID 填成模型名。",
		},
		{
			Code: "siliconflow", Name: "硅基流动", Vendor: "SiliconFlow",
			DefaultBaseURL: "https://api.siliconflow.cn/v1", Protocol: "openai",
			Capabilities: []string{"chat", "vision", "image", "embedding"},
			Models: []SuggestedModel{
				{ModelID: "deepseek-ai/DeepSeek-V3", DisplayName: "DeepSeek V3", Capabilities: []string{"chat"}},
			},
			Notes: "聚合多家模型的 OpenAI 兼容网关。",
		},
		{
			Code: "gemini", Name: "Gemini", Vendor: "Google",
			DefaultBaseURL: "https://generativelanguage.googleapis.com/v1beta/openai", Protocol: "openai",
			Capabilities: []string{"chat", "vision", "image"},
			Models: []SuggestedModel{
				{ModelID: "gemini-2.0-flash", DisplayName: "Gemini 2.0 Flash", Capabilities: []string{"chat", "vision"}},
			},
			Notes: "使用 Gemini 的 OpenAI 兼容端点。",
		},
		{
			Code: "anthropic", Name: "Claude", Vendor: "Anthropic",
			DefaultBaseURL: "https://api.anthropic.com/v1", Protocol: "openai",
			Capabilities: []string{"chat", "vision"},
			Models: []SuggestedModel{
				{ModelID: "claude-sonnet-4-5", DisplayName: "Claude Sonnet", Capabilities: []string{"chat", "vision"}},
			},
			Notes: "官方协议不是 OpenAI。若直连失败，把地址改成兼容网关。",
		},
		{
			Code: "private", Name: "私有模型", Vendor: "自建",
			DefaultBaseURL: "", Protocol: "openai",
			Capabilities: []string{"chat", "vision", "image", "embedding"},
			Models:       nil,
			Notes:        "填写 OpenAI 兼容的 Base URL，例如 http://10.0.0.8:8000/v1。",
		},
	}
}

func Find(code string) (Item, bool) {
	for _, item := range Catalog() {
		if item.Code == code {
			return item, true
		}
	}
	return Item{}, false
}
