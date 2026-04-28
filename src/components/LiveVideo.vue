<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'

const props = withDefaults(defineProps<{
  taskCode?: string
  refreshSeconds?: number
  streamBase?: string
}>(), {
  taskCode: 'TASK_001',
  refreshSeconds: 1,
  streamBase: '/api/backend/live'
})

type StreamState = 'loading' | 'connected' | 'error'

const state = ref<StreamState>('loading')
const imgKey = ref(Date.now())
const forceRefresh = ref(true)
let refreshTimer: number | null = null

const streamUrl = computed(() => {
  const params = new URLSearchParams({
    task_code: props.taskCode,
    format: 'mjpeg'
  })

  // 兼容“后端非长连接场景”：定时刷新 URL 来拿新帧
  if (forceRefresh.value) {
    params.set('_ts', String(imgKey.value))
  }

  return `${props.streamBase}?${params.toString()}`
})

function startRefreshTimer() {
  stopRefreshTimer()
  if (props.refreshSeconds <= 0) return
  refreshTimer = window.setInterval(() => {
    imgKey.value = Date.now()
  }, props.refreshSeconds * 1000)
}

function stopRefreshTimer() {
  if (refreshTimer) {
    window.clearInterval(refreshTimer)
    refreshTimer = null
  }
}

function handleLoad() {
  state.value = 'connected'
  // 若后端是标准 MJPEG 长连接，首次成功后不再强制拼时间戳
  forceRefresh.value = false
}

function handleError() {
  state.value = 'error'
  forceRefresh.value = true
}

function manualRefresh() {
  state.value = 'loading'
  forceRefresh.value = true
  imgKey.value = Date.now()
}

watch(
  () => props.taskCode,
  () => {
    state.value = 'loading'
    forceRefresh.value = true
    imgKey.value = Date.now()
  },
  { immediate: true }
)

watch(
  () => props.refreshSeconds,
  () => {
    startRefreshTimer()
  },
  { immediate: true }
)

onBeforeUnmount(() => {
  stopRefreshTimer()
})
</script>

<template>
  <div class="live-video">
    <div class="frame-box">
      <img
        :src="streamUrl"
        alt="live video stream"
        class="frame-image"
        @load="handleLoad"
        @error="handleError"
      />

      <div v-if="state === 'loading'" class="overlay overlay-loading">
        加载中...
      </div>
      <div v-else-if="state === 'error'" class="overlay overlay-error">
        加载失败
      </div>
    </div>

    <div class="meta">
      <span>Task: {{ taskCode }}</span>
      <button type="button" class="refresh-btn" @click="manualRefresh">刷新</button>
    </div>
  </div>
</template>

<style scoped>
.live-video {
  width: 100%;
}

.frame-box {
  position: relative;
  width: 100%;
  aspect-ratio: 16 / 9;
  border-radius: 10px;
  overflow: hidden;
  background: #0b0d10;
  border: 1px solid var(--border-color, #2a2f38);
}

.frame-image {
  width: 100%;
  height: 100%;
  object-fit: contain;
  display: block;
}

.overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 500;
  color: #fff;
  background: rgba(0, 0, 0, 0.45);
}

.overlay-error {
  background: rgba(120, 20, 20, 0.48);
}

.meta {
  margin-top: 8px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: var(--text-secondary, #a8b0be);
  font-size: 12px;
}

.refresh-btn {
  cursor: pointer;
  border: 1px solid var(--border-color, #2a2f38);
  background: transparent;
  color: inherit;
  border-radius: 6px;
  padding: 2px 8px;
}

.refresh-btn:hover {
  border-color: var(--primary, #22b8ff);
  color: var(--primary, #22b8ff);
}
</style>
