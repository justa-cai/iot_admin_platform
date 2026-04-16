import { ref, watch, onMounted, onUnmounted, type Ref } from 'vue'
import * as echarts from 'echarts'

export function useECharts(chartRef: Ref<HTMLElement | null>) {
  const chart = ref<echarts.ECharts | null>(null)
  let resizeObserver: ResizeObserver | null = null

  function initChart() {
    if (!chartRef.value) return
    chart.value = echarts.init(chartRef.value, 'dark')
    resizeObserver = new ResizeObserver(() => {
      chart.value?.resize()
    })
    resizeObserver.observe(chartRef.value)
  }

  function setOption(option: echarts.EChartsOption, notMerge: boolean = false) {
    if (!chart.value) initChart()
    chart.value?.setOption(option, notMerge)
  }

  function resize() {
    chart.value?.resize()
  }

  function dispose() {
    if (resizeObserver) {
      resizeObserver.disconnect()
      resizeObserver = null
    }
    chart.value?.dispose()
    chart.value = null
  }

  onMounted(() => {
    initChart()
  })

  onUnmounted(() => {
    dispose()
  })

  return { chart, setOption, resize, dispose }
}
