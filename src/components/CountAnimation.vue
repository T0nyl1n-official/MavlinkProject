<template>
  <span class="count-animation lcd-display">{{ displayValue }}</span>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'

interface Props {
  value: number
  duration?: number
  suffix?: string
}

const props = withDefaults(defineProps<Props>(), {
  duration: 2000,
  suffix: ''
})

const displayValue = ref('0')

const animateCount = (start: number, end: number, duration: number) => {
  let startTime: number | null = null
  const step = (timestamp: number) => {
    if (!startTime) startTime = timestamp
    const progress = Math.min((timestamp - startTime) / duration, 1)
    const currentValue = Math.floor(start + (end - start) * progress)
    displayValue.value = currentValue + props.suffix
    if (progress < 1) {
      requestAnimationFrame(step)
    }
  }
  requestAnimationFrame(step)
}

watch(() => props.value, (newValue, oldValue) => {
  animateCount(oldValue || 0, newValue, props.duration)
}, { immediate: true })

onMounted(() => {
  animateCount(0, props.value, props.duration)
})
</script>

<style scoped>
.count-animation {
  font-family: var(--font-lcd);
  font-size: 24px;
  font-weight: 700;
  color: var(--cyan-glow);
  text-shadow: 0 0 10px var(--cyan-glow), 0 0 20px rgba(0, 212, 255, 0.5);
  letter-spacing: 2px;
  transition: all 0.3s ease;
}
</style>