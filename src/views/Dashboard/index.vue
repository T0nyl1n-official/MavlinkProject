<template>
  <div class="dashboard">
    <h1 class="gradient-title">📊 仪表盘</h1>
    
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon">📋</div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.totalChains }}</div>
          <div class="stat-label">总任务链</div>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon">🚀</div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.activeChains }}</div>
          <div class="stat-label">运行中任务链</div>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon">📱</div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.totalBoards }}</div>
          <div class="stat-label">总板子数</div>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon">🟢</div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.onlineBoards }}</div>
          <div class="stat-label">在线板子</div>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon">⚠️</div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.errorCount }}</div>
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
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'

const stats = ref({
  totalChains: 0,
  activeChains: 0,
  totalBoards: 0,
  onlineBoards: 0,
  errorCount: 0
})

const loading = ref(true)

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

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.dashboard {
  padding: 24px;
  min-height: 100vh;
}

.gradient-title {
  font-size: 2rem;
  font-weight: 600;
  margin-bottom: 32px;
  text-align: center;
  letter-spacing: 2px;
  background: linear-gradient(135deg, var(--primary), var(--primary-light));
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
}

.stat-card:hover {
  transform: translateY(-3px);
  box-shadow: var(--shadow-hover);
}

.stat-icon {
  font-size: 32px;
  width: 60px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--primary-soft);
  border-radius: var(--radius-md);
  color: var(--primary);
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 2rem;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 4px;
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

