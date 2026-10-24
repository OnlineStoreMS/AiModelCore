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
			Capabilities: []string{"chat", "vision", "image", "video", "voice", "embedding"},
			Models: []SuggestedModel{
				{ModelID: "qwen3.8-max", DisplayName: "Qwen3.8 Max", Capabilities: []string{"chat", "vision"}},
				{ModelID: "qwen3.7-plus", DisplayName: "Qwen3.7 Plus", Capabilities: []string{"chat", "vision"}},
				{ModelID: "qwen3.8-flash", DisplayName: "Qwen3.8 Flash", Capabilities: []string{"chat"}},
				{ModelID: "qwen3.8-omni-flash", DisplayName: "Qwen3.8 Omni Flash", Capabilities: []string{"chat", "vision"}},
				{ModelID: "qwen-image-3.0-pro", DisplayName: "Qwen Image 3.0 Pro", Capabilities: []string{"image"}},
				{ModelID: "qwen-image-3.0", DisplayName: "Qwen Image 3.0", Capabilities: []string{"image"}},
				{ModelID: "wan2.7-image-pro", DisplayName: "Wan2.7 Image Pro", Capabilities: []string{"image"}},
				{ModelID: "wan3.0-video", DisplayName: "Wan3.0 Video", Capabilities: []string{"video"}},
				{ModelID: "happyhorse-1.1-t2v", DisplayName: "HappyHorse 1.1 文生视频", Capabilities: []string{"video"}},
				{ModelID: "happyhorse-1.1-i2v", DisplayName: "HappyHorse 1.1 图生视频", Capabilities: []string{"video"}},
				{ModelID: "happyhorse-1.1-r2v", DisplayName: "HappyHorse 1.1 参考生视频", Capabilities: []string{"video"}},
				{ModelID: "happyhorse-1.0-video-edit", DisplayName: "HappyHorse 1.0 视频编辑", Capabilities: []string{"video"}},
				{ModelID: "qwen-audio-3.0-tts-plus", DisplayName: "Qwen Audio 3.0 TTS Plus", Capabilities: []string{"voice"}},
				{ModelID: "qwen-audio-3.1-asr-flash-streaming", DisplayName: "Qwen Audio 3.1 ASR 流式", Capabilities: []string{"voice"}},
				{ModelID: "qwen-audio-3.1-asr-flash-filetrans", DisplayName: "Qwen Audio 3.1 ASR 文件", Capabilities: []string{"voice"}},
				{ModelID: "qwen-audio-3.1-realtime-plus", DisplayName: "Qwen Audio 3.1 Realtime", Capabilities: []string{"voice"}},
				{ModelID: "qwen3.5-omni-plus", DisplayName: "Qwen3.5 Omni Plus", Capabilities: []string{"voice"}},
				{ModelID: "qwen3.8-omni-flash-realtime", DisplayName: "Qwen3.8 Omni Flash Realtime", Capabilities: []string{"voice"}},
				{ModelID: "qwen3.7-text-embedding", DisplayName: "Qwen3.7 Text Embedding", Capabilities: []string{"embedding"}},
				{ModelID: "qwen3.7-text-embedding-flash", DisplayName: "Qwen3.7 Text Embedding Flash", Capabilities: []string{"embedding"}},
				{ModelID: "qwen3-vl-embedding", DisplayName: "Qwen3 VL Embedding", Capabilities: []string{"embedding"}},
			},
			Notes: "地址填 openAiCompatible，业务空间在域名里。模型 ID 以百炼「选择模型」页的千问当前推荐为准。视频、语音和向量已登记，内部调用尚未接通。",
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
			Notes: "模型 ID 以 xAI 文档为准。文生图走 /images/generations。访问 api.x.ai 时在接入里填写 HTTP 代理。视频和语音已登记，内部调用尚未接通。",
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
			Notes:        "内网或本机的 OpenAI 兼容服务，例如 vLLM、Ollama、LM Studio。地址如 http://192.168.3.41:8000/v1，密钥可留空。客服识图把该模型能力勾成「识图」并设为默认。",
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
