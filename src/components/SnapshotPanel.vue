<template>
  <div class="snapshot-panel tech-card">
    <div class="panel-header">
      <h3 class="panel-title gradient-title">现场快照回传区</h3>
      <div class="panel-actions">
        <el-upload
          class="upload-button"
          :show-file-list="false"
          :on-change="handleFileChange"
          :before-upload="beforeUpload"
        >
          <el-button size="small" type="primary">
            <span class="upload-icon">📤</span> 上传照片
          </el-button>
        </el-upload>
        <el-button size="small" type="info" @click="loadLatestPhoto">
          <span class="refresh-icon">🔄</span> 刷新快照
        </el-button>
      </div>
    </div>

    <div class="snapshot-container hud-overlay">
      <!-- 图片容器 -->
      <div class="image-wrapper">
        <!-- 加载状态 -->
        <div v-if="loading" class="image-placeholder">
          <div class="placeholder-icon">🔄</div>
          <div class="placeholder-text">加载中...</div>
          <div class="placeholder-subtext">正在处理图片</div>
        </div>
        
        <!-- 错误状态 -->
        <div v-else-if="error || !snapshot.imageUrl" class="image-placeholder">
          <div class="placeholder-icon">📷</div>
          <div class="placeholder-text">等待回传</div>
          <div class="placeholder-subtext">请上传照片或检查设备连接</div>
        </div>
        
        <!-- 真实图片 -->
        <img v-else :src="snapshot.imageUrl" alt="现场快照" class="snapshot-image">

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
      </div>

      <!-- 快照信息 -->
      <div class="snapshot-info">
        <div class="info-item">
          <span class="info-label">点位名称:</span>
          <span class="info-value lcd-number">{{ snapshot.pointName }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">采集时间:</span>
          <span class="info-value lcd-display">{{ snapshot.timestamp }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">图片URL:</span>
          <span class="info-value lcd-display url">{{ snapshot.imageUrl }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">状态:</span>
          <span class="info-value" :class="snapshot.hasAlert ? 'alert' : 'normal'">
            {{ snapshot.hasAlert ? '警报' : '正常' }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { uploadPhoto, getLatestPhoto, getMockLatestPhoto } from '@/api/photo'
import { USE_REAL_API } from '@/utils/constants'

interface Snapshot {
  id: string
  pointName: string
  timestamp: string
  imageUrl: string
  coordinates: {
    lat: string
    lon: string
  }
  hasAlert: boolean
  h2sLevel?: number
}

const snapshot = ref<Snapshot>({
  id: '1',
  pointName: '检测点A-001',
  timestamp: new Date().toISOString(),
  imageUrl: '',
  coordinates: {
    lat: '22.5431',
    lon: '114.0523'
  },
  hasAlert: false
})

const loading = ref(false)
const error = ref(false)

const currentTimestamp = computed(() => {
  return new Date().toISOString().split('T')[1].split('.')[0] + 'Z'
})

const leakBoxStyle = {
  top: '30%',
  left: '40%',
  width: '20%',
  height: '20%'
}

const emit = defineEmits<{
  'update:alert': [data: {
    hasAlert: boolean
    coordinates?: {
      lat: string
      lon: string
    }
  }]
  'upload-success': [data: {
    imageUrl: string
    timestamp: string
  }]
}>()

const loadLatestPhoto = async () => {
  loading.value = true
  error.value = false
  try {
    if (USE_REAL_API) {
      // 调用后端接口获取最新照片
      const latestPhoto = await getLatestPhoto()
      
      if (latestPhoto) {
        updateSnapshot(latestPhoto)
      } else {
        // 如果接口不存在，使用mock数据
        const mockPhoto = getMockLatestPhoto()
        updateSnapshot(mockPhoto)
      }
    } else {
      // Mock模式，直接使用mock数据
      console.log('[Mock] 跳过请求，使用mock数据')
      const mockPhoto = getMockLatestPhoto()
      updateSnapshot(mockPhoto)
    }
    ElMessage.success('快照已刷新')
  } catch (err) {
    console.error('Failed to load latest photo:', err)
    error.value = true
    ElMessage.error('加载快照失败')
  } finally {
    loading.value = false
  }
}

const updateSnapshot = (photoData: any) => {
  const newSnapshot = {
    id: String(Math.random()),
    pointName: '检测点A-001',
    timestamp: photoData.timestamp || new Date().toISOString(),
    imageUrl: photoData.photo_url || photoData.url || '',
    coordinates: {
      lat: photoData.lat?.toString() || '22.5431',
      lon: photoData.lng?.toString() || '114.0523'
    },
    hasAlert: photoData.alert || (photoData.h2s_level && photoData.h2s_level > 10),
    h2sLevel: photoData.h2s_level
  }
  
  snapshot.value = newSnapshot
  
  // 触发警报状态更新事件
  emit('update:alert', {
    hasAlert: newSnapshot.hasAlert,
    coordinates: newSnapshot.coordinates
  })
  
  // 触发上传成功事件
  emit('upload-success', {
    imageUrl: newSnapshot.imageUrl,
    timestamp: newSnapshot.timestamp
  })
}

const handleFileChange = async (file: any) => {
  await uploadFile(file.raw)
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

const uploadFile = async (file: File) => {
  loading.value = true
  error.value = false
  try {
    let imageUrl = ''
    let timestamp = new Date().toISOString()
    
    if (USE_REAL_API) {
      // 上传照片到后端
      const response = await uploadPhoto(file, 'drone-001')
      imageUrl = response.url
      // 从返回的URL解析时间戳（如果文件名包含时间戳）
      timestamp = parseTimestampFromFilename(response.filename) || timestamp
    } else {
      // Mock模式，直接使用本地文件
      console.log('[Mock] 跳过请求，使用本地文件')
      imageUrl = URL.createObjectURL(file)
    }
    
    // 更新快照
    const newSnapshot = {
      id: String(Math.random()),
      pointName: '检测点A-001',
      timestamp: timestamp,
      imageUrl: imageUrl,
      coordinates: {
        lat: '22.5431', // 暂时使用mock坐标
        lon: '114.0523'
      },
      hasAlert: true, // 上传新图片时触发警报
      h2sLevel: 15.5
    }
    
    snapshot.value = newSnapshot
    
    // 触发警报状态更新事件
    emit('update:alert', {
      hasAlert: newSnapshot.hasAlert,
      coordinates: newSnapshot.coordinates
    })
    
    // 触发上传成功事件
    emit('upload-success', {
      imageUrl: newSnapshot.imageUrl,
      timestamp: newSnapshot.timestamp
    })
    
    ElMessage.success('照片上传成功')
  } catch (err) {
    console.error('Failed to upload photo:', err)
    error.value = true
    ElMessage.error('上传失败，请重试')
  } finally {
    loading.value = false
  }
}

const parseTimestampFromFilename = (filename: string): string | null => {
  // 假设文件名格式为：photo_20230425_123456.jpg
  const timestampMatch = filename.match(/photo_(\d{8})_(\d{6})/)
  if (timestampMatch) {
    const [, date, time] = timestampMatch
    const year = date.substring(0, 4)
    const month = date.substring(4, 6)
    const day = date.substring(6, 8)
    const hour = time.substring(0, 2)
    const minute = time.substring(2, 4)
    const second = time.substring(4, 6)
    return `${year}-${month}-${day}T${hour}:${minute}:${second}Z`
  }
  return null
}

onMounted(() => {
  // 初始化时加载最新照片
  loadLatestPhoto()
})
</script>

<style scoped>
.snapshot-panel {
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

.upload-icon {
  margin-right: 6px;
}

.refresh-icon {
  margin-right: 6px;
}

.upload-button {
  display: inline-block;
}

.info-item .url {
  font-size: 10px;
  word-break: break-all;
  line-height: 1.2;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.snapshot-container {
  position: relative;
  background: rgba(0, 0, 0, 0.8);
  border: 1px solid var(--border-color);
  padding: 16px;
  clip-path: polygon(0 8px, 8px 0, calc(100% - 8px) 0, 100% 8px, 100% calc(100% - 8px), calc(100% - 8px) 100%, 8px 100%, 0 calc(100% - 8px));
  overflow: hidden;
}

.image-wrapper {
  position: relative;
  width: 100%;
  height: 300px;
  background: rgba(0, 0, 0, 0.5);
  overflow: hidden;
  clip-path: polygon(0 6px, 6px 0, 100% 0, 100% calc(100% - 6px), calc(100% - 6px) 100%, 0 100%);
}

.snapshot-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: all 0.3s ease;
}

.image-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-muted);
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
  background: rgba(57, 255, 20, 0.2);
  pointer-events: none;
  mix-blend-mode: overlay;
}

/* 十字瞄准线 */
.crosshair {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 100px;
  height: 100px;
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
  background: conic-gradient(from 0deg, transparent 0deg, rgba(0, 255, 255, 0.3) 30deg, transparent 60deg);
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

/* 快照信息 */
.snapshot-info {
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
  min-width: 80px;
  font-size: 12px;
}

.info-value {
  color: var(--text-primary);
  font-weight: 500;
  flex: 1;
  min-width: 0;
}

.snapshot-info .lcd-number,
.snapshot-info .lcd-display {
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
  .snapshot-container {
    padding: 12px;
  }
  
  .image-wrapper {
    height: 200px;
  }
  
  .snapshot-info {
    grid-template-columns: 1fr;
  }

  .snapshot-info .lcd-number,
  .snapshot-info .lcd-display {
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
}
</style>