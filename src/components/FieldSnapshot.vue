<template>
  <div class="snapshot-container">
    <div class="snapshot-header">
      <h3 class="snapshot-title gradient-title">现场快照回传区</h3>
      <div class="snapshot-actions">
        <el-upload
          class="upload-button"
          :show-file-list="false"
          :on-change="handleFileChange"
          :before-upload="beforeUpload"
          :auto-upload="false"
        >
          <el-button size="small" type="primary">
            <span class="upload-icon">📤</span> 上传照片
          </el-button>
        </el-upload>
        <el-button size="small" type="warning" @click="simulateAlert">
          <span class="alert-icon">⚠</span> 模拟警报
        </el-button>
        <el-button size="small" type="info" @click="refreshSnapshots">
          <span class="refresh-icon">🔄</span> 刷新
        </el-button>
      </div>
    </div>

    <div class="snapshots-grid">
      <div v-for="snapshot in snapshots" :key="snapshot.id" class="snapshot-card" :class="{ 'has-alert': snapshot.hasAlert }">
        <!-- 快照信息 -->
        <div class="snapshot-info">
          <div class="snapshot-point">
            <span class="point-label">点位:</span>
            <span class="point-value lcd-number">{{ snapshot.pointName }}</span>
          </div>
          <div class="snapshot-timestamp">
            <span class="timestamp-label">时间:</span>
            <span class="timestamp-value lcd-display">{{ currentTimestamp }}</span>
          </div>
          <div class="snapshot-coords">
            <span class="coords-label">坐标:</span>
            <span class="coords-value lcd-display">{{ snapshot.coordinates.lat }}, {{ snapshot.coordinates.lon }}</span>
          </div>
        </div>

        <!-- 图片占位区 -->
        <div class="snapshot-image-container">
          <!-- 有图片时显示图片 -->
          <img v-if="snapshot.imageUrl" :src="snapshot.imageUrl" alt="现场快照" class="snapshot-image">
          <!-- 无图片时显示占位符 -->
          <div v-else class="snapshot-image-placeholder">
            <div class="placeholder-icon">📷</div>
            <div class="placeholder-text">树莓派回传图片</div>
            <div class="placeholder-subtext">预留占位区</div>
          </div>

          <!-- 绿色色调叠加 -->
          <div class="green-tint-overlay"></div>

          <!-- 十字瞄准线 -->
          <div class="crosshair"></div>

          <!-- 泄漏点红色呼吸闪烁框 -->
          <div v-if="snapshot.hasAlert" class="gas-leak-warning" :style="leakBoxStyle">
            <div class="gas-leak-label">
              WARNING: H2S LEVEL CRITICAL
            </div>
          </div>

          <!-- 动态时间戳 -->
          <div class="hud-timestamp">
            {{ currentTimestamp }}
          </div>

          <!-- 经纬度坐标 -->
          <div class="hud-coords">
            LAT: {{ snapshot.coordinates.lat }}
            <br>
            LON: {{ snapshot.coordinates.lon }}
          </div>

          <!-- 雷达扫描波纹 -->
          <div class="radar-sweep"></div>

          <!-- 警报边框 -->
          <div v-if="snapshot.hasAlert" class="alert-border">
            <div class="alert-label">
              <span class="alert-icon">⚠</span>
              {{ snapshot.alertMessage }}
            </div>
          </div>
        </div>

        <!-- 路径示意图 -->
        <div v-if="snapshot.showPath" class="path-container">
          <div class="path-title">飞行路径</div>
          <div class="path-svg">
            <svg width="100%" height="60" viewBox="0 0 300 60">
              <!-- 机巢 -->
              <circle cx="30" cy="30" r="8" fill="#39ff14" stroke="#00aa33" stroke-width="2">
                <animate attributeName="r" values="8;10;8" dur="2s" repeatCount="indefinite" />
              </circle>
              <text x="30" y="50" text-anchor="middle" fill="#39ff14" font-size="10" font-family="var(--font-lcd)">机巢</text>
              
              <!-- 泄漏点 -->
              <circle cx="270" cy="30" r="8" fill="#ff073a" stroke="#cc0033" stroke-width="2">
                <animate attributeName="r" values="8;12;8" dur="1.5s" repeatCount="indefinite" />
                <animate attributeName="opacity" values="0.8;1;0.8" dur="1.5s" repeatCount="indefinite" />
              </circle>
              <text x="270" y="50" text-anchor="middle" fill="#ff073a" font-size="10" font-family="var(--font-lcd)">泄漏点</text>
              
              <!-- 飞行轨迹 -->
              <path 
                d="M30,30 L100,15 L180,45 L270,30" 
                stroke="#ff073a" 
                stroke-width="2" 
                fill="none" 
                stroke-dasharray="8,4" 
                stroke-linecap="round"
              >
                <animate attributeName="stroke-dashoffset" values="0;12" dur="1.5s" repeatCount="indefinite" />
              </path>
              
              <!-- 距离标签 -->
              <rect x="125" y="22" width="50" height="16" fill="rgba(0,0,0,0.7)" stroke="#ff073a" stroke-width="1" rx="3" />
              <text x="150" y="34" text-anchor="middle" fill="#ff073a" font-size="10" font-family="var(--font-lcd)">2.3km</text>
            </svg>
          </div>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-if="snapshots.length === 0" class="empty-snapshot">
        <div class="empty-icon">📷</div>
        <div class="empty-text">暂无现场快照数据</div>
        <div class="empty-subtext">点击上传照片或模拟警报按钮</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { uploadPhoto } from '@/api/photo'
import { USE_REAL_API } from '@/utils/constants'

interface Coordinates {
  lat: string
  lon: string
  alt?: string
}

interface Snapshot {
  id: string
  pointName: string
  timestamp: string
  coordinates: Coordinates
  hasAlert: boolean
  alertMessage: string
  showPath: boolean
  imageUrl?: string
}

const emit = defineEmits<{
  'upload-success': [data: { imageUrl: string; coordinates: Coordinates }]
  'alert-triggered': [data: { coordinates: Coordinates; message: string }]
  'trigger-trajectory': []
}>()

const snapshots = ref<Snapshot[]>([])

const currentTimestamp = computed(() => {
  return new Date().toISOString().split('T')[1].split('.')[0] + 'Z'
})

const leakBoxStyle = {
  top: '35%',
  left: '35%',
  width: '30%',
  height: '25%'
}

const generateMockSnapshots = () => {
  return [
    {
      id: '1',
      pointName: '检测点A-001',
      timestamp: new Date().toISOString(),
      coordinates: {
        lat: '22.5531',
        lon: '114.0623',
        alt: '120'
      },
      hasAlert: true,
      alertMessage: 'H2S 浓度超标',
      showPath: true,
      imageUrl: 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=drone%20aerial%20view%20of%20industrial%20facility%20with%20gas%20leak%20detection&image_size=landscape_16_9'
    }
  ]
}

const refreshSnapshots = () => {
  snapshots.value = generateMockSnapshots()
  ElMessage.success('快照已刷新')
}

const simulateAlert = () => {
  const mockCoords = {
    lat: '22.5531',
    lon: '114.0623'
  }
  
  // 触发警报
  if (snapshots.value.length > 0) {
    snapshots.value[0].hasAlert = true
    snapshots.value[0].showPath = true
  } else {
    snapshots.value = [{
      id: '1',
      pointName: '检测点A-001',
      timestamp: new Date().toISOString(),
      coordinates: mockCoords,
      hasAlert: true,
      alertMessage: 'H2S 浓度超标',
      showPath: true,
      imageUrl: ''
    }]
  }
  
  // 发送警报事件
  emit('alert-triggered', {
    coordinates: mockCoords,
    message: `H₂S 泄漏警报 | 坐标 ${mockCoords.lat}, ${mockCoords.lon}`
  })
  
  // 触发轨迹动画
  emit('trigger-trajectory')
  
  ElMessage.warning('警报已触发，无人机正在前往泄漏点')
}

const beforeUpload = (file: any) => {
  const isJpgOrPng = file.type === 'image/jpeg' || file.type === 'image/png'
  const isLt2M = file.size / 1024 / 1024 < 2

  if (!isJpgOrPng) {
    ElMessage.error('只能上传 JPG/PNG 图片!')
    return false
  }
  if (!isLt2M) {
    ElMessage.error('图片大小不能超过 2MB!')
    return false
  }
  return true
}

const handleFileChange = async (file: any) => {
  await uploadFile(file.raw)
}

const uploadFile = async (file: File) => {
  try {
    const mockCoords = {
      lat: '22.5531',
      lon: '114.0623'
    }
    
    let imageUrl = ''
    
    if (USE_REAL_API) {
      // 调用后端接口上传
      const response = await uploadPhoto(file, 'drone-001')
      imageUrl = response.url || URL.createObjectURL(file)
    } else {
      // Mock模式，直接使用本地文件
      imageUrl = URL.createObjectURL(file)
    }
    
    // 更新快照
    snapshots.value = [{
      id: String(Math.random()),
      pointName: '检测点A-001',
      timestamp: new Date().toISOString(),
      coordinates: mockCoords,
      hasAlert: true,
      alertMessage: 'H2S 浓度超标',
      showPath: true,
      imageUrl: imageUrl
    }]
    
    // 发送上传成功事件
    emit('upload-success', {
      imageUrl: imageUrl,
      coordinates: mockCoords
    })
    
    // 发送警报事件
    emit('alert-triggered', {
      coordinates: mockCoords,
      message: `H₂S 泄漏警报 | 坐标 ${mockCoords.lat}, ${mockCoords.lon}`
    })
    
    // 触发轨迹动画
    emit('trigger-trajectory')
    
    ElMessage.success('照片上传成功')
  } catch (error) {
    console.error('Upload failed:', error)
    ElMessage.error('上传失败，请重试')
  }
}

onMounted(() => {
  refreshSnapshots()
})
</script>

<style scoped>
.snapshot-container {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  padding: 24px;
  min-height: 600px;
  clip-path: polygon(0 10px, 10px 0, 100% 0, 100% calc(100% - 10px), calc(100% - 10px) 100%, 0 100%);
  position: relative;
  overflow: visible;
}

.snapshot-container::after {
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

.snapshot-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  flex-wrap: wrap;
  gap: 16px;
}

.snapshot-title {
  font-size: 18px;
  font-weight: 600;
  margin: 0;
  text-transform: uppercase;
  letter-spacing: 2px;
}

.snapshot-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.upload-icon,
.alert-icon,
.refresh-icon {
  margin-right: 6px;
}

.upload-button {
  display: inline-block;
}

.snapshots-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 20px;
  margin-top: 24px;
}

.snapshot-card {
  background: rgba(11, 26, 42, 0.9);
  border: 1px solid var(--border-color);
  padding: 20px;
  transition: all 0.3s ease;
  position: relative;
  clip-path: polygon(0 8px, 8px 0, 100% 0, 100% calc(100% - 8px), calc(100% - 8px) 100%, 0 100%);
  overflow: visible;
}

.snapshot-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  border: 1px solid rgba(0, 255, 255, 0.2);
  pointer-events: none;
  clip-path: polygon(0 8px, 8px 0, 100% 0, 100% calc(100% - 8px), calc(100% - 8px) 100%, 0 100%);
}

.snapshot-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-glow), var(--shadow-hover);
  border-color: var(--cyan-glow);
}

.snapshot-card.has-alert {
  border-color: var(--neon-red);
  box-shadow: 0 0 15px rgba(255, 7, 58, 0.3);
}

.snapshot-info {
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-color);
}

.snapshot-point,
.snapshot-timestamp,
.snapshot-coords {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  font-size: 13px;
}

.point-label,
.timestamp-label,
.coords-label {
  color: var(--text-muted);
  min-width: 40px;
  font-size: 12px;
}

.point-value,
.timestamp-value,
.coords-value {
  color: var(--text-primary);
  font-weight: 500;
}

/* 图片容器 */
.snapshot-image-container {
  position: relative;
  margin: 16px 0;
  height: 200px;
  background: rgba(0, 0, 0, 0.5);
  border: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  clip-path: polygon(0 6px, 6px 0, 100% 0, 100% calc(100% - 6px), calc(100% - 6px) 100%, 0 100%);
}

.snapshot-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.snapshot-image-placeholder {
  text-align: center;
  color: var(--text-muted);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.placeholder-icon {
  font-size: 48px;
  opacity: 0.5;
}

.placeholder-text {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-secondary);
}

.placeholder-subtext {
  font-size: 12px;
  opacity: 0.7;
}

/* 绿色色调叠加 */
.green-tint-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(57, 255, 20, 0.15);
  pointer-events: none;
  mix-blend-mode: overlay;
}

/* 十字瞄准线 */
.crosshair {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 80px;
  height: 80px;
  pointer-events: none;
}

.crosshair::before,
.crosshair::after {
  content: '';
  position: absolute;
  background: var(--neon-green);
  box-shadow: 0 0 5px var(--neon-green);
}

.crosshair::before {
  top: 50%;
  left: 0;
  right: 0;
  height: 1px;
  transform: translateY(-50%);
}

.crosshair::after {
  left: 50%;
  top: 0;
  bottom: 0;
  width: 1px;
  transform: translateX(-50%);
}

/* 气体泄漏警告框 */
.gas-leak-warning {
  position: absolute;
  border: 2px solid var(--neon-red);
  background: rgba(255, 7, 58, 0.1);
  animation: gasLeakPulse 1.5s ease-in-out infinite;
  box-shadow: 0 0 15px rgba(255, 7, 58, 0.5);
}

@keyframes gasLeakPulse {
  0%, 100% {
    border-color: var(--neon-red);
    box-shadow: 0 0 15px rgba(255, 7, 58, 0.5);
  }
  50% {
    border-color: #ff4466;
    box-shadow: 0 0 30px rgba(255, 7, 58, 0.8);
  }
}

.gas-leak-label {
  position: absolute;
  top: -25px;
  left: 50%;
  transform: translateX(-50%);
  color: var(--neon-red);
  font-family: var(--font-lcd);
  font-size: 10px;
  letter-spacing: 1px;
  white-space: nowrap;
  text-shadow: 0 0 5px var(--neon-red);
  background: rgba(0, 0, 0, 0.8);
  padding: 2px 8px;
  clip-path: polygon(4px 0, 100% 0, calc(100% - 4px) 100%, 0 100%);
}

/* HUD 数据叠加层 */
.hud-timestamp {
  position: absolute;
  top: 10px;
  right: 10px;
  font-size: 10px;
  font-family: var(--font-lcd);
  color: var(--neon-green);
  text-shadow: 0 0 5px var(--neon-green);
  background: rgba(0, 0, 0, 0.6);
  padding: 4px 8px;
  clip-path: polygon(4px 0, 100% 0, calc(100% - 4px) 100%, 0 100%);
}

.hud-coords {
  position: absolute;
  bottom: 10px;
  left: 10px;
  font-size: 10px;
  font-family: var(--font-lcd);
  color: var(--neon-green);
  text-shadow: 0 0 5px var(--neon-green);
  background: rgba(0, 0, 0, 0.6);
  padding: 4px 8px;
  line-height: 1.5;
  clip-path: polygon(4px 0, 100% 0, calc(100% - 4px) 100%, 0 100%);
}

/* 雷达扫描效果 */
.radar-sweep {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 100%;
  height: 100%;
  transform: translate(-50%, -50%);
  background: conic-gradient(from 0deg, transparent 0deg, rgba(0, 255, 255, 0.25) 30deg, transparent 60deg);
  animation: radarSweep 4s linear infinite;
  pointer-events: none;
}

@keyframes radarSweep {
  from {
    transform: translate(-50%, -50%) rotate(0deg);
  }
  to {
    transform: translate(-50%, -50%) rotate(360deg);
  }
}

/* 警报边框 */
.alert-border {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  border: 2px solid var(--neon-red);
  box-shadow: 0 0 20px rgba(255, 7, 58, 0.6);
  clip-path: polygon(0 6px, 6px 0, 100% 0, 100% calc(100% - 6px), calc(100% - 6px) 100%, 0 100%);
  animation: alertPulse 1.5s ease-in-out infinite;
  pointer-events: none;
}

@keyframes alertPulse {
  0%, 100% {
    border-color: var(--neon-red);
    box-shadow: 0 0 20px rgba(255, 7, 58, 0.6);
  }
  50% {
    border-color: #ff4466;
    box-shadow: 0 0 35px rgba(255, 7, 58, 0.9);
  }
}

.alert-label {
  position: absolute;
  top: -30px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(0, 0, 0, 0.9);
  border: 1px solid var(--neon-red);
  padding: 4px 12px;
  font-family: var(--font-lcd);
  font-size: 10px;
  letter-spacing: 1px;
  color: var(--neon-red);
  text-shadow: 0 0 5px var(--neon-red);
  display: flex;
  align-items: center;
  gap: 6px;
}

.alert-icon {
  font-size: 12px;
  animation: blink 0.5s step-end infinite;
}

@keyframes blink {
  50% { opacity: 0; }
}

/* 路径容器 */
.path-container {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--border-color);
}

.path-title {
  font-size: 12px;
  color: var(--text-muted);
  margin-bottom: 8px;
  font-family: var(--font-lcd);
  letter-spacing: 1px;
}

.path-svg {
  background: rgba(0, 0, 0, 0.3);
  padding: 12px;
  border-radius: 4px;
  border: 1px solid var(--border-color);
}

.empty-snapshot {
  grid-column: 1 / -1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 40px;
  text-align: center;
  border: 1px dashed var(--border-color);
  background: rgba(11, 26, 42, 0.6);
  clip-path: polygon(0 10px, 10px 0, 100% 0, 100% calc(100% - 10px), calc(100% - 10px) 100%, 0 100%);
}

.empty-icon {
  font-size: 64px;
  opacity: 0.5;
  margin-bottom: 16px;
}

.empty-text {
  font-size: 16px;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.empty-subtext {
  font-size: 14px;
  color: var(--text-muted);
  opacity: 0.8;
}

@media (max-width: 768px) {
  .snapshots-grid {
    grid-template-columns: 1fr;
  }
  .snapshot-header {
    flex-direction: column;
    align-items: stretch;
  }
  .snapshot-actions {
    justify-content: center;
  }
}
</style>