<template>
  <div class="stat-card" :class="gradientClass">
    <div class="stat-icon">
      <el-icon :size="56"><component :is="icon" /></el-icon>
    </div>
    <div class="stat-value">{{ animatedValue }}</div>
    <div class="stat-label">{{ label }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'

const props = defineProps<{
  value: number
  label: string
  icon: string
  gradient?: 'blue' | 'green' | 'orange' | 'cyan' | 'indigo' | 'rose'
}>()

const gradientClass = ref(`gradient-${props.gradient || 'blue'}`)
const animatedValue = ref(0)

function animateValue(target: number) {
  const duration = 800
  const startTime = Date.now()
  const startVal = animatedValue.value
  const diff = target - startVal

  function step() {
    const elapsed = Date.now() - startTime
    const progress = Math.min(elapsed / duration, 1)
    const eased = 1 - Math.pow(1 - progress, 3)
    animatedValue.value = Math.round(startVal + diff * eased)
    if (progress < 1) {
      requestAnimationFrame(step)
    }
  }
  requestAnimationFrame(step)
}

watch(() => props.value, (newVal) => {
  animateValue(newVal)
})

onMounted(() => {
  animateValue(props.value)
})
</script>
