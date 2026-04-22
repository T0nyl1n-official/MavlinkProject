<template>
  <div class="dashboard page-transition">
    <h1 class="gradient-title">📊 仪表盘</h1>
    
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon pulse-glow">📋</div>
        <div class="stat-content">
          <div class="stat-value">{{ displayTotalChains }}</div>
          <div class="stat-label">总任务链</div>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon pulse-glow">🚀</div>
        <div class="stat-content">
          <div class="stat-value">{{ displayActiveChains }}</div>
          <div class="stat-label">运行中任务链</div>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon pulse-glow">📱</div>
        <div class="stat-content">
          <div class="stat-value">{{ displayTotalBoards }}</div>
          <div class="stat-label">总板子数</div>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon pulse-glow">🟢</div>
        <div class="stat-content">
          <div class="stat-value">{{ displayOnlineBoards }}</div>
          <div class="stat-label">在线板子</div>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon pulse-glow">⚠️</div>
        <div class="stat-content">
          <div class="stat-value">{{ displayErrorCount }}</div>
          <div class="stat-label">错误数</div>
        </div>
      </div>
    </div>
    
    <div class="refresh-section">
      <el-button type="primary" @click="loadData" :loading="loading">
        🔄 刷新数据
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useCountAnimation } from '@/composables/useCountAnimation'

const stats = ref({
  totalChains: 0,
  activeChains: 0,
  totalBoards: 0,
  onlineBoards: 0,
  errorCount: 0
})

const loading = ref(true)

const { displayValue: displayTotalChains, setValue: setTotalChains } = useCountAnimation(0)
const { displayValue: displayActiveChains, setValue: setActiveChains } = useCountAnimation(0)
const { displayValue: displayTotalBoards, setValue: setTotalBoards } = useCountAnimation(0)
const { displayValue: displayOnlineBoards, setValue: setOnlineBoards } = useCountAnimation(0)
const { displayValue: displayErrorCount, setValue: setErrorCount } = useCountAnimation(0)

const loadData = async () => {
  loading.value = true
  try {
    // 模拟数据，实际项目中替换为真实API调用
    // const [chainsRes, boardsRes] = await Promise.all([
    //   getChainListApi(),
    //   getBoardListApi()
    // ])
    
    // 模拟数据
    await new Promise(resolve => setTimeout(resolve, 500))
    stats.value = {
      totalChains: 12,
      activeChains: 3,
      totalBoards: 8,
      onlineBoards: 6,
      errorCount: 1
    }
  } catch (error) {
    ElMessage.error('加载统计数据失败')
  } finally {
    loading.value = false
  }
}

// 监听数据变化，更新动画目标值
watch(stats, (newStats) => {
  setTotalChains(newStats.totalChains)
  setActiveChains(newStats.activeChains)
  setTotalBoards(newStats.totalBoards)
  setOnlineBoards(newStats.onlineBoards)
  setErrorCount(newStats.errorCount)
}, { deep: true })

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
}

.gradient-title {
  font-size: 2rem;
  font-weight: 600;
  margin-bottom: 32px;
  text-align: center;
  letter-spacing: 2px;
  background: linear-gradient(135deg, #fff, #6c9bd1);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 24px;
  margin-bottom: 32px;
}

.stat-card {
  background: var(--bg-card);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-md);
  padding: 24px;
  display: flex;
  align-items: center;
  gap: 16px;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
  border: 1px solid var(--border-color);
}

/* 卡片左上角发光点 */
.stat-card::before {
  content: '';
  position: absolute;
  top: 8px;
  left: 8px;
  width: 8px;
  height: 8px;
  background: linear-gradient(135deg, #2a5298, #6c9bd1);
  border-radius: 50%;
  box-shadow: 0 0 12px #2a5298;
  animation: pulse 2s ease-in-out infinite;
  z-index: 1;
}

@keyframes pulse {
  0%, 100% {
    box-shadow: 0 0 8px #2a5298;
  }
  50% {
    box-shadow: 0 0 20px #6c9bd1;
  }
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-hover);
  border-color: transparent;
}

/* 悬浮时渐变发光边框 */
.stat-card:hover::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 1px;
  background: linear-gradient(135deg, #2a5298, #1e3c72);
  border-radius: var(--radius-md);
  -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
  pointer-events: none;
  animation: borderGlow 2s ease-in-out infinite alternate;
}

@keyframes borderGlow {
  0% {
    opacity: 0.6;
  }
  100% {
    opacity: 1;
  }
}

.stat-icon {
  font-size: 32px;
  width: 60px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(42, 82, 152, 0.2);
  border-radius: var(--radius-md);
  color: #6c9bd1;
}

/* 脉冲光晕效果 */
.pulse-glow {
  animation: pulseIcon 2s ease-in-out infinite;
}

@keyframes pulseIcon {
  0%, 100% {
    box-shadow: 0 0 8px rgba(42, 82, 152, 0.6);
  }
  50% {
    box-shadow: 0 0 24px rgba(108, 155, 209, 0.8);
  }
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 2rem;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 4px;
  font-family: 'Courier New', monospace;
}

.stat-label {
  font-size: 14px;
  color: var(--text-secondary);
}

.refresh-section {
  text-align: center;
  margin-top: 32px;
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .stat-card {
    flex-direction: column;
    text-align: center;
  }
  
  .stat-icon {
    width: 50px;
    height: 50px;
    font-size: 24px;
  }
  
  .stat-value {
    font-size: 1.5rem;
  }
}
</style>

