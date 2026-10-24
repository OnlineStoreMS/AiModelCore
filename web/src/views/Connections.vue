<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createConnection,
  deleteConnection,
  listConnections,
  listProviders,
  suggestModels,
  testConnection,
  updateConnection,
  type ConnectionItem,
  type ProviderItem,
} from '../api/aimodel'

const loading = ref(false)
const rows = ref<ConnectionItem[]>([])
const providers = ref<ProviderItem[]>([])
const dialog = ref(false)
const saving = ref(false)
const editingId = ref(0)
const form = reactive({
  name: '',
  providerCode: 'deepseek',
  baseUrl: '',
  httpProxy: '',
  apiKey: '',
  enabled: 1,
  remark: '',
})

const currentProvider = () => providers.value.find((item) => item.code === form.providerCode)

async function load() {
  loading.value = true
  try {
    const [list, catalog] = await Promise.all([listConnections(), listProviders()])
    rows.value = list
    providers.value = catalog
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = 0
  form.name = ''
  form.providerCode = 'deepseek'
  form.baseUrl = currentProvider()?.defaultBaseUrl || ''
  form.httpProxy = ''
  form.apiKey = ''
  form.enabled = 1
  form.remark = ''
  dialog.value = true
}

function openEdit(row: ConnectionItem) {
  editingId.value = row.id
  form.name = row.name
  form.providerCode = row.providerCode
  form.baseUrl = row.baseUrl
  form.httpProxy = row.httpProxy || ''
  form.apiKey = ''
  form.enabled = row.enabled
  form.remark = row.remark
  dialog.value = true
}

function onProviderChange() {
  const item = currentProvider()
  if (!editingId.value && item) form.baseUrl = item.defaultBaseUrl
}

async function save() {
  saving.value = true
  try {
    const body = { ...form }
    if (editingId.value) await updateConnection(editingId.value, body)
    else await createConnection(body)
    ElMessage.success('已保存')
    dialog.value = false
    await load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    saving.value = false
  }
}

async function test(row: ConnectionItem) {
  const result = await testConnection(row.id)
  if (result.ok) ElMessage.success(result.message)
  else ElMessage.error(result.message)
}

async function suggest(row: ConnectionItem) {
  const result = await suggestModels(row.id)
  ElMessage.success(result.added ? `已添加 ${result.added} 个推荐模型` : '推荐模型都已存在')
}

async function remove(row: ConnectionItem) {
  await ElMessageBox.confirm(`删除接入「${row.name}」及其模型？`, '确认', { type: 'warning' })
  await deleteConnection(row.id)
  ElMessage.success('已删除')
  await load()
}

onMounted(load)
</script>

<template>
  <el-card shadow="never">
    <template #header>
      <div class="toolbar">
        <span>厂商接入</span>
        <el-button type="primary" @click="openCreate">新建接入</el-button>
      </div>
    </template>
    <el-table v-loading="loading" :data="rows">
      <el-table-column prop="name" label="名称" min-width="140" />
      <el-table-column prop="providerName" label="厂商" width="140" />
      <el-table-column prop="baseUrl" label="地址" min-width="220" show-overflow-tooltip />
      <el-table-column prop="httpProxy" label="HTTP 代理" min-width="180" show-overflow-tooltip />
      <el-table-column label="密钥" width="140">
        <template #default="{ row }">{{ row.hasApiKey ? row.apiKeyHint : '未配置' }}</template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '启用' : '停用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="test(row)">测试</el-button>
          <el-button link type="primary" @click="suggest(row)">导入推荐模型</el-button>
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>

  <el-dialog v-model="dialog" :title="editingId ? '编辑接入' : '新建接入'" width="560px">
    <el-form label-width="108px">
      <el-form-item label="名称">
        <el-input v-model="form.name" placeholder="例如 百炼-商品图" />
      </el-form-item>
      <el-form-item label="厂商">
        <el-select v-model="form.providerCode" style="width: 100%" @change="onProviderChange">
          <el-option v-for="item in providers" :key="item.code" :label="`${item.name} · ${item.vendor}`" :value="item.code" />
        </el-select>
      </el-form-item>
      <el-form-item label="地址">
        <el-input v-model="form.baseUrl" placeholder="OpenAI 兼容 Base URL" />
      </el-form-item>
      <el-form-item label="HTTP 代理">
        <el-input v-model="form.httpProxy" placeholder="http://host.docker.internal:7890，留空则直连" />
      </el-form-item>
      <el-form-item label="密钥">
        <el-input v-model="form.apiKey" type="password" show-password :placeholder="form.providerCode === 'private' ? '本地模型可留空' : (editingId ? '留空则不修改' : 'API Key')" />
      </el-form-item>
      <el-form-item label="启用">
        <el-switch v-model="form.enabled" :active-value="1" :inactive-value="0" />
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="form.remark" type="textarea" :rows="2" />
      </el-form-item>
      <el-alert v-if="currentProvider()?.notes" :title="currentProvider()?.notes" type="info" :closable="false" />
    </el-form>
    <template #footer>
      <el-button @click="dialog = false">取消</el-button>
      <el-button type="primary" :loading="saving" @click="save">保存</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.toolbar { display: flex; align-items: center; justify-content: space-between; }
</style>
