import { onMounted, ref } from 'vue'

export function useCountAnimation(initialValue: number = 0, duration = 1500) {
  const displayValue = ref(initialValue)
  const target = ref(initialValue)
  const isAnimating = ref(false)

  const animateTo = (newTarget: number) => {
    if (newTarget === displayValue.value) return

    target.value = newTarget
    isAnimating.value = true
    const startValue = displayValue.value
    const endValue = newTarget
    const startTime = performance.now()

    const update = (currentTime: number) => {
      const elapsed = currentTime - startTime
      const progress = Math.min(elapsed / duration, 1)

      // 使用 easeOutQuart 缓动函数
      const easeProgress = 1 - Math.pow(1 - progress, 4)

      displayValue.value = Math.round(startValue + (endValue - startValue) * easeProgress)

      if (progress < 1) {
        requestAnimationFrame(update)
      } else {
        isAnimating.value = false
      }
    }

    requestAnimationFrame(update)
  }

  const setValue = (value: number) => {
    animateTo(value)
  }

  onMounted(() => {
    if (initialValue > 0) {
      displayValue.value = 0
      animateTo(initialValue)
    }
  })

  return {
    displayValue,
    target,
    isAnimating,
    setValue,
    animateTo
  }
}
