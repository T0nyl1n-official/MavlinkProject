<template>
  <div class="globe-container">
    <div class="globe">
      <div class="globe-surface">
        <div class="globe-grid"></div>
        <div class="globe-glow"></div>
      </div>
      <div class="globe-ring ring-1"></div>
      <div class="globe-ring ring-2"></div>
      <div class="globe-ring ring-3"></div>
      <!-- 位置标记点 -->
      <div
        v-for="(marker, index) in markers"
        :key="index"
        class="marker"
        :style="getMarkerStyle(marker)"
      >
        <div class="marker-dot"></div>
        <div class="marker-pulse"></div>
      </div>
    </div>
    <div class="globe-info">
      <div class="info-label">TRACKING SATELLITE</div>
      <div class="info-value">{{ satelliteName }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Marker {
  lat: number
  lon: number
  label?: string
}

interface Props {
  size?: number
  markers?: Marker[]
  satelliteName?: string
}

const props = withDefaults(defineProps<Props>(), {
  size: 300,
  markers: () => [],
  satelliteName: 'UAV-01'
})

const getMarkerStyle = (marker: Marker) => {
  // 将经纬度转换为3D球体上的位置
  const x = ((marker.lon + 180) / 360) * 100
  const y = ((90 - marker.lat) / 180) * 100

  return {
    left: `${x}%`,
    top: `${y}%`
  }
}
</script>

<style scoped>
.globe-container {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
}

.globe {
  position: relative;
  width: v-bind('props.size + "px"');
  height: v-bind('props.size + "px"');
  animation: globeRotate 60s linear infinite;
}

@keyframes globeRotate {
  from { transform: rotateY(0deg); }
  to { transform: rotateY(360deg); }
}

.globe-surface {
  position: absolute;
  width: 100%;
  height: 100%;
  border-radius: 50%;
  background: linear-gradient(135deg, #0a1628 0%, #0d2d4a 50%, #0a1628 100%);
  box-shadow:
    inset -20px -20px 60px rgba(0, 0, 0, 0.8),
    inset 10px 10px 40px rgba(0, 212, 255, 0.1),
    0 0 60px rgba(0, 212, 255, 0.3),
    0 0 100px rgba(0, 102, 255, 0.2);
  overflow: hidden;
}

.globe-grid {
  position: absolute;
  width: 100%;
  height: 100%;
  background-image:
    linear-gradient(rgba(0, 212, 255, 0.3) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0, 212, 255, 0.3) 1px, transparent 1px),
    linear-gradient(rgba(0, 212, 255, 0.15) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0, 212, 255, 0.15) 1px, transparent 1px);
  background-size:
    50% 50px,
    50px 50%,
    10% 10px,
    10px 10%;
  background-position: center center;
  border-radius: 50%;
  opacity: 0.6;
}

.globe-glow {
  position: absolute;
  top: 10%;
  left: 15%;
  width: 30%;
  height: 30%;
  background: radial-gradient(circle, rgba(0, 212, 255, 0.2) 0%, transparent 70%);
  border-radius: 50%;
  filter: blur(10px);
}

.globe-ring {
  position: absolute;
  border: 2px solid rgba(0, 212, 255, 0.4);
  border-radius: 50%;
  animation: ringPulse 3s ease-in-out infinite;
}

.ring-1 {
  top: -20px;
  left: -20px;
  right: -20px;
  bottom: -20px;
  animation-delay: 0s;
}

.ring-2 {
  top: -40px;
  left: -40px;
  right: -40px;
  bottom: -40px;
  animation-delay: 0.5s;
  border-color: rgba(0, 212, 255, 0.2);
}

.ring-3 {
  top: -60px;
  left: -60px;
  right: -60px;
  bottom: -60px;
  animation-delay: 1s;
  border-color: rgba(0, 212, 255, 0.1);
}

@keyframes ringPulse {
  0%, 100% {
    opacity: 0.6;
    transform: scale(1);
  }
  50% {
    opacity: 0.3;
    transform: scale(1.02);
  }
}

.marker {
  position: absolute;
  transform: translate(-50%, -50%);
  z-index: 10;
}

.marker-dot {
  width: 8px;
  height: 8px;
  background: #ff073a;
  border-radius: 50%;
  box-shadow: 0 0 10px #ff073a, 0 0 20px rgba(255, 7, 58, 0.5);
  animation: markerBlink 1s ease-in-out infinite;
}

.marker-pulse {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 20px;
  height: 20px;
  background: rgba(255, 7, 58, 0.3);
  border-radius: 50%;
  animation: markerPulse 1.5s ease-out infinite;
}

@keyframes markerBlink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

@keyframes markerPulse {
  0% {
    transform: translate(-50%, -50%) scale(0.5);
    opacity: 1;
  }
  100% {
    transform: translate(-50%, -50%) scale(2);
    opacity: 0;
  }
}

.globe-info {
  text-align: center;
  background: rgba(11, 26, 42, 0.8);
  border: 1px solid rgba(0, 212, 255, 0.3);
  padding: 10px 20px;
  clip-path: polygon(10px 0, 100% 0, calc(100% - 10px) 100%, 0 100%);
}

.info-label {
  font-family: var(--font-lcd);
  font-size: 9px;
  letter-spacing: 2px;
  color: var(--text-muted);
  margin-bottom: 4px;
}

.info-value {
  font-family: var(--font-lcd);
  font-size: 14px;
  letter-spacing: 2px;
  color: var(--cyan-glow);
  text-shadow: 0 0 10px var(--cyan-glow);
}
</style>
