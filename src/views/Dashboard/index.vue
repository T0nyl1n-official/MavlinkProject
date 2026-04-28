<template>
  <div class="dashboard page-transition">
    <h1 class="gradient-title">📊 仪表盘</h1>
    
    <!-- 统计卡片 -->
    <DashboardStats />
    
    <!-- 图表区域 -->
    <div class="charts-grid">
      <TechChart title="无人机活动趋势" />
    </div>
    
    <!-- 快照和地图区域 -->
    <div class="main-content">
      <div class="left-panel">
        <SnapshotPanel 
          @update:alert="handleAlertUpdate"
          @upload-success="handleUploadSuccess"
        />
      </div>
      <div class="right-panel">
        <MapTrace :hasAlert="hasAlert" :leakCoordinates="leakCoordinates" />
      </div>
    </div>
    
    <div class="refresh-section">
      <el-button type="primary" @click="loadData" :loading="loading">
        🔄 刷新数据
      </el-button>
      <el-button type="info" @click="terminalRef?.open()">
        💻 打开终端
      </el-button>
    </div>
    
    <!-- 全局终端 -->
    <GlobalTerminal ref="terminalRef" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import DashboardStats from '@/components/DashboardStats.vue'
import SnapshotPanel from '@/components/SnapshotPanel.vue'
import MapTrace from '@/components/MapTrace.vue'
import TechChart from '@/components/TechChart.vue'
import GlobalTerminal from '@/components/GlobalTerminal.vue'

const loading = ref(true)
const terminalRef = ref<any>(null)
const hasAlert = ref(false)
const leakCoordinates = ref({
  lat: '22.5531',
  lon: '114.0623'
})

const loadData = async () => {
  loading.value = true
  try {
    // 模拟数据加载，实际项目中替换为真实API调用
    await new Promise(resolve => setTimeout(resolve, 500))
    ElMessage.success('数据刷新成功')
  } catch (error) {
    ElMessage.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

const handleAlertUpdate = (alertData: {
  hasAlert: boolean
  coordinates?: {
    lat: string
    lon: string
  }
}) => {
  hasAlert.value = alertData.hasAlert
  if (alertData.coordinates) {
    leakCoordinates.value = alertData.coordinates
  }
}

const handleUploadSuccess = (_uploadData: {
  imageUrl: string
  timestamp: string
}) => {
  // 上传成功时触发警报状态，显示地图连线
  hasAlert.value = true
  // 使用固定的泄漏点坐标
  leakCoordinates.value = {
    lat: '22.5531',
    lon: '114.0623'
  }
  
  // 在终端输出警报日志
  if (terminalRef.value) {
    terminalRef.value.addLog(`[ALERT] H2S detected at ${leakCoordinates.value.lat}, ${leakCoordinates.value.lon} - ${new Date().toISOString()}`)
  }
  
  // 3秒后关闭警报状态
  setTimeout(() => {
    hasAlert.value = false
  }, 3000)
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.dashboard {
  padding: 24px;
  min-height: 100vh;
  position: relative;
  z-index: 1;
  max-width: 1400px;
  margin: 0 auto;
}

.gradient-title {
  font-size: 2rem;
  font-weight: 600;
  margin-bottom: 32px;
  text-align: center;
  letter-spacing: 2px;
  background: linear-gradient(135deg, #fff, #00d4ff);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* 图表区域 */
.charts-grid {
  margin-bottom: 32px;
}

/* 主要内容区域 */
.main-content {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
  margin-bottom: 32px;
}

.left-panel,
.right-panel {
  flex: 1;
}

.refresh-section {
  text-align: center;
  margin-top: 32px;
}

@media (max-width: 1200px) {
  .main-content {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .dashboard {
    padding: 16px;
  }
  
  .gradient-title {
    font-size: 1.5rem;
    margin-bottom: 24px;
  }
}
</style>

