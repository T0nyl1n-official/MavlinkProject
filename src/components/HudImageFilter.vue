<template>
  <div class="hud-image-container">
    <!-- 原始图片 -->
    <img :src="imageSrc" :alt="alt" class="hud-image" />

    <!-- HUD绿色滤镜层 -->
    <div class="hud-filter-layer"></div>

    <!-- 十字瞄准线 -->
    <div class="crosshair" v-if="showCrosshair">
      <div class="crosshair-h"></div>
      <div class="crosshair-v"></div>
      <div class="crosshair-circle"></div>
    </div>

    <!-- 气体泄漏警告框 -->
    <div
      v-if="gasLeakPosition"
      class="gas-leak-warning"
      :style="gasLeakStyle"
    >
      <div class="gas-leak-label">
        <span class="warning-icon">⚠</span>
        WARNING: {{ gasType }} LEVEL CRITICAL
      </div>
      <div class="gas-leak-line"></div>
    </div>

    <!-- HUD数据叠加层 -->
    <div class="hud-data-overlay">
      <!-- 时间戳 -->
      <div class="hud-timestamp">
        <div class="timestamp-label">UTC TIME</div>
        <div class="timestamp-value">{{ currentTime }}</div>
      </div>

      <!-- 坐标 -->
      <div class="hud-coords">
        <div class="coord-item">
          <span class="coord-label">LAT</span>
          <span class="coord-value">{{ coordinates.lat }}</span>
        </div>
        <div class="coord-item">
          <span class="coord-label">LON</span>
          <span class="coord-value">{{ coordinates.lon }}</span>
        </div>
        <div class="coord-item">
          <span class="coord-label">ALT</span>
          <span class="coord-value">{{ coordinates.alt }}m</span>
        </div>
      </div>

      <!-- 雷达扫描波纹 -->
      <div class="radar-sweep" v-if="showRadar"></div>

      <!-- 状态信息 -->
      <div class="hud-status">
        <div class="status-item">
          <span class="status-indicator online"></span>
          LIVE FEED
        </div>
      </div>
    </div>

    <!-- 边框装饰 -->
    <div class="hud-border">
      <div class="corner-tl"></div>
      <div class="corner-tr"></div>
      <div class="corner-bl"></div>
      <div class="corner-br"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'

interface Props {
  imageSrc: string
  alt?: string
  showCrosshair?: boolean
  showRadar?: boolean
  gasLeakPosition?: { x: number; y: number; width: number; height: number }
  gasType?: string
  coordinates?: { lat: string; lon: string; alt: string }
}

const props = withDefaults(defineProps<Props>(), {
  alt: 'HUD Image',
  showCrosshair: true,
  showRadar: true,
  gasType: 'H2S',
  coordinates: () => ({
    lat: '31.2304',
    lon: '121.4737',
    alt: '120'
  })
})

const currentTime = ref('')
let timeInterval: number | undefined

const updateTime = () => {
  const now = new Date()
  currentTime.value = now.toISOString().split('T')[1].split('.')[0] + 'Z'
}

const gasLeakStyle = computed(() => {
  if (!props.gasLeakPosition) return {}
  return {
    left: `${props.gasLeakPosition.x}px`,
    top: `${props.gasLeakPosition.y}px`,
    width: `${props.gasLeakPosition.width}px`,
    height: `${props.gasLeakPosition.height}px`
  }
})

onMounted(() => {
  updateTime()
  timeInterval = window.setInterval(updateTime, 1000)
})

onUnmounted(() => {
  if (timeInterval) {
    clearInterval(timeInterval)
  }
})
</script>

<style scoped>
.hud-image-container {
  position: relative;
  width: 100%;
  max-width: 800px;
  aspect-ratio: 16 / 9;
  overflow: hidden;
  background: #000;
  border: 2px solid rgba(0, 255, 255, 0.5);
  box-shadow: 0 0 20px rgba(0, 255, 255, 0.3), inset 0 0 30px rgba(0, 0, 0, 0.8);
}

.hud-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

/* HUD绿色滤镜层 */
.hud-filter-layer {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background:
    rgba(0, 255, 0, 0.05)
    repeating-linear-gradient(
      0deg,
      transparent,
      transparent 2px,
      rgba(0, 255, 0, 0.03) 2px,
      rgba(0, 255, 0, 0.03) 4px
    );
  pointer-events: none;
  mix-blend-mode: overlay;
}

/* 十字瞄准线 */
.crosshair {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 120px;
  height: 120px;
  pointer-events: none;
}

.crosshair-h,
.crosshair-v {
  position: absolute;
  background: rgba(57, 255, 20, 0.8);
  box-shadow: 0 0 5px rgba(57, 255, 20, 0.8);
}

.crosshair-h {
  top: 50%;
  left: 0;
  right: 0;
  height: 1px;
  transform: translateY(-50%);
}

.crosshair-v {
  left: 50%;
  top: 0;
  bottom: 0;
  width: 1px;
  transform: translateX(-50%);
}

.crosshair-circle {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 60px;
  height: 60px;
  border: 1px solid rgba(57, 255, 20, 0.6);
  border-radius: 50%;
  box-shadow: 0 0 10px rgba(57, 255, 20, 0.3);
}

.crosshair-circle::before,
.crosshair-circle::after {
  content: '';
  position: absolute;
  background: rgba(57, 255, 20, 0.6);
}

.crosshair-circle::before {
  top: 50%;
  left: -10px;
  right: -10px;
  height: 1px;
  transform: translateY(-50%);
}

.crosshair-circle::after {
  left: 50%;
  top: -10px;
  bottom: -10px;
  width: 1px;
  transform: translateX(-50%);
}

/* 气体泄漏警告框 */
.gas-leak-warning {
  position: absolute;
  border: 2px solid #ff073a;
  background: rgba(255, 7, 58, 0.15);
  animation: gasLeakPulse 1.5s ease-in-out infinite;
  box-shadow: 0 0 20px rgba(255, 7, 58, 0.6), inset 0 0 15px rgba(255, 7, 58, 0.2);
}

@keyframes gasLeakPulse {
  0%, 100% {
    border-color: #ff073a;
    box-shadow: 0 0 20px rgba(255, 7, 58, 0.6);
  }
  50% {
    border-color: #ff4466;
    box-shadow: 0 0 35px rgba(255, 7, 58, 0.9);
  }
}

.gas-leak-label {
  position: absolute;
  top: -30px;
  left: 50%;
  transform: translateX(-50%);
  white-space: nowrap;
  background: rgba(0, 0, 0, 0.9);
  border: 1px solid #ff073a;
  padding: 4px 12px;
  font-family: var(--font-lcd);
  font-size: 10px;
  letter-spacing: 1px;
  color: #ff073a;
  text-shadow: 0 0 5px #ff073a;
  display: flex;
  align-items: center;
  gap: 6px;
}

.warning-icon {
  font-size: 12px;
  animation: blink 0.5s step-end infinite;
}

@keyframes blink {
  50% { opacity: 0; }
}

.gas-leak-line {
  position: absolute;
  top: -30px;
  left: 50%;
  width: 1px;
  height: 20px;
  background: #ff073a;
  box-shadow: 0 0 5px #ff073a;
}

/* HUD数据叠加层 */
.hud-data-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
  font-family: var(--font-lcd);
  font-size: 11px;
  color: #39ff14;
  text-shadow: 0 0 5px #39ff14;
}

.hud-timestamp {
  position: absolute;
  top: 15px;
  right: 15px;
  text-align: right;
  background: rgba(0, 0, 0, 0.6);
  padding: 8px 12px;
  border: 1px solid rgba(57, 255, 20, 0.3);
}

.timestamp-label {
  font-size: 9px;
  letter-spacing: 2px;
  margin-bottom: 4px;
  opacity: 0.7;
}

.timestamp-value {
  font-size: 14px;
  letter-spacing: 1px;
}

.hud-coords {
  position: absolute;
  bottom: 15px;
  left: 15px;
  background: rgba(0, 0, 0, 0.6);
  padding: 10px 14px;
  border: 1px solid rgba(57, 255, 20, 0.3);
  line-height: 1.8;
}

.coord-item {
  display: flex;
  gap: 8px;
}

.coord-label {
  font-size: 9px;
  letter-spacing: 1px;
  opacity: 0.7;
  width: 35px;
}

.coord-value {
  font-size: 12px;
  letter-spacing: 1px;
}

.hud-status {
  position: absolute;
  top: 15px;
  left: 15px;
  background: rgba(0, 0, 0, 0.6);
  padding: 6px 10px;
  border: 1px solid rgba(57, 255, 20, 0.3);
  font-size: 10px;
  letter-spacing: 1px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  animation: statusBlink 1s ease-in-out infinite;
}

.status-indicator.online {
  background: #39ff14;
  box-shadow: 0 0 10px #39ff14;
}

@keyframes statusBlink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

/* 雷达扫描波纹 */
.radar-sweep {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 300px;
  height: 300px;
  transform: translate(-50%, -50%);
  background: conic-gradient(
    from 0deg,
    transparent 0deg,
    rgba(57, 255, 20, 0.15) 20deg,
    transparent 40deg
  );
  animation: radarSweep 3s linear infinite;
  pointer-events: none;
}

@keyframes radarSweep {
  from { transform: translate(-50%, -50%) rotate(0deg); }
  to { transform: translate(-50%, -50%) rotate(360deg); }
}

/* 边框装饰 */
.hud-border {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
}

.corner-tl,
.corner-tr,
.corner-bl,
.corner-br {
  position: absolute;
  width: 40px;
  height: 40px;
  border-color: rgba(0, 255, 255, 0.6);
  border-style: solid;
  border-width: 0;
}

.corner-tl {
  top: 10px;
  left: 10px;
  border-top-width: 2px;
  border-left-width: 2px;
}

.corner-tr {
  top: 10px;
  right: 10px;
  border-top-width: 2px;
  border-right-width: 2px;
}

.corner-bl {
  bottom: 10px;
  left: 10px;
  border-bottom-width: 2px;
  border-left-width: 2px;
}

.corner-br {
  bottom: 10px;
  right: 10px;
  border-bottom-width: 2px;
  border-right-width: 2px;
}
</style>
