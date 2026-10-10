# AI 大模型管理中心

统一配置 DeepSeek、OpenAI、阿里云百炼、Grok、Kimi、智谱、千帆、火山方舟、硅基流动、Gemini、Claude 以及私有 OpenAI 兼容模型。密钥只保存在本服务，其他 OSMS 应用通过内部接口调用。

| | |
|---|---|
| Web | `5194`，公网路径 `/apps/ai-model/` |
| API | `8109` |
| 镜像 | `aimodelcore-api`、`aimodelcore-web` |

内部调用（商品图片修改等）：

- `POST /api/v1/internal/chat` 对话 / 识图
- `POST /api/v1/internal/images` 生成；请求里带 `imageUrl` 或 `imageBase64` 时走编辑
- 请求头 `X-Internal-Token`

用量按调用写入。厂商余额拉取尚未对接，页面上先占位。
