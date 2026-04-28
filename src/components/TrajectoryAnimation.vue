<template>
  <div class="trajectory-container">
    <div class="trajectory-header">
      <h3 class="trajectory-title gradient-title">无人机飞行轨迹</h3>
    </div>
    
    <div class="trajectory-content">
      <!-- 轨迹动画区域 -->
      <div class="animation-area">
        <svg class="trajectory-svg" viewBox="0 0 600 300">
          <!-- 机巢 -->
          <g class="nest">
            <circle cx="100" cy="150" r="12" fill="#39ff14" stroke="#00aa33" stroke-width="2">
              <animate attributeName="r" values="12;14;12" dur="2s" repeatCount="indefinite" />
            </circle>
            <text x="100" y="180" text-anchor="middle" fill="#39ff14" font-size="12" font-family="var(--font-lcd)">机巢</text>
            <text x="100" y="195" text-anchor="middle" fill="var(--text-muted)" font-size="10">22.5431, 114.0523</text>
          </g>
          
          <!-- 泄漏点 -->
          <g class="leak">
            <circle cx="500" cy="150" r="12" fill="#ff073a" stroke="#cc0033" stroke-width="2">
              <animate attributeName="r" values="12;16;12" dur="1.5s" repeatCount="indefinite" />
              <animate attributeName="opacity" values="0.8;1;0.8" dur="1.5s" repeatCount="indefinite" />
            </circle>
            <text x="500" y="180" text-anchor="middle" fill="#ff073a" font-size="12" font-family="var(--font-lcd)">泄漏点</text>
            <text x="500" y="195" text-anchor="middle" fill="var(--text-muted)" font-size="10">22.5531, 114.0623</text>
          </g>
          
          <!-- 飞行轨迹 -->
          <g v-if="showTrajectory" class="trajectory">
            <!-- 轨迹路径 -->
            <path 
              d="M100,150 Q200,100 300,180 T500,150" 
              stroke="#ff073a" 
              stroke-width="3" 
              fill="none" 
              stroke-dasharray="10,5" 
              stroke-linecap="round"
              :style="trajectoryStyle"
            >
              <animate 
                v-if="animate" 
                attributeName="stroke-dashoffset" 
                from="0" 
                to="15" 
                dur="1.5s" 
                repeatCount="indefinite" 
              />
            </path>
            
            <!-- 轨迹点 -->
            <circle cx="200" cy="100" r="5" fill="#ff073a" opacity="0.6" v-if="showTrajectory" />
            <circle cx="300" cy="180" r="5" fill="#ff073a" opacity="0.6" v-if="showTrajectory" />
            <circle cx="400" cy="120" r="5" fill="#ff073a" opacity="0.6" v-if="showTrajectory" />
            
            <!-- 距离标签 -->
            <g class="distance-label" v-if="showTrajectory">
              <rect x="280" y="160" width="120" height="30" fill="rgba(0,0,0,0.7)" stroke="#ff073a" stroke-width="1" rx="4" />
              <text x="340" y="180" text-anchor="middle" fill="#ff073a" font-size="12" font-family="var(--font-lcd)">2.3km</text>
            </g>
          </g>
        </svg>
      </div>
      
      <!-- 控制按钮 -->
      <div class="control-buttons">
        <el-button type="danger" @click="triggerAlert" :disabled="isAnimating">
          <span class="alert-icon">⚠️</span> 模拟警报
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'

const emit = defineEmits<{
  'alert-triggered': [message: string]
}>()

const showTrajectory = ref(false)
const animate = ref(false)
const isAnimating = ref(false)

const trajectoryStyle = computed(() => {
  return {
    opacity: showTrajectory.value ? '1' : '0',
    transition: 'opacity 0.5s ease'
  }
})

const triggerAlert = () => {
  if (isAnimating.value) return
  
  isAnimating.value = true
  
  // 显示Alert提示条
  ElMessage.warning({
    message: '⚠️ H₂S 泄漏警报',
    duration: 3000
  })
  
  // 显示轨迹
  showTrajectory.value = true
  
  // 开始动画
  setTimeout(() => {
    animate.value = true
  }, 200)
  
  // 发送警报事件
  emit('alert-triggered', 'H₂S 泄漏警报 | 坐标 22.5531, 114.0623')
  
  // 5秒后停止动画
  setTimeout(() => {
    animate.value = false
    showTrajectory.value = false
    isAnimating.value = false
  }, 5000)
}

onMounted(() => {
  // 初始化
})

// 暴露方法
defineExpose({
  triggerAlert
})
</script>

<style scoped>
.trajectory-container {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  padding: 24px;
  min-height: 400px;
  clip-path: polygon(0 10px, 10px 0, 100% 0, 100% calc(100% - 10px), calc(100% - 10px) 100%, 0 100%);
  position: relative;
  overflow: visible;
}

.trajectory-container::after {
  content: '';
  position: absolute;
  top: 6px;
  left: 6px;
  width: 20px;
  height: 20px;
  border-left: 2px solid var(--cyan-glow);
  border-top: 2px solid var(--cyan-glow);
  box-shadow: 0 0 5px var(--cyan-glow);
  clip-path: polygon(0 0, 100% 0, 0 100%);
}

.trajectory-header {
  margin-bottom: 24px;
}

.trajectory-title {
  font-size: 18px;
  font-weight: 600;
  margin: 0;
  text-transform: uppercase;
  letter-spacing: 2px;
}

.trajectory-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.animation-area {
  background: rgba(11, 26, 42, 0.9);
  border: 1px solid var(--border-color);
  padding: 20px;
  clip-path: polygon(0 8px, 8px 0, 100% 0, 100% calc(100% - 8px), calc(100% - 8px) 100%, 0 100%);
}

.trajectory-svg {
  width: 100%;
  height: 300px;
}

.control-buttons {
  display: flex;
  justify-content: center;
  gap: 12px;
  margin-top: 16px;
}

.alert-icon {
  margin-right: 6px;
}

@media (max-width: 768px) {
  .trajectory-container {
    padding: 16px;
  }
  
  .trajectory-svg {
    height: 200px;
  }
}
</style>