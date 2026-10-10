import client, { unwrap } from './client'

export interface ProviderModel {
  modelId: string
  displayName: string
  capabilities: string[]
}

export interface ProviderItem {
  code: string
  name: string
  vendor: string
  defaultBaseUrl: string
  protocol: string
  capabilities: string[]
  models: ProviderModel[]
  notes: string
}

export interface ConnectionItem {
  id: number
  name: string
  providerCode: string
  providerName: string
  baseUrl: string
  apiKeyHint: string
  hasApiKey: boolean
  enabled: number
  remark: string
}

export interface ConnectionInput {
  name: string
  providerCode: string
  baseUrl: string
  apiKey?: string
  enabled: number
  remark: string
}

export interface ModelItem {
  id: number
  connectionId: number
  connectionName: string
  providerCode: string
  modelId: string
  displayName: string
  capabilities: string[]
  isDefault: number
  enabled: number
  remark: string
}

export interface ModelInput {
  connectionId: number
  modelId: string
  displayName: string
  capabilities: string[]
  isDefault: number
  enabled: number
  remark: string
}

export interface DashboardSummary {
  connections: number
  enabledConnections: number
  models: number
  todayCalls: number
  todayTokens: number
  todayErrors: number
  balanceReady: boolean
  balanceNote: string
}

export interface UsageItem {
  id: number
  callerApp: string
  capability: string
  modelName: string
  promptTokens: number
  completionTokens: number
  totalTokens: number
  status: string
  errorMessage: string
  latencyMs: number
  createdAt: string
}

export interface TestResult {
  ok: boolean
  message: string
  models?: string[]
}

export const capabilityOptions = [
  { value: 'chat', label: '对话' },
  { value: 'vision', label: '识图' },
  { value: 'image', label: '图片生成/编辑' },
  { value: 'embedding', label: '向量' },
  { value: 'video', label: '视频' },
  { value: 'voice', label: '语音' },
]

export function capabilityLabel(code: string) {
  return capabilityOptions.find((item) => item.value === code)?.label || code
}

export const listProviders = () => unwrap<ProviderItem[]>(client.get('/providers'))
export const getDashboard = () => unwrap<DashboardSummary>(client.get('/dashboard'))
export const listUsage = () => unwrap<UsageItem[]>(client.get('/usage'))
export const listConnections = () => unwrap<ConnectionItem[]>(client.get('/connections'))
export const createConnection = (body: ConnectionInput) => unwrap<ConnectionItem>(client.post('/connections', body))
export const updateConnection = (id: number, body: ConnectionInput) => unwrap<ConnectionItem>(client.put(`/connections/${id}`, body))
export const deleteConnection = (id: number) => unwrap(client.delete(`/connections/${id}`))
export const testConnection = (id: number) => unwrap<TestResult>(client.post(`/connections/${id}/test`))
export const suggestModels = (id: number) => unwrap<{ added: number }>(client.post(`/connections/${id}/suggest-models`))
export const listModels = (connectionId?: number) =>
  unwrap<ModelItem[]>(client.get('/models', { params: connectionId ? { connectionId } : {} }))
export const createModel = (body: ModelInput) => unwrap<ModelItem>(client.post('/models', body))
export const updateModel = (id: number, body: ModelInput) => unwrap<ModelItem>(client.put(`/models/${id}`, body))
export const deleteModel = (id: number) => unwrap(client.delete(`/models/${id}`))

export interface PlaygroundInput {
  capability: string
  prompt?: string
  system?: string
  imageUrl?: string
  imageBase64?: string
  audioBase64?: string
  audioName?: string
  language?: string
  thinking?: boolean
  duration?: number
  aspectRatio?: string
  resolution?: string
  videoRequestId?: string
}

export interface PlaygroundResult {
  ok: boolean
  endpoint?: string
  status?: number
  latencyMs?: number
  pending: boolean
  requestId?: string
  progress?: number
  phase?: string
  text?: string
  reasoning?: string
  imageUrl?: string
  imageData?: string
  videoUrl?: string
  raw?: unknown
  error?: string
}

export const playground = (id: number, body: PlaygroundInput) =>
  unwrap<PlaygroundResult>(client.post(`/models/${id}/playground`, body, { timeout: 180000 }))
