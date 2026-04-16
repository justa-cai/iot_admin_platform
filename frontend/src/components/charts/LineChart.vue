<template>
  <div ref="chartRef" :style="{ width: width, height: height }"></div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue'
import * as echarts from 'echarts'

const props = withDefaults(defineProps<{
  option: echarts.EChartsOption
  width?: string
  height?: string
  theme?: string
}>(), {
  width: '100%',
  height: '400px',
  theme: 'default',
})

const chartRef = ref<HTMLElement | null>(null)
let chartInstance: echarts.ECharts | null = null
let resizeObserver: ResizeObserver | null = null

function initChart() {
  if (!chartRef.value) return
  chartInstance = echarts.init(chartRef.value)
  chartInstance.setOption(props.option)
  resizeObserver = new ResizeObserver(() => {
    chartInstance?.resize()
  })
  resizeObserver.observe(chartRef.value)
}

watch(() => props.option, (newOption) => {
  if (chartInstance) {
    chartInstance.setOption(newOption, true)
  }
}, { deep: true })

onMounted(() => {
  initChart()
})

onUnmounted(() => {
  resizeObserver?.disconnect()
  chartInstance?.dispose()
  chartInstance = null
})

defineExpose({
  getChart: () => chartInstance,
  resize: () => chartInstance?.resize(),
})
</script>
