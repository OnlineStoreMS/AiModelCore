<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getDashboard, type DashboardSummary } from '../api/aimodel'

const router = useRouter()
const loading = ref(false)
const summary = ref<DashboardSummary | null>(null)

async function load() {
  loading.value = true
  try {
    summary.value = await getDashboard()
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div v-loading="loading">
    <el-row :gutter="16">
      <el-col :span="6">
        <el-card shadow="never">
          <div class="stat-label">接入</div>
          <div class="stat-value">{{ summary?.enabledConnections ?? 0 }}<span>/{{ summary?.connections ?? 0 }}</span></div>
          <el-button link type="primary" @click="router.push('/connections')">去配置</el-button>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <div class="stat-label">可用模型</div>
          <div class="stat-value">{{ summary?.models ?? 0 }}</div>
          <el-button link type="primary" @click="router.push('/models')">管理模型</el-button>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <div class="stat-label">今日调用</div>
          <div class="stat-value">{{ summary?.todayCalls ?? 0 }}</div>
          <div class="stat-sub">失败 {{ summary?.todayErrors ?? 0 }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <div class="stat-label">今日 Token</div>
          <div class="stat-value">{{ summary?.todayTokens ?? 0 }}</div>
          <div class="stat-sub">由本平台调用累计</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card class="block" shadow="never">
      <template #header>给其他应用调用</template>
      <p class="desc">
        商品管理中心等应用通过内网调用本服务，按能力自动选用已启用的默认模型。对话与识图走
        <code>POST /api/v1/internal/chat</code>，商品图片修改走
        <code>POST /api/v1/internal/images</code>（带原图时为编辑，不带则为生成）。
        请求头使用 <code>X-Internal-Token</code>，正文带 <code>tenantId</code>、<code>callerApp</code>。
      </p>
    </el-card>

    <el-card class="block" shadow="never">
      <template #header>余额</template>
      <el-empty :description="summary?.balanceNote || '厂商余额接口尚未对接'" :image-size="72" />
    </el-card>
  </div>
</template>

<style scoped>
.stat-label { color: #909399; font-size: 13px; }
.stat-value { margin: 8px 0; font-size: 28px; font-weight: 600; color: #303133; }
.stat-value span, .stat-sub { font-size: 13px; color: #909399; font-weight: 400; }
.block { margin-top: 16px; }
.desc { margin: 0; line-height: 1.7; color: #606266; }
code { background: #f4f4f5; padding: 0 4px; border-radius: 4px; }
</style>
