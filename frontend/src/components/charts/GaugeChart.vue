<template>
  <div ref="chartRef" :style="{ width: width, height: height }"></div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue'
import * as echarts from 'echarts'

const props = withDefaults(defineProps<{
  value: number
  name: string
  min?: number
  max?: number
  width?: string
  height?: string
  color?: string[]
}>(), {
  min: 0,
  max: 100,
  width: '100%',
  height: '300px',
  color: () => ['#6366f1', '#8b5cf6', '#a78bfa'],
})

const chartRef = ref<HTMLElement | null>(null)
let chartInstance: echarts.ECharts | null = null
let resizeObserver: ResizeObserver | null = null

function buildOption(): echarts.EChartsOption {
  return {
    series: [
      {
        type: 'gauge',
        min: props.min,
        max: props.max,
        progress: {
          show: true,
          width: 18,
        },
        axisLine: {
          lineStyle: {
            width: 18,
          },
        },
        axisTick: {
          show: false,
        },
        splitLine: {
          length: 10,
          lineStyle: {
            width: 2,
            color: '#999',
          },
        },
        axisLabel: {
          distance: 25,
          color: '#999',
          fontSize: 12,
        },
        anchor: {
          show: true,
          showAbove: true,
          size: 20,
          itemStyle: {
            borderWidth: 8,
            borderColor: props.color[0],
          },
        },
        detail: {
          valueAnimation: true,
          fontSize: 28,
          offsetCenter: [0, '70%'],
          formatter: `{value}%`,
          color: '#1f2937',
        },
        data: [
          {
            value: props.value,
            name: props.name,
          },
        ],
      },
    ],
  }
}

function initChart() {
  if (!chartRef.value) return
  chartInstance = echarts.init(chartRef.value)
  chartInstance.setOption(buildOption())
  resizeObserver = new ResizeObserver(() => {
    chartInstance?.resize()
  })
  resizeObserver.observe(chartRef.value)
}

watch(() => props.value, () => {
  if (chartInstance) {
    chartInstance.setOption(buildOption())
  }
})

onMounted(() => {
  initChart()
})

onUnmounted(() => {
  resizeObserver?.disconnect()
  chartInstance?.dispose()
  chartInstance = null
})
</script>
