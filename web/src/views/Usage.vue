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

function formatBytes(value?: number) {
  const n = value || 0
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
  return `${(n / 1024 / 1024).toFixed(2)} MB`
}

function formatLatency(value?: number) {
  const n = value || 0
  if (n < 1000) return `${n} ms`
  return `${(n / 1000).toFixed(1)} s`
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
  <el-row :gutter="16" class="stats">
    <el-col :span="6">
      <el-card shadow="never">
        <div class="stat-label">今日发送</div>
        <div class="stat-value">{{ formatBytes(summary?.todayRequestBytes) }}</div>
      </el-card>
    </el-col>
    <el-col :span="6">
      <el-card shadow="never">
        <div class="stat-label">今日接收</div>
        <div class="stat-value">{{ formatBytes(summary?.todayResponseBytes) }}</div>
      </el-card>
    </el-col>
    <el-col :span="6">
      <el-card shadow="never">
        <div class="stat-label">今日平均耗时</div>
        <div class="stat-value">{{ formatLatency(summary?.todayAvgLatencyMs) }}</div>
        <div class="stat-sub">最长 {{ formatLatency(summary?.todayMaxLatencyMs) }}</div>
      </el-card>
    </el-col>
    <el-col :span="6">
      <el-card shadow="never">
        <div class="stat-label">今日 Token</div>
        <div class="stat-value">{{ summary?.todayTokens || 0 }}</div>
        <div class="stat-sub">调用 {{ summary?.todayCalls || 0 }} 次</div>
      </el-card>
    </el-col>
  </el-row>

  <el-row :gutter="16">
    <el-col :span="16">
      <el-card shadow="never">
        <template #header>近期调用</template>
        <el-table v-loading="loading" :data="rows">
          <el-table-column label="时间" width="170">
            <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
          </el-table-column>
          <el-table-column prop="callerApp" label="来源应用" width="120" />
          <el-table-column label="能力" width="110">
            <template #default="{ row }">{{ capabilityLabel(row.capability) }}</template>
          </el-table-column>
          <el-table-column prop="modelName" label="模型" min-width="140" />
          <el-table-column label="发送" width="100">
            <template #default="{ row }">{{ formatBytes(row.requestBytes) }}</template>
          </el-table-column>
          <el-table-column label="接收" width="100">
            <template #default="{ row }">{{ formatBytes(row.responseBytes) }}</template>
          </el-table-column>
          <el-table-column label="耗时" width="90">
            <template #default="{ row }">{{ formatLatency(row.latencyMs) }}</template>
          </el-table-column>
          <el-table-column prop="totalTokens" label="Token" width="80" />
          <el-table-column label="结果" width="80">
            <template #default="{ row }">
              <el-tag :type="row.status === 'ok' ? 'success' : 'danger'" size="small">{{ row.status === 'ok' ? '成功' : '失败' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="errorMessage" label="说明" min-width="140" show-overflow-tooltip />
        </el-table>
      </el-card>
    </el-col>
    <el-col :span="8">
      <el-card shadow="never">
        <template #header>余额</template>
        <p class="note">{{ summary?.balanceNote }}</p>
        <el-alert title="各厂商控制台余额、账单拉取尚未接入，表结构已预留。" type="info" :closable="false" />
      </el-card>
    </el-col>
  </el-row>
</template>

<style scoped>
.stats { margin-bottom: 16px; }
.stat-label { color: #909399; font-size: 13px; }
.stat-value { margin-top: 8px; font-size: 24px; font-weight: 600; color: #303133; }
.stat-sub { margin-top: 6px; font-size: 13px; color: #909399; }
.note { color: #606266; line-height: 1.6; }
</style>
