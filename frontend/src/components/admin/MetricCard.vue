<script setup>
defineProps({
  title: {
    type: String,
    required: true
  },
  value: {
    type: [String, Number],
    required: true
  },
  subtitle: {
    type: String,
    default: ''
  },
  icon: {
    type: String,
    default: ''
  },
  trend: {
    type: String,
    default: '' // 'up', 'down', 'neutral'
  }
})
</script>

<template>
  <div class="metric-card">
    <div class="metric-header">
      <span class="metric-title">{{ title }}</span>
      <span v-if="icon" class="metric-icon">{{ icon }}</span>
    </div>

    <div class="metric-body">
      <span class="metric-value">{{ value }}</span>
    </div>

    <div v-if="subtitle" class="metric-footer">
      <span class="metric-subtitle">{{ subtitle }}</span>
    </div>
  </div>
</template>

<style scoped>
.metric-card {
  background-color: var(--color-surface, #ffffff);
  border-radius: var(--radius-xl, 12px);
  border: 1px solid var(--color-border-soft, #e7edf5);
  padding: 20px;
  box-shadow: var(--shadow-card);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.metric-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-hover);
}

.metric-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.metric-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text-muted, #64748b);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.metric-icon {
  font-size: 18px;
}

.metric-body {
  margin-bottom: 8px;
}

.metric-value {
  font-size: 28px;
  font-weight: 800;
  color: var(--color-text, #162033);
  letter-spacing: -0.5px;
}

.metric-footer {
  font-size: 12px;
  color: var(--color-text-soft, #94a3b8);
}
</style>
