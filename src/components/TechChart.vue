<template>
  <div class="tech-chart tech-card">
    <div class="panel-header">
      <h3 class="panel-title gradient-title">{{ title }}</h3>
    </div>
    <div class="chart-container">
      <div ref="chartRef" class="chart"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import * as echarts from 'echarts'

interface Props {
  title: string
  data?: number[]
  categories?: string[]
}

const props = withDefaults(defineProps<Props>(), {
  data: () => [120, 190, 130, 150, 170, 160, 180, 200, 190, 210, 230, 220],
  categories: () => ['1月', '2月', '3月', '4月', '5月', '6月', '7月', '8月', '9月', '10月', '11月', '12月']
})

const chartRef = ref<HTMLElement | null>(null)
let chart: echarts.ECharts | null = null

const initChart = () => {
  if (chartRef.value) {
    chart = echarts.init(chartRef.value)
    updateChart()
  }
}

const updateChart = () => {
  if (!chart) return

  const option = {
    backgroundColor: 'transparent',
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: props.categories,
      axisLine: {
        lineStyle: {
          color: 'rgba(0, 212, 255, 0.3)'
        }
      },
      axisLabel: {
        color: 'rgba(153, 204, 238, 0.8)',
        fontSize: 10,
        fontFamily: 'var(--font-tech)'
      },
      axisTick: {
        show: false
      }
    },
    yAxis: {
      type: 'value',
      axisLine: {
        lineStyle: {
          color: 'rgba(0, 212, 255, 0.3)'
        }
      },
      axisLabel: {
        color: 'rgba(153, 204, 238, 0.8)',
        fontSize: 10,
        fontFamily: 'var(--font-tech)'
      },
      axisTick: {
        show: false
      },
      splitLine: {
        lineStyle: {
          color: 'rgba(0, 212, 255, 0.1)'
        }
      }
    },
    series: [
      {
        name: '数据',
        type: 'line',
        stack: 'Total',
        data: props.data,
        lineStyle: {
          color: '#00ffff',
          width: 2,
          shadowColor: 'rgba(0, 255, 255, 0.5)',
          shadowBlur: 10,
          shadowOffsetY: 5
        },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            {
              offset: 0,
              color: 'rgba(0, 255, 255, 0.3)'
            },
            {
              offset: 1,
              color: 'rgba(0, 255, 255, 0.05)'
            }
          ])
        },
        itemStyle: {
          color: '#00ffff',
          borderColor: '#00ffff',
          borderWidth: 2,
          shadowColor: 'rgba(0, 255, 255, 0.5)',
          shadowBlur: 5
        },
        symbol: 'circle',
        symbolSize: 6,
        emphasis: {
          focus: 'series',
          itemStyle: {
            color: '#00ffff',
            borderColor: '#ffffff',
            borderWidth: 3,
            shadowColor: 'rgba(0, 255, 255, 0.8)',
            shadowBlur: 10
          }
        }
      }
    ]
  }

  chart.setOption(option)
}

watch(() => props.data, () => {
  updateChart()
}, { deep: true })

watch(() => props.categories, () => {
  updateChart()
}, { deep: true })

const handleResize = () => {
  chart?.resize()
}

onMounted(() => {
  initChart()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  chart?.dispose()
})
</script>

<style scoped>
.tech-chart {
  padding: 20px;
  min-height: 300px;
}

.panel-header {
  margin-bottom: 20px;
}

.panel-title {
  font-size: 16px;
  font-weight: 600;
  margin: 0;
  text-transform: uppercase;
  letter-spacing: 2px;
}

.chart-container {
  position: relative;
  width: 100%;
  height: 250px;
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid var(--border-color);
  clip-path: polygon(0 6px, 6px 0, 100% 0, 100% calc(100% - 6px), calc(100% - 6px) 100%, 0 100%);
}

.chart {
  width: 100%;
  height: 100%;
}

@media (max-width: 768px) {
  .chart-container {
    height: 200px;
  }
}
</style>