<script setup lang="ts">
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { capabilityLabel, playground, type ModelItem, type PlaygroundResult } from '../api/aimodel'

const open = defineModel<boolean>({ default: false })
const props = defineProps<{ model: ModelItem | null }>()

const sending = ref(false)
const capability = ref('chat')
const form = reactive({
  prompt: '',
  system: '',
  imageUrl: '',
  imageBase64: '',
  imageName: '',
  audioBase64: '',
  audioName: '',
  language: 'zh',
  thinking: true,
  duration: 5,
  aspectRatio: '16:9',
  resolution: '720p',
})
const result = ref<PlaygroundResult | null>(null)
let pollTimer = 0

const caps = computed(() => props.model?.capabilities || [])
const realtimeVoice = computed(() => capability.value === 'voice' && !props.model?.modelId.includes('transcribe'))
const rawText = computed(() => {
  if (!result.value?.raw) return ''
  return JSON.stringify(result.value.raw, null, 2)
})

watch(open, (visible) => {
  if (!visible) {
    stopPoll()
    return
  }
  capability.value = preferredCap(caps.value)
  form.prompt = ''
  form.system = ''
  form.imageUrl = ''
  form.imageBase64 = ''
  form.imageName = ''
  form.audioBase64 = ''
  form.audioName = ''
  form.language = 'zh'
  form.thinking = props.model?.providerCode === 'deepseek'
  form.duration = 5
  form.aspectRatio = '16:9'
  form.resolution = '720p'
  result.value = null
})

function preferredCap(list: string[]) {
  for (const cap of ['chat', 'vision', 'image', 'video', 'voice']) {
    if (list.includes(cap)) return cap
  }
  return list[0] || 'chat'
}

function onImage(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return
  form.imageName = file.name
  const reader = new FileReader()
  reader.onload = () => {
    form.imageBase64 = String(reader.result || '')
  }
  reader.readAsDataURL(file)
}

function onAudio(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return
  form.audioName = file.name
  const reader = new FileReader()
  reader.onload = () => {
    form.audioBase64 = String(reader.result || '')
  }
  reader.readAsDataURL(file)
}

function stopPoll() {
  if (pollTimer) {
    window.clearTimeout(pollTimer)
    pollTimer = 0
  }
}

async function send(videoRequestId = '') {
  if (!props.model) return
  const polling = videoRequestId !== ''
  if (!polling) {
    stopPoll()
    sending.value = true
  }
  try {
    const data = await playground(props.model.id, {
      capability: capability.value,
      prompt: form.prompt,
      system: form.system,
      imageUrl: form.imageUrl,
      imageBase64: capability.value === 'vision' ? form.imageBase64 : '',
      audioBase64: capability.value === 'voice' ? form.audioBase64 : '',
      audioName: form.audioName,
      language: form.language,
      thinking: props.model.providerCode === 'deepseek' ? form.thinking : undefined,
      duration: form.duration,
      aspectRatio: form.aspectRatio,
      resolution: form.resolution,
      videoRequestId,
    })
    result.value = data
    if (data.pending && data.requestId && open.value) {
      pollTimer = window.setTimeout(() => {
        void send(data.requestId)
      }, 4000)
    }
  } catch (err) {
    stopPoll()
    ElMessage.error((err as Error).message)
  } finally {
    if (!polling) sending.value = false
  }
}

onBeforeUnmount(stopPoll)
</script>

<template>
  <el-dialog
    v-model="open"
    :title="model ? `体验 ${model.displayName}` : '体验'"
    width="760px"
    destroy-on-close
  >
    <p v-if="model" class="meta">{{ model.connectionName }} · {{ model.modelId }}。这次请求会按厂商原价计费。</p>
    <el-radio-group v-if="caps.length > 1" v-model="capability" class="caps">
      <el-radio-button v-for="cap in caps" :key="cap" :value="cap">{{ capabilityLabel(cap) }}</el-radio-button>
    </el-radio-group>

    <el-alert
      v-if="realtimeVoice"
      type="info"
      :closable="false"
      title="这个模型的原生接口是实时语音对话，体验窗不能代替一轮通话。转录模型可以上传音频。"
      show-icon
    />

    <el-form v-else label-width="88px">
      <el-form-item v-if="capability === 'chat' || capability === 'vision'" label="系统提示">
        <el-input v-model="form.system" type="textarea" :rows="2" placeholder="可选" />
      </el-form-item>
      <el-form-item v-if="model?.providerCode === 'deepseek' && (capability === 'chat' || capability === 'vision')" label="思考">
        <el-switch v-model="form.thinking" active-text="开启" inactive-text="关闭" />
      </el-form-item>
      <el-form-item v-if="capability !== 'voice'" :label="capability === 'image' || capability === 'video' ? '描述' : '内容'">
        <el-input v-model="form.prompt" type="textarea" :rows="4" placeholder="发给模型的原文" />
      </el-form-item>
      <el-form-item v-if="capability === 'vision'" label="图片">
        <el-input v-model="form.imageUrl" placeholder="图片 http(s) 地址，或在下面选择本地文件" />
        <input class="file" type="file" accept="image/*" @change="onImage" />
        <span v-if="form.imageName" class="file-name">{{ form.imageName }}</span>
      </el-form-item>
      <template v-if="capability === 'video'">
        <el-form-item label="时长">
          <el-input-number v-model="form.duration" :min="1" :max="15" />
          <span class="unit">秒</span>
        </el-form-item>
        <el-form-item label="画幅">
          <el-select v-model="form.aspectRatio" style="width: 160px">
            <el-option label="16:9" value="16:9" />
            <el-option label="9:16" value="9:16" />
            <el-option label="1:1" value="1:1" />
          </el-select>
        </el-form-item>
        <el-form-item label="分辨率">
          <el-select v-model="form.resolution" style="width: 160px">
            <el-option label="720p" value="720p" />
            <el-option label="1080p" value="1080p" />
          </el-select>
        </el-form-item>
      </template>
      <el-form-item v-if="capability === 'voice'" label="音频">
        <input class="file" type="file" accept="audio/*,.mp3,.wav,.m4a,.webm" @change="onAudio" />
        <span v-if="form.audioName" class="file-name">{{ form.audioName }}</span>
      </el-form-item>
      <el-form-item v-if="capability === 'voice'" label="语言">
        <el-select v-model="form.language" style="width: 160px">
          <el-option label="中文" value="zh" />
          <el-option label="English" value="en" />
        </el-select>
      </el-form-item>
    </el-form>

    <div v-if="result" class="result">
      <el-alert v-if="result.error" type="error" :closable="false" :title="result.error" show-icon />
      <p v-if="result.pending" class="pending">生成中<span v-if="result.progress"> {{ result.progress }}%</span></p>
      <div v-if="result.reasoning" class="block">
        <div class="label">思考</div>
        <pre>{{ result.reasoning }}</pre>
      </div>
      <div v-if="result.text" class="block">
        <div class="label">回复</div>
        <pre>{{ result.text }}</pre>
      </div>
      <img v-if="result.imageData || result.imageUrl" class="preview" :src="result.imageData || result.imageUrl" alt="生成图片" />
      <video v-if="result.videoUrl" class="preview" :src="result.videoUrl" controls />
      <el-collapse v-if="rawText">
        <el-collapse-item title="原始响应">
          <p v-if="result.endpoint" class="endpoint">{{ result.status }} {{ result.endpoint }}<span v-if="result.latencyMs"> · {{ result.latencyMs }} ms</span></p>
          <pre>{{ rawText }}</pre>
        </el-collapse-item>
      </el-collapse>
    </div>

    <template #footer>
      <el-button v-if="result?.pending" @click="stopPoll">停止刷新</el-button>
      <el-button @click="open = false">关闭</el-button>
      <el-button v-if="!realtimeVoice" type="primary" :loading="sending" @click="send()">发送</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.meta { margin: 0 0 12px; color: #909399; font-size: 13px; }
.caps { margin-bottom: 16px; }
.file { margin-top: 8px; }
.file-name, .unit { margin-left: 8px; color: #606266; }
.result { margin-top: 8px; display: flex; flex-direction: column; gap: 12px; }
.pending { margin: 0; color: #e6a23c; }
.block .label { margin-bottom: 4px; color: #909399; font-size: 12px; }
pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  background: #f5f7fa;
  border-radius: 6px;
  padding: 10px 12px;
  max-height: 240px;
  overflow: auto;
}
.preview { max-width: 100%; border-radius: 6px; background: #f5f7fa; }
.endpoint { margin: 0 0 8px; color: #909399; font-size: 12px; }
</style>
