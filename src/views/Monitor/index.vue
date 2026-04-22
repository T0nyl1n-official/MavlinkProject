<script setup lang="ts">
import { watch, onMounted, ref } from 'vue'
import { usePolling } from '@/composables/usePolling'
import { useMonitorStore } from '@/stores/monitor'
import { ElMessage, ElTable, ElTableColumn, ElTag, ElTabs, ElTabPane, ElButton } from 'element-plus'

const monitorStore = useMonitorStore()

// 轮询配置
const pollingInterval = ref(5000) // 5秒

const { data, error, loading, stop, start } = usePolling(async () => {
  await monitorStore.fetchAllData()
  return monitorStore.lastSnapshot
}, pollingInterval.value)

onMounted(() => {
  monitorStore.fetchAllData()
  start() // 自动开始轮询
})

watch(data, snapshot => {
  if (!snapshot) return
  monitorStore.setSnapshot(snapshot)
})

watch(error, msg => {
  if (!msg) return
  monitorStore.setError(msg)
  ElMessage.error(msg)
})

// 方法
function handleRefresh() {
  monitorStore.fetchAllData()
  ElMessage.info('手动刷新数据')
}

function handleStartPolling() {
  start()
  ElMessage.info('开始轮询')
}

function handleStopPolling() {
  stop()
  ElMessage.info('停止轮询')
}

function handleChangeInterval() {
  stop()
  start()
  ElMessage.info(`轮询间隔已设置为 ${pollingInterval.value}ms`)
}

function handleClearErrors() {
  monitorStore.errors = []
  ElMessage.info('错误日志已清空')
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
        <div class="monitor-content">
          <div v-if="monitorStore.chainStatus" class="card">
            <div class="card-title">任务链状态</div>
            <el-table :data="[monitorStore.chainStatus]" style="width: 100%">
              <el-table-column prop="chainId" label="任务链ID" />
              <el-table-column prop="status" label="状态">
                <template #default="{ row }">
                  <el-tag :type="row.status === 'running' ? 'success' : row.status === 'completed' ? 'info' : row.status === 'failed' ? 'danger' : 'warning'">
                    {{ row.status }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="currentNode" label="当前节点" />
              <el-table-column prop="progress" label="进度">
                <template #default="{ row }">
                  <div class="progress-bar">
                    <div class="progress-fill" :style="{ width: row.progress + '%' }"></div>
                    <span class="progress-text">{{ row.progress }}%</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column prop="startTime" label="开始时间" />
              <el-table-column prop="lastUpdate" label="最后更新" />
            </el-table>
          </div>

          <div v-if="!monitorStore.chainStatus" class="empty">
            暂无任务链数据
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="错误信息">
        <div class="monitor-content">
          <div v-if="monitorStore.errors.length > 0" class="card error">
            <div class="card-header">
              <div class="card-title">错误日志</div>
              <el-button size="small" type="danger" @click="handleClearErrors">清空日志</el-button>
            </div>
            <el-table :data="monitorStore.errors" style="width: 100%">
              <el-table-column prop="id" label="错误ID" />
              <el-table-column prop="chainId" label="任务链ID" />
              <el-table-column prop="error" label="错误信息" />
              <el-table-column prop="timestamp" label="时间" />
            </el-table>
          </div>

          <div v-if="monitorStore.errors.length === 0" class="empty">
            暂无错误信息
          </div>
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

          <div v-if="data" class="card">
            <div class="card-title">最近一次轮询快照</div>
            <pre class="pre">{{ JSON.stringify(data, null, 2) }}</pre>
          </div>

          <div v-if="!monitorStore.systemInfo && !data" class="empty">
            暂无系统数据
          </div>
        </div>
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

.card.error {
  border-color: #FF4488;
  background: #FFF5F7;
}

.card.error:hover {
  border-color: #FF4488;
  box-shadow: 0 4px 12px rgba(255, 68, 136, 0.2);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
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

.error-message {
  color: #ff4488;
  font-size: 14px;
}

.empty {
  grid-column: 1 / -1;
  color: var(--text-secondary);
  text-align: center;
  padding: 40px;
  border: 1px dashed var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-card);
}

/* 进度条样式 */
.progress-bar {
  position: relative;
  height: 20px;
  background: #E0E0E0;
  border-radius: 10px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: var(--primary-gradient);
  border-radius: 10px;
  transition: width 0.3s ease;
}

.progress-text {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-primary);
}

/* 表格样式 */
:deep(.el-table) {
  --el-table-bg-color: var(--bg-card);
  --el-table-border-color: var(--border-color);
  --el-table-header-bg-color: var(--bg-table-header);
  --el-table-header-text-color: var(--text-primary);
  --el-table-row-hover-bg-color: #F5F7FA;
  --el-table-text-color: var(--text-secondary);
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

