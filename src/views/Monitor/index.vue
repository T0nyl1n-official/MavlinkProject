<script setup lang="ts">
import { watch, onMounted, ref } from 'vue'
import { usePolling } from '@/composables/usePolling'
import { useMonitorStore } from '@/stores/monitor'
import { ElMessage } from 'element-plus'
import FieldSnapshot from '@/components/FieldSnapshot.vue'
import TrajectoryAnimation from '@/components/TrajectoryAnimation.vue'
import VideoStream from '@/components/VideoStream.vue'

const monitorStore = useMonitorStore()
const pollingInterval = ref(5000)
const trajectoryAnimationRef = ref<any>(null)

const { data: latestSnapshot, error: pollingError, loading, stop, start } = usePolling(async () => {
  await monitorStore.fetchAllData()
  return monitorStore.lastSnapshot
}, pollingInterval.value)

const statusTagMap: Record<string, 'success' | 'primary' | 'danger' | 'info' | 'warning'> = {
  running: 'success',
  completed: 'info',
  failed: 'danger',
  pending: 'warning',
  stopped: 'warning'
}

function getStatusType(status: string): 'success' | 'primary' | 'danger' | 'info' | 'warning' {
  return statusTagMap[status] || 'info'
}

function pushToast(message: string, type: 'info' | 'warning' | 'error' = 'info') {
  if (type === 'error') {
    ElMessage.error(message)
    return
  }

  if (type === 'warning') {
    ElMessage.warning(message)
    return
  }

  ElMessage.info(message)
}

onMounted(() => {
  monitorStore.fetchAllData()
  start()
})

watch(latestSnapshot, snapshot => {
  if (!snapshot) return
  monitorStore.setSnapshot(snapshot)
})

watch(pollingError, message => {
  if (!message) return
  monitorStore.setError(message)
  pushToast(message, 'error')
})

function handleRefresh() {
  monitorStore.fetchAllData()
  pushToast('已手动刷新')
}

function handleStartPolling() {
  start()
  pushToast('开始轮询')
}

function handleStopPolling() {
  stop()
  pushToast('已暂停轮询')
}

function handleChangeInterval() {
  stop()
  start()
  pushToast(`轮询间隔已调整为 ${pollingInterval.value}ms`)
}

function handleClearErrors() {
  monitorStore.errors = []
  pushToast('错误日志已清空')
}

function handleAlertTriggered(event: { coordinates: { lat: string; lon: string }; message: string } | string) {
  const alertMessage = typeof event === 'string' ? event : event.message
  const alertLog = {
    id: `alert-${Date.now()}`,
    chainId: 'CHAIN-001',
    error: alertMessage,
    timestamp: new Date().toISOString()
  }

  monitorStore.errors = [alertLog, ...monitorStore.errors]
  pushToast(alertMessage, 'warning')
}

function handleTriggerTrajectory() {
  trajectoryAnimationRef.value?.triggerAlert()
}
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h2 class="title">实时监控</h2>
      <div class="actions">
        <div class="status">
          轮询：<span class="accent">{{ loading ? '进行中...' : '空闲' }}</span>
        </div>
        <el-input-number v-model="pollingInterval" :min="1000" :max="10000" :step="500" size="small" style="width: 120px; margin-left: 10px;"></el-input-number>
        <el-button size="small" @click="handleChangeInterval" style="margin-left: 10px;">设置</el-button>
        <el-button size="small" @click="handleRefresh" style="margin-left: 10px;">刷新</el-button>
        <el-button size="small" @click="handleStartPolling" v-if="!loading" style="margin-left: 10px;">开始轮询</el-button>
        <el-button size="small" @click="handleStopPolling" v-else style="margin-left: 10px;">停止轮询</el-button>
      </div>
    </div>

    <el-tabs>
      <el-tab-pane label="无人机状态">
        <div class="monitor-content">
          <div v-if="monitorStore.droneStatus" class="card">
            <div class="card-title">基本状态</div>
            <div class="status-grid">
              <div class="status-item">
                <span class="label">武装状态</span>
                <span class="value" :class="{ active: monitorStore.droneStatus.armed }">
                  {{ monitorStore.droneStatus.armed ? '已武装' : '未武装' }}
                </span>
              </div>
              <div class="status-item">
                <span class="label">飞行模式</span>
                <span class="value">{{ monitorStore.droneStatus.mode }}</span>
              </div>
              <div class="status-item">
                <span class="label">电池电量</span>
                <span class="value">{{ monitorStore.droneStatus.battery }}%</span>
              </div>
              <div class="status-item">
                <span class="label">高度</span>
                <span class="value">{{ monitorStore.droneStatus.altitude }}m</span>
              </div>
              <div class="status-item">
                <span class="label">速度</span>
                <span class="value">{{ monitorStore.droneStatus.speed }}m/s</span>
              </div>
            </div>
          </div>

          <div v-if="monitorStore.dronePosition" class="card">
            <div class="card-title">位置信息</div>
            <div class="position-grid">
              <div class="position-item">
                <span class="label">纬度</span>
                <span class="value">{{ monitorStore.dronePosition.latitude }}</span>
              </div>
              <div class="position-item">
                <span class="label">经度</span>
                <span class="value">{{ monitorStore.dronePosition.longitude }}</span>
              </div>
              <div class="position-item">
                <span class="label">高度</span>
                <span class="value">{{ monitorStore.dronePosition.altitude }}m</span>
              </div>
              <div class="position-item">
                <span class="label">航向</span>
                <span class="value">{{ monitorStore.dronePosition.heading }}°</span>
              </div>
              <div class="position-item">
                <span class="label">速度</span>
                <span class="value">{{ monitorStore.dronePosition.speed }}m/s</span>
              </div>
            </div>
          </div>

          <div v-if="!monitorStore.droneStatus && !monitorStore.dronePosition" class="empty">
            暂无无人机数据
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="任务链状态">
        <div class="table-section">
          <el-table v-if="monitorStore.chainStatus" :data="[monitorStore.chainStatus]" class="monitor-table">
            <el-table-column prop="chainId" label="任务链ID" />
            <el-table-column prop="status" label="状态">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="currentNode" label="当前节点" />
            <el-table-column prop="progress" label="进度">
              <template #default="{ row }">
                <el-progress :percentage="row.progress" :stroke-width="8" />
              </template>
            </el-table-column>
            <el-table-column prop="startTime" label="开始时间" />
            <el-table-column prop="lastUpdate" label="最后更新" />
          </el-table>
          <div v-else class="empty">暂无任务链数据</div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="错误信息">
        <div class="table-section">
          <div v-if="monitorStore.errors.length > 0" class="error-header">
            <span>错误日志 ({{ monitorStore.errors.length }})</span>
            <el-button size="small" type="danger" @click="handleClearErrors">清空日志</el-button>
          </div>
          <el-table v-if="monitorStore.errors.length > 0" :data="monitorStore.errors" class="monitor-table">
            <el-table-column prop="id" label="错误ID" />
            <el-table-column prop="chainId" label="任务链ID" />
            <el-table-column prop="error" label="错误信息" />
            <el-table-column prop="timestamp" label="时间" />
          </el-table>
          <div v-else class="empty">暂无错误信息</div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="系统信息">
        <div class="monitor-content">
          <div v-if="monitorStore.systemInfo" class="card">
            <div class="card-title">系统信息</div>
            <div class="system-grid">
              <div class="system-item">
                <span class="label">CPU使用率</span>
                <span class="value">{{ monitorStore.systemInfo.cpuUsage }}%</span>
              </div>
              <div class="system-item">
                <span class="label">内存使用率</span>
                <span class="value">{{ monitorStore.systemInfo.memoryUsage }}%</span>
              </div>
              <div class="system-item">
                <span class="label">磁盘使用率</span>
                <span class="value">{{ monitorStore.systemInfo.diskUsage }}%</span>
              </div>
              <div class="system-item">
                <span class="label">网络状态</span>
                <span class="value">{{ monitorStore.systemInfo.networkStatus }}</span>
              </div>
              <div class="system-item">
                <span class="label">系统时间</span>
                <span class="value">{{ monitorStore.systemInfo.systemTime }}</span>
              </div>
            </div>
          </div>

          <div v-if="latestSnapshot" class="card">
            <div class="card-title">最近一次轮询快照</div>
            <pre class="pre">{{ JSON.stringify(latestSnapshot, null, 2) }}</pre>
          </div>

          <div v-if="!monitorStore.systemInfo && !latestSnapshot" class="empty">
            暂无系统数据
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="现场快照">
        <FieldSnapshot @alert-triggered="handleAlertTriggered" @trigger-trajectory="handleTriggerTrajectory" />
      </el-tab-pane>

      <el-tab-pane label="实时视频">
        <div class="monitor-content">
          <div class="card">
            <div class="card-title">视频流</div>
            <VideoStream mode="mjpeg" task-code="TASK_001" />
          </div>
        </div>
      </el-tab-pane>
      
      <el-tab-pane label="飞行轨迹">
        <TrajectoryAnimation ref="trajectoryAnimationRef" @alert-triggered="handleAlertTriggered" />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<style scoped>
.page {
  color: var(--text-primary);
  padding: 24px;
  min-height: 100vh;
  background: var(--bg-body);
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
  flex-wrap: wrap;
  gap: 10px;
}

.title {
  font-size: 18px;
  margin: 0;
  font-weight: 700;
  color: var(--text-primary);
}

.actions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.status {
  color: var(--text-secondary);
}

.accent {
  color: var(--primary);
  font-weight: 700;
}

.monitor-content {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 16px;
  margin-top: 16px;
}

.card {
  border: 1px solid var(--border-color);
  background: var(--bg-card);
  border-radius: var(--radius-md);
  padding: 16px;
  transition: all 0.3s;
}

.card:hover {
  border-color: var(--primary);
  box-shadow: 0 4px 12px rgba(30, 136, 229, 0.1);
}

.card-title {
  color: var(--text-primary);
  font-weight: 600;
  font-size: 14px;
}

.status-grid,
.position-grid,
.system-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 12px;
}

.status-item,
.position-item,
.system-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.label {
  font-size: 12px;
  color: var(--text-secondary);
}

.value {
  font-size: 14px;
  color: var(--text-primary);
  font-weight: 500;
}

.value.active {
  color: var(--primary);
}

.pre {
  margin: 0;
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
  color: var(--text-primary);
  max-height: 200px;
  overflow-y: auto;
  padding: 12px;
  background: var(--bg-body);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
}

.empty {
  color: var(--text-secondary);
  text-align: center;
  padding: 40px;
  border: 1px dashed var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-card);
}

.table-section {
  margin-top: 16px;
}

.monitor-table {
  border-radius: var(--radius-md);
  overflow: hidden;
}

.monitor-table :deep(.el-table__header) {
  th {
    background: var(--bg-table-header) !important;
    color: var(--text-primary);
    font-weight: 600;
    font-size: 13px;
    padding: 12px 0;
    height: 48px;
  }
}

.monitor-table :deep(.el-table__body) {
  td {
    padding: 12px 0;
    height: 48px;
    font-size: 13px;
    color: var(--text-secondary);
  }
  
  tr:hover > td {
    background: var(--bg-table-header) !important;
  }
}

.monitor-table :deep(.el-table__row) {
  height: 48px;
}

.error-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding: 8px 12px;
  background: rgba(255, 7, 58, 0.1);
  border: 1px solid rgba(255, 7, 58, 0.3);
  border-radius: var(--radius-sm);
  color: var(--neon-red);
  font-size: 13px;
  font-weight: 500;
}

/* 标签页样式 */
:deep(.el-tabs__header) {
  margin-bottom: 16px;
}

:deep(.el-tabs__tab) {
  color: var(--text-secondary);
}

:deep(.el-tabs__tab.is-active) {
  color: var(--primary);
}

:deep(.el-tabs__active-bar) {
  background-color: var(--primary);
}

/* 输入框样式 */
:deep(.el-input-number) {
  --el-input-bg-color: var(--bg-body);
  --el-input-border-color: var(--border-color);
  --el-input-text-color: var(--text-primary);
}

:deep(.el-input-number__decrease),
:deep(.el-input-number__increase) {
  --el-input-button-bg-color: var(--bg-card);
  --el-input-button-hover-bg-color: #F5F7FA;
  --el-input-button-border-color: var(--border-color);
  --el-input-button-text-color: var(--text-primary);
}


</style>