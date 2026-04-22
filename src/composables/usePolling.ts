import { onBeforeUnmount, onMounted, ref } from 'vue'

export interface UsePollingOptions {
  immediate?: boolean
}

export function usePolling<T extends unknown>(
  fetcher: () => Promise<T>,
  intervalMs: number,
  options: UsePollingOptions = { immediate: true }
) {
  const data = ref<T | null>(null)
  const error = ref<string | null>(null)
  const loading = ref(false)

  let timer: number | undefined

  async function execute() {
    loading.value = true
    error.value = null
    try {
      data.value = await fetcher()
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : '轮询请求失败'
    } finally {
      loading.value = false
    }
  }

  function start() {
    if (timer) return
    if (options.immediate) void execute()
    timer = window.setInterval(() => void execute(), intervalMs)
  }

  function stop() {
    if (!timer) return
    window.clearInterval(timer)
    timer = undefined
  }

  onMounted(start)
  onBeforeUnmount(stop)

  return {
    data,
    error,
    loading,
    start,
    stop,
    refresh: execute
  }
}

