<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { ElTag } from 'element-plus'
import flvjs from 'flv.js'
import { USE_REAL_API } from '@/utils/constants'

type StreamMode = 'mjpeg' | 'ws' | 'flv'
type ConnectionStatus = 'loading' | 'connected' | 'disconnected'

const JPEG_SOI = [0xff, 0xd8]
const JPEG_EOI = [0xff, 0xd9]

const props = withDefaults(defineProps<{
  taskCode?: string
  mode?: StreamMode
  autoReconnect?: boolean
  reconnectInterval?: number
  maxReconnectAttempts?: number
  width?: number | string
  height?: number | string
}>(), {
  taskCode: 'TASK_001',
  mode: 'mjpeg',
  autoReconnect: true,
  reconnectInterval: 3000,
  maxReconnectAttempts: 10,
  width: '100%',
  height: 'auto'
})

const emit = defineEmits<{
  'status-change': [status: ConnectionStatus]
  'error': [error: Error]
}>()

const status = ref<ConnectionStatus>('loading')
const lastError = ref('')
const canvasRef = ref<HTMLCanvasElement | null>(null)
const videoRef = ref<HTMLVideoElement | null>(null)
const wsRef = ref<WebSocket | null>(null)
const flvPlayerRef = ref<flvjs.Player | null>(null)
const reconnectTimer = ref<number | null>(null)
const reconnectAttempts = ref(0)
const mjpegAbortController = ref<AbortController | null>(null)
const mjpegFrameUrl = ref('')
const isMounted = ref(false)
let currentFrameObjectUrl: string | null = null

const token = computed(() => localStorage.getItem('token') || '')

const containerStyle = computed(() => ({
  width: typeof props.width === 'number' ? `${props.width}px` : props.width,
  height: typeof props.height === 'number' ? `${props.height}px` : props.height
}))

const httpUrl = computed(() => {
  const params = new URLSearchParams({
    task_code: props.taskCode
  })
  return `/api/backend/live?${params.toString()}`
})

const mjpegUrl = computed(() => `${httpUrl.value}&format=mjpeg`)
const flvUrl = computed(() => {
  const url = new URL(httpUrl.value, window.location.origin)
  url.searchParams.set('format', 'flv')
  if (token.value) {
    url.searchParams.set('token', token.value)
  }
  return `${url.pathname}${url.search}`
})

const wsUrl = computed(() => {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const params = new URLSearchParams({
    task_code: props.taskCode
  })
  if (token.value) {
    // 浏览器原生 WebSocket 无法自定义 Authorization 头，这里回退为 query 透传 token。
    params.set('token', token.value)
  }
  return `${protocol}//${window.location.host}/api/backend/live/ws?${params.toString()}`
})

const statusType = computed(() => {
  switch (status.value) {
    case 'connected':
      return 'success'
    case 'loading':
      return 'warning'
    default:
      return 'danger'
  }
})

const statusText = computed(() => {
  switch (status.value) {
    case 'connected':
      return '已连接'
    case 'loading':
      return '加载中'
    default:
      return '断开'
  }
})

const modeText = computed(() => {
  switch (props.mode) {
    case 'ws':
      return 'WebSocket'
    case 'flv':
      return 'FLV'
    default:
      return 'MJPEG'
  }
})

function setStatus(nextStatus: ConnectionStatus) {
  status.value = nextStatus
  emit('status-change', nextStatus)
}

function cleanupFrameUrl() {
  if (currentFrameObjectUrl) {
    URL.revokeObjectURL(currentFrameObjectUrl)
    currentFrameObjectUrl = null
  }
  mjpegFrameUrl.value = ''
}

function clearReconnectTimer() {
  if (reconnectTimer.value !== null) {
    window.clearTimeout(reconnectTimer.value)
    reconnectTimer.value = null
  }
}

function stopMjpegStream() {
  mjpegAbortController.value?.abort()
  mjpegAbortController.value = null
  cleanupFrameUrl()
}

function stopWebSocketStream() {
  if (wsRef.value) {
    wsRef.value.onopen = null
    wsRef.value.onmessage = null
    wsRef.value.onerror = null
    wsRef.value.onclose = null
    wsRef.value.close()
    wsRef.value = null
  }
}

function stopFlvStream() {
  if (flvPlayerRef.value) {
    flvPlayerRef.value.pause()
    flvPlayerRef.value.unload()
    flvPlayerRef.value.detachMediaElement()
    flvPlayerRef.value.destroy()
    flvPlayerRef.value = null
  }
}

function cleanupAll() {
  clearReconnectTimer()
  stopMjpegStream()
  stopWebSocketStream()
  stopFlvStream()
}

function handleStreamError(message: string, error?: unknown) {
  lastError.value = message
  setStatus('disconnected')
  emit('error', error instanceof Error ? error : new Error(message))
  scheduleReconnect()
}

function scheduleReconnect() {
  if (!props.autoReconnect || reconnectTimer.value !== null) {
    return
  }
  if (reconnectAttempts.value >= props.maxReconnectAttempts) {
    return
  }

  reconnectAttempts.value += 1
  reconnectTimer.value = window.setTimeout(() => {
    reconnectTimer.value = null
    void startStream()
  }, props.reconnectInterval)
}

function findMarkerIndex(source: Uint8Array, marker: number[], startIndex = 0) {
  for (let i = startIndex; i < source.length - 1; i += 1) {
    if (source[i] === marker[0] && source[i + 1] === marker[1]) {
      return i
    }
  }
  return -1
}

function mergeUint8Arrays(left: Uint8Array, right: Uint8Array) {
  const merged = new Uint8Array(left.length + right.length)
  merged.set(left)
  merged.set(right, left.length)
  return merged
}

function updateMjpegFrame(frameBytes: Uint8Array) {
  const safeBytes = frameBytes.slice()
  const blob = new Blob([safeBytes.buffer], { type: 'image/jpeg' })
  const objectUrl = URL.createObjectURL(blob)
  const previousUrl = currentFrameObjectUrl
  currentFrameObjectUrl = objectUrl
  mjpegFrameUrl.value = objectUrl

  if (previousUrl) {
    URL.revokeObjectURL(previousUrl)
  }

  if (status.value !== 'connected') {
    setStatus('connected')
    reconnectAttempts.value = 0
    lastError.value = ''
  }
}

async function startMjpegStream() {
  stopMjpegStream()
  setStatus('loading')

  const controller = new AbortController()
  mjpegAbortController.value = controller

  try {
    const response = await fetch(mjpegUrl.value, {
      method: 'GET',
      headers: token.value ? { Authorization: `Bearer ${token.value}` } : undefined,
      signal: controller.signal
    })

    if (!response.ok || !response.body) {
      throw new Error(`MJPEG stream request failed: ${response.status}`)
    }

    const reader = response.body.getReader()
    let buffer = new Uint8Array()

    while (isMounted.value) {
      const { done, value } = await reader.read()
      if (done) {
        break
      }

      if (!value) {
        continue
      }

      buffer = mergeUint8Arrays(buffer, value)

      while (true) {
        const start = findMarkerIndex(buffer, JPEG_SOI)
        if (start < 0) {
          break
        }

        const end = findMarkerIndex(buffer, JPEG_EOI, start + 2)
        if (end < 0) {
          buffer = buffer.slice(start)
          break
        }

        const frame = buffer.slice(start, end + 2)
        updateMjpegFrame(frame)
        buffer = buffer.slice(end + 2)
      }
    }

    if (!controller.signal.aborted && isMounted.value) {
      handleStreamError('MJPEG 连接已断开')
    }
  } catch (error) {
    if (controller.signal.aborted) {
      return
    }
    handleStreamError('MJPEG 连接失败', error)
  }
}

function drawFrameToCanvas(frameBlob: Blob) {
  const canvas = canvasRef.value
  const context = canvas?.getContext('2d')

  if (!canvas || !context) {
    return
  }

  const url = URL.createObjectURL(frameBlob)
  const img = new Image()
  img.onload = () => {
    canvas.width = img.width
    canvas.height = img.height
    context.drawImage(img, 0, 0, img.width, img.height)
    URL.revokeObjectURL(url)

    if (status.value !== 'connected') {
      setStatus('connected')
      reconnectAttempts.value = 0
      lastError.value = ''
    }
  }
  img.onerror = () => {
    URL.revokeObjectURL(url)
  }
  img.src = url
}

function startWebSocketStream() {
  stopWebSocketStream()
  setStatus('loading')

  try {
    const ws = new WebSocket(wsUrl.value)
    ws.binaryType = 'blob'
    wsRef.value = ws

    ws.onopen = () => {
      lastError.value = ''
    }

    ws.onmessage = async (event) => {
      if (typeof event.data === 'string') {
        return
      }

      const frameBlob = event.data instanceof Blob
        ? event.data
        : new Blob([await event.data.arrayBuffer()], { type: 'image/jpeg' })

      drawFrameToCanvas(frameBlob)
    }

    ws.onerror = () => {
      handleStreamError('WebSocket 视频流连接失败')
    }

    ws.onclose = () => {
      if (!isMounted.value) {
        return
      }
      handleStreamError('WebSocket 视频流已断开')
    }
  } catch (error) {
    handleStreamError('WebSocket 初始化失败', error)
  }
}

async function startFlvStream() {
  stopFlvStream()
  setStatus('loading')

  if (!flvjs.isSupported()) {
    handleStreamError('当前浏览器不支持 FLV 播放')
    return
  }

  await nextTick()

  if (!videoRef.value) {
    handleStreamError('FLV 播放器初始化失败')
    return
  }

  try {
    const player = flvjs.createPlayer({
      type: 'flv',
      isLive: true,
      hasAudio: false,
      hasVideo: true,
      url: flvUrl.value,
      cors: true
    }, {
      enableWorker: false,
      enableStashBuffer: false,
      autoCleanupSourceBuffer: true
    })

    flvPlayerRef.value = player
    player.attachMediaElement(videoRef.value)
    player.load()

    const handleVideoReady = () => {
      setStatus('connected')
      reconnectAttempts.value = 0
      lastError.value = ''
    }

    videoRef.value.addEventListener('loadeddata', handleVideoReady, { once: true })

    player.on(flvjs.Events.ERROR, (_type, _detail, info) => {
      handleStreamError(info?.msg || 'FLV 播放失败')
    })

    void videoRef.value.play().catch(() => {
      // 某些浏览器需要用户手势，保持已加载状态即可。
    })
  } catch (error) {
    handleStreamError('FLV 播放器创建失败', error)
  }
}

async function startStream() {
  cleanupAll()
  lastError.value = ''

  if (!USE_REAL_API) {
    setStatus('connected')
    return
  }

  if (!props.taskCode) {
    handleStreamError('缺少 taskCode 参数')
    return
  }

  switch (props.mode) {
    case 'ws':
      startWebSocketStream()
      break
    case 'flv':
      await startFlvStream()
      break
    default:
      await startMjpegStream()
      break
  }
}

function refreshStream() {
  reconnectAttempts.value = 0
  void startStream()
}

watch(
  () => [props.taskCode, props.mode, token.value],
  () => {
    if (!isMounted.value) {
      return
    }
    void startStream()
  }
)

onMounted(() => {
  isMounted.value = true
  void startStream()
})

onUnmounted(() => {
  isMounted.value = false
  cleanupAll()
})
</script>

<template>
  <div class="video-stream-container" :style="containerStyle">
    <div class="stream-wrapper">
      <div v-if="!USE_REAL_API" class="mock-placeholder">
        <div class="mock-icon">VIDEO</div>
        <div class="mock-text">实时视频流</div>
        <div class="mock-subtext">Mock 模式 · 视频占位</div>
        <div class="mock-info">
          <span>任务: {{ taskCode || '-' }}</span>
          <span>模式: {{ modeText }}</span>
        </div>
      </div>

      <template v-else>
        <img
          v-if="mode === 'mjpeg'"
          :src="mjpegFrameUrl"
          alt="mjpeg video stream"
          class="stream-media"
        />

        <canvas
          v-else-if="mode === 'ws'"
          ref="canvasRef"
          class="stream-media"
        />

        <video
          v-else
          ref="videoRef"
          class="stream-media"
          autoplay
          muted
          controls
        />
      </template>

      <div v-if="status !== 'connected'" class="stream-mask">
        <span>{{ statusText }}</span>
      </div>

      <div class="stream-overlay">
        <div class="status-tag">
          <el-tag :type="statusType" size="small" effect="dark">
            {{ statusText }}
          </el-tag>
        </div>
        <div class="stream-info">
          <span>任务: {{ taskCode || '-' }}</span>
          <span>模式: {{ modeText }}</span>
        </div>
      </div>

      <div v-if="lastError" class="stream-error">
        {{ lastError }}
      </div>
    </div>

    <div class="stream-controls">
      <button class="control-btn" type="button" @click="refreshStream">刷新</button>
      <span class="stream-mode">{{ modeText }}</span>
    </div>
  </div>
</template>

<style scoped>
.video-stream-container {
  position: relative;
  background: rgba(11, 26, 42, 0.95);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.stream-wrapper {
  position: relative;
  flex: 1;
  min-height: 0;
  aspect-ratio: 16 / 9;
  background: #000;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stream-media {
  width: 100%;
  height: 100%;
  object-fit: contain;
  display: block;
  background: #000;
}

.mock-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  color: var(--text-muted);
  background: linear-gradient(135deg, rgba(30, 136, 229, 0.1) 0%, rgba(11, 26, 42, 0.8) 100%);
  gap: 12px;
}

.mock-icon {
  font-size: 32px;
  opacity: 0.7;
  letter-spacing: 2px;
}

.mock-text {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-secondary);
}

.mock-subtext {
  font-size: 12px;
  color: var(--text-muted);
}

.mock-info {
  display: flex;
  gap: 16px;
  font-size: 11px;
  color: var(--text-muted);
  margin-top: 8px;
}

.stream-mask {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.45);
  color: #fff;
  font-size: 14px;
  z-index: 1;
}

.stream-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  padding: 8px 12px;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  pointer-events: none;
  background: linear-gradient(to bottom, rgba(0, 0, 0, 0.6) 0%, transparent 100%);
}

.status-tag {
  display: flex;
  align-items: center;
}

.stream-info {
  display: flex;
  gap: 12px;
  font-size: 10px;
  color: #9ef7c8;
}

.stream-error {
  position: absolute;
  left: 12px;
  right: 12px;
  bottom: 12px;
  padding: 8px 10px;
  border-radius: 6px;
  background: rgba(255, 68, 68, 0.2);
  color: #ffd6d6;
  font-size: 12px;
}

.stream-controls {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: var(--bg-card);
  border-top: 1px solid var(--border-color);
}

.control-btn {
  background: transparent;
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
  min-width: 56px;
  height: 28px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s;
}

.control-btn:hover {
  border-color: var(--primary);
  color: var(--primary);
  background: rgba(30, 136, 229, 0.1);
}

.stream-mode {
  font-size: 11px;
  color: var(--text-muted);
  letter-spacing: 1px;
}
</style>
