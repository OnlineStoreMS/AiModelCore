<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  capabilityLabel,
  capabilityOptions,
  createModel,
  deleteModel,
  listConnections,
  listModels,
  updateModel,
  type ConnectionItem,
  type ModelItem,
} from '../api/aimodel'
import PlaygroundDialog from './PlaygroundDialog.vue'

const loading = ref(false)
const rows = ref<ModelItem[]>([])
const connections = ref<ConnectionItem[]>([])
const filterConnection = ref<number | undefined>()
const dialog = ref(false)
const playOpen = ref(false)
const playModel = ref<ModelItem | null>(null)
const saving = ref(false)
const editingId = ref(0)
const form = reactive({
  connectionId: undefined as number | undefined,
  modelId: '',
  displayName: '',
  capabilities: ['chat'] as string[],
  isDefault: 0,
  enabled: 1,
  remark: '',
})

async function load() {
  loading.value = true
  try {
    const [models, conns] = await Promise.all([
      listModels(filterConnection.value),
      listConnections(),
    ])
    rows.value = models
    connections.value = conns
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = 0
  form.connectionId = filterConnection.value || connections.value[0]?.id
  form.modelId = ''
  form.displayName = ''
  form.capabilities = ['chat']
  form.isDefault = 0
  form.enabled = 1
  form.remark = ''
  dialog.value = true
}

function openPlay(row: ModelItem) {
  playModel.value = row
  playOpen.value = true
}

function openEdit(row: ModelItem) {
  editingId.value = row.id
  form.connectionId = row.connectionId
  form.modelId = row.modelId
  form.displayName = row.displayName
  form.capabilities = [...row.capabilities]
  form.isDefault = row.isDefault
  form.enabled = row.enabled
  form.remark = row.remark
  dialog.value = true
}

async function save() {
  if (!form.connectionId) {
    ElMessage.warning('请选择接入')
    return
  }
  saving.value = true
  try {
    const body = { ...form, connectionId: form.connectionId }
    if (editingId.value) await updateModel(editingId.value, body)
    else await createModel(body)
    ElMessage.success('已保存')
    dialog.value = false
    await load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    saving.value = false
  }
}

async function remove(row: ModelItem) {
  await ElMessageBox.confirm(`删除模型「${row.displayName}」？`, '确认', { type: 'warning' })
  await deleteModel(row.id)
  ElMessage.success('已删除')
  await load()
}

onMounted(load)
</script>

<template>
  <el-card shadow="never">
    <template #header>
      <div class="toolbar">
        <div class="filters">
          <span>模型</span>
          <el-select v-model="filterConnection" clearable placeholder="全部接入" style="width: 220px" @change="load">
            <el-option v-for="item in connections" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </div>
        <el-button type="primary" @click="openCreate">添加模型</el-button>
      </div>
    </template>
    <el-table v-loading="loading" :data="rows">
      <el-table-column prop="displayName" label="名称" min-width="160" />
      <el-table-column prop="modelId" label="模型 ID" min-width="180" />
      <el-table-column prop="connectionName" label="接入" width="160" />
      <el-table-column label="能力" min-width="200">
        <template #default="{ row }">
          <el-tag v-for="cap in row.capabilities" :key="cap" class="cap" size="small">{{ capabilityLabel(cap) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="默认" width="80">
        <template #default="{ row }">{{ row.isDefault ? '是' : '' }}</template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '启用' : '停用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openPlay(row)">体验</el-button>
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>

  <el-dialog v-model="dialog" :title="editingId ? '编辑模型' : '添加模型'" width="560px">
    <el-form label-width="88px">
      <el-form-item label="接入">
        <el-select v-model="form.connectionId" style="width: 100%">
          <el-option v-for="item in connections" :key="item.id" :label="item.name" :value="item.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="模型 ID">
        <el-input v-model="form.modelId" placeholder="例如 qwen-vl-max" />
      </el-form-item>
      <el-form-item label="显示名">
        <el-input v-model="form.displayName" />
      </el-form-item>
      <el-form-item label="能力">
        <el-checkbox-group v-model="form.capabilities">
          <el-checkbox v-for="item in capabilityOptions" :key="item.value" :value="item.value">{{ item.label }}</el-checkbox>
        </el-checkbox-group>
      </el-form-item>
      <el-form-item label="默认">
        <el-switch v-model="form.isDefault" :active-value="1" :inactive-value="0" />
      </el-form-item>
      <el-form-item label="启用">
        <el-switch v-model="form.enabled" :active-value="1" :inactive-value="0" />
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="form.remark" type="textarea" :rows="2" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialog = false">取消</el-button>
      <el-button type="primary" :loading="saving" @click="save">保存</el-button>
    </template>
  </el-dialog>
  <PlaygroundDialog v-model="playOpen" :model="playModel" />
</template>

<style scoped>
.toolbar, .filters { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.cap { margin-right: 6px; }
</style>
