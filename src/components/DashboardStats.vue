<template>
  <div class="dashboard-stats">
    <div class="stats-grid">
      <!-- 无人机数量 -->
      <div class="stat-card tech-card">
        <div class="stat-icon drone-icon">
          🚁
        </div>
        <div class="stat-content">
          <div class="stat-label">无人机数量</div>
          <div class="stat-value">
            <CountAnimation :value="droneCount" />
          </div>
          <div class="stat-status normal">正常运行</div>
        </div>
      </div>

      <!-- 任务链数量 -->
      <div class="stat-card tech-card">
        <div class="stat-icon chain-icon">
          🔗
        </div>
        <div class="stat-content">
          <div class="stat-label">任务链数量</div>
          <div class="stat-value">
            <CountAnimation :value="chainCount" />
          </div>
          <div class="stat-status normal">运行中</div>
        </div>
      </div>

      <!-- 板子数量 -->
      <div class="stat-card tech-card">
        <div class="stat-icon board-icon">
          📟
        </div>
        <div class="stat-content">
          <div class="stat-label">板子数量</div>
          <div class="stat-value">
            <CountAnimation :value="boardCount" />
          </div>
          <div class="stat-status normal">在线</div>
        </div>
      </div>

      <!-- 错误数 -->
      <div class="stat-card tech-card">
        <div class="stat-icon error-icon">
          ⚠️
        </div>
        <div class="stat-content">
          <div class="stat-label">错误数</div>
          <div class="stat-value">
            <CountAnimation :value="errorCount" />
          </div>
          <div class="stat-status" :class="errorCount > 0 ? 'alert' : 'normal'">
            {{ errorCount > 0 ? '需要处理' : '正常' }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import CountAnimation from './CountAnimation.vue'

const droneCount = ref(3)
const chainCount = ref(2)
const boardCount = ref(5)
const errorCount = ref(1)

// 模拟数据变化
onMounted(() => {
  // 每5秒随机更新错误数
  setInterval(() => {
    errorCount.value = Math.floor(Math.random() * 5)
  }, 5000)
})
</script>

<style scoped>
.dashboard-stats {
  margin-bottom: 24px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
}

.stat-card {
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  transition: all 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-glow), var(--shadow-hover);
  border-color: var(--cyan-glow);
}

.stat-icon {
  font-size: 32px;
  width: 60px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 212, 255, 0.1);
  border: 1px solid var(--border-color);
  clip-path: polygon(0 8px, 8px 0, calc(100% - 8px) 0, 100% 8px, 100% calc(100% - 8px), calc(100% - 8px) 100%, 8px 100%, 0 calc(100% - 8px));
  flex-shrink: 0;
}

.drone-icon {
  color: var(--cyan-glow);
  box-shadow: 0 0 15px rgba(0, 212, 255, 0.3);
}

.chain-icon {
  color: var(--neon-green);
  box-shadow: 0 0 15px rgba(57, 255, 20, 0.3);
}

.board-icon {
  color: var(--warning-yellow);
  box-shadow: 0 0 15px rgba(255, 204, 0, 0.3);
}

.error-icon {
  color: var(--neon-red);
  box-shadow: 0 0 15px rgba(255, 7, 58, 0.3);
}

.stat-content {
  flex: 1;
}

.stat-label {
  font-size: 12px;
  color: var(--text-muted);
  font-family: var(--font-lcd);
  letter-spacing: 1px;
  text-transform: uppercase;
  margin-bottom: 4px;
}

.stat-value {
  font-size: 32px;
  font-weight: 700;
  margin-bottom: 4px;
}

.stat-status {
  font-size: 11px;
  font-family: var(--font-lcd);
  letter-spacing: 1px;
  text-transform: uppercase;
  padding: 2px 8px;
  display: inline-block;
  clip-path: polygon(4px 0, 100% 0, calc(100% - 4px) 100%, 0 100%);
}

.stat-status.normal {
  background: rgba(57, 255, 20, 0.1);
  color: var(--neon-green);
  border: 1px solid var(--neon-green);
}

.stat-status.alert {
  background: rgba(255, 7, 58, 0.1);
  color: var(--neon-red);
  border: 1px solid var(--neon-red);
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .stat-card {
    flex-direction: column;
    text-align: center;
    gap: 12px;
  }
  
  .stat-icon {
    width: 50px;
    height: 50px;
    font-size: 24px;
  }
  
  .stat-value {
    font-size: 24px;
  }
}
</style>