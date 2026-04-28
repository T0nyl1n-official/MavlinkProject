<template>
  <div class="map-trace tech-card">
    <div class="panel-header">
      <h3 class="panel-title gradient-title">无人机飞行轨迹</h3>
    </div>

    <div class="map-container">
      <div class="map-wrapper">
        <!-- 模拟地图背景 -->
        <div class="map-background"></div>
        
        <!-- 地图网格 -->
        <div class="map-grid"></div>
        
        <!-- SVG轨迹图 -->
        <svg class="map-svg" viewBox="0 0 800 600">
          <!-- 机巢位置 -->
          <g class="nest-position">
            <circle cx="200" cy="300" r="10" fill="#39ff14" stroke="#00aa33" stroke-width="2">
              <animate attributeName="r" values="10;12;10" dur="2s" repeatCount="indefinite" />
            </circle>
            <text x="200" y="330" text-anchor="middle" fill="#39ff14" font-size="12" font-family="var(--font-lcd)">机巢</text>
            <text x="200" y="345" text-anchor="middle" fill="var(--text-muted)" font-size="10">{{ nestCoords.lat }}, {{ nestCoords.lon }}</text>
          </g>
          
          <!-- 泄漏点位置 -->
          <g class="leak-position">
            <circle cx="600" cy="300" r="10" fill="#ff073a" stroke="#cc0033" stroke-width="2">
              <animate attributeName="r" values="10;14;10" dur="1.5s" repeatCount="indefinite" />
              <animate attributeName="opacity" values="0.8;1;0.8" dur="1.5s" repeatCount="indefinite" />
            </circle>
            <text x="600" y="330" text-anchor="middle" fill="#ff073a" font-size="12" font-family="var(--font-lcd)">泄漏点</text>
            <text x="600" y="345" text-anchor="middle" fill="var(--text-muted)" font-size="10">{{ leakCoords.lat }}, {{ leakCoords.lon }}</text>
          </g>
          
          <!-- 飞行轨迹 -->
          <g v-if="hasAlert" class="flight-path">
            <!-- 轨迹线 -->
            <path 
              d="M200,300 L300,250 L400,350 L500,280 L600,300" 
              stroke="#ff073a" 
              stroke-width="3" 
              fill="none" 
              stroke-dasharray="10,5" 
              stroke-linecap="round"
            >
              <animate attributeName="stroke-dashoffset" values="0;15" dur="2s" repeatCount="indefinite" />
            </path>
            
            <!-- 轨迹点 -->
            <circle cx="300" cy="250" r="4" fill="#ff073a" opacity="0.6" />
            <circle cx="400" cy="350" r="4" fill="#ff073a" opacity="0.6" />
            <circle cx="500" cy="280" r="4" fill="#ff073a" opacity="0.6" />
            
            <!-- 距离显示 -->
            <g class="distance-label">
              <rect x="380" y="300" width="100" height="30" fill="rgba(0,0,0,0.7)" stroke="#ff073a" stroke-width="1" rx="4" />
              <text x="430" y="320" text-anchor="middle" fill="#ff073a" font-size="12" font-family="var(--font-lcd)">{{ distance }}km</text>
            </g>
          </g>
        </svg>
        
        <!-- 地图图例 -->
        <div class="map-legend">
          <div class="legend-item">
            <div class="legend-color nest-color"></div>
            <span class="legend-text">机巢</span>
          </div>
          <div class="legend-item">
            <div class="legend-color leak-color"></div>
            <span class="legend-text">泄漏点</span>
          </div>
          <div class="legend-item" v-if="hasAlert">
            <div class="legend-line"></div>
            <span class="legend-text">飞行轨迹</span>
          </div>
        </div>
      </div>

      <!-- 地图信息 -->
      <div class="map-info">
        <div class="info-item">
          <span class="info-label">机巢坐标:</span>
          <span class="info-value lcd-display">{{ nestCoords.lat }}, {{ nestCoords.lon }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">泄漏点坐标:</span>
          <span class="info-value lcd-display">{{ leakCoords.lat }}, {{ leakCoords.lon }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">距离:</span>
          <span class="info-value lcd-display">{{ distance }}km</span>
        </div>
        <div class="info-item">
          <span class="info-label">状态:</span>
          <span class="info-value" :class="hasAlert ? 'alert' : 'normal'">
            {{ hasAlert ? '警报触发' : '正常' }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'

// 机巢位置
const nestCoords = ref({
  lat: '22.5431',
  lon: '114.0523'
})

// 泄漏点位置
const leakCoords = ref({
  lat: '22.5531',
  lon: '114.0623'
})

// 警报状态
const props = defineProps<{
  hasAlert?: boolean
  leakCoordinates?: {
    lat: string
    lon: string
  }
}>()

// 计算警报状态
const hasAlert = computed(() => props.hasAlert || false)

// 距离
const distance = ref('2.3')

// 监听泄漏点坐标变化
watch(() => props.leakCoordinates, (newCoords) => {
  if (newCoords) {
    leakCoords.value = newCoords
  }
}, { deep: true })
</script>

<style scoped>
.map-trace {
  padding: 20px;
  min-height: 500px;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 16px;
}

.panel-title {
  font-size: 16px;
  font-weight: 600;
  margin: 0;
  text-transform: uppercase;
  letter-spacing: 2px;
}

.panel-actions {
  display: flex;
  gap: 12px;
}

.map-container {
  position: relative;
}

.map-wrapper {
  position: relative;
  width: 100%;
  height: 400px;
  background: rgba(0, 0, 0, 0.8);
  border: 1px solid var(--border-color);
  overflow: hidden;
  clip-path: polygon(0 8px, 8px 0, calc(100% - 8px) 0, 100% 8px, 100% calc(100% - 8px), calc(100% - 8px) 100%, 8px 100%, 0 calc(100% - 8px));
}

.map-background {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, #0B1A2A 0%, #0D2137 100%);
  z-index: 0;
}

.map-grid {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-image:
    linear-gradient(rgba(0, 212, 255, 0.06) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0, 212, 255, 0.06) 1px, transparent 1px);
  background-size: 40px 40px;
  z-index: 1;
  pointer-events: none;
}

.map-svg {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 2;
}

/* 地图图例 */
.map-legend {
  position: absolute;
  bottom: 16px;
  right: 16px;
  background: rgba(0, 0, 0, 0.7);
  border: 1px solid var(--border-color);
  padding: 12px;
  z-index: 3;
  clip-path: polygon(0 6px, 6px 0, 100% 0, 100% calc(100% - 6px), calc(100% - 6px) 100%, 0 100%);
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  font-size: 12px;
  color: var(--text-secondary);
}

.legend-item:last-child {
  margin-bottom: 0;
}

.legend-color {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.nest-color {
  background: #39ff14;
  box-shadow: 0 0 5px #39ff14;
}

.leak-color {
  background: #ff073a;
  box-shadow: 0 0 5px #ff073a;
}

.legend-line {
  width: 20px;
  height: 2px;
  background: #ff073a;
  background-image: linear-gradient(90deg, #ff073a 50%, transparent 50%);
  background-size: 8px 2px;
}

.legend-text {
  font-size: 11px;
  font-family: var(--font-lcd);
  letter-spacing: 1px;
}

/* 地图信息 */
.map-info {
  margin-top: 16px;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 12px;
  padding-top: 16px;
  border-top: 1px solid var(--border-color);
}

.info-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  font-size: 13px;
}

.info-label {
  color: var(--text-muted);
  min-width: 100px;
  font-size: 12px;
}

.info-value {
  color: var(--text-primary);
  font-weight: 500;
  flex: 1;
  min-width: 0;
}

.map-info .lcd-display,
.map-info .lcd-number {
  font-size: clamp(1rem, 1.4vw, 1.35rem) !important;
  letter-spacing: 1px;
  line-height: 1.35;
  word-break: break-word;
}

.info-value.alert {
  color: var(--neon-red);
  text-shadow: 0 0 5px var(--neon-red);
}

.info-value.normal {
  color: var(--neon-green);
  text-shadow: 0 0 5px var(--neon-green);
}

@media (max-width: 768px) {
  .map-wrapper {
    height: 300px;
  }
  
  .map-info {
    grid-template-columns: 1fr;
  }

  .map-info .lcd-display,
  .map-info .lcd-number {
    font-size: 0.95rem !important;
    letter-spacing: 0.5px;
  }
  
  .panel-header {
    flex-direction: column;
    align-items: stretch;
  }
  
  .panel-actions {
    justify-content: center;
  }
  
  .map-legend {
    bottom: 8px;
    right: 8px;
    padding: 8px;
  }
  
  .legend-item {
    font-size: 10px;
  }
}
</style>