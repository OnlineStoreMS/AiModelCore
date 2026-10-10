<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { capabilityLabel, getDashboard, listUsage, type DashboardSummary, type UsageItem } from '../api/aimodel'

const loading = ref(false)
const rows = ref<UsageItem[]>([])
const summary = ref<DashboardSummary | null>(null)

function formatTime(value: string) {
  if (!value) return ''
  return value.replace('T', ' ').slice(0, 19)
}

onMounted(async () => {
  loading.value = true
  try {
    const [usage, dash] = await Promise.all([listUsage(), getDashboard()])
    rows.value = usage
    summary.value = dash
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <el-row :gutter="16">
    <el-col :span="16">
      <el-card shadow="never">
        <template #header>近期调用</template>
        <el-table v-loading="loading" :data="rows">
          <el-table-column label="时间" width="180">
            <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
          </el-table-column>
          <el-table-column prop="callerApp" label="来源应用" width="140" />
          <el-table-column label="能力" width="120">
            <template #default="{ row }">{{ capabilityLabel(row.capability) }}</template>
          </el-table-column>
          <el-table-column prop="modelName" label="模型" min-width="160" />
          <el-table-column prop="totalTokens" label="Token" width="90" />
          <el-table-column label="结果" width="90">
            <template #default="{ row }">
              <el-tag :type="row.status === 'ok' ? 'success' : 'danger'" size="small">{{ row.status === 'ok' ? '成功' : '失败' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="errorMessage" label="说明" min-width="180" show-overflow-tooltip />
        </el-table>
      </el-card>
    </el-col>
    <el-col :span="8">
      <el-card shadow="never">
        <template #header>余额</template>
        <el-statistic title="今日 Token" :value="summary?.todayTokens || 0" />
        <p class="note">{{ summary?.balanceNote }}</p>
        <el-alert title="各厂商控制台余额、账单拉取尚未接入，表结构已预留。" type="info" :closable="false" />
      </el-card>
    </el-col>
  </el-row>
</template>

<style scoped>
.note { color: #606266; line-height: 1.6; }
</style>
