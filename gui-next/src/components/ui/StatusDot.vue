<script setup lang="ts">
defineProps<{ online: boolean; label?: string }>()
</script>

<template>
  <span class="inline-flex items-center gap-2 text-[12px] mono">
    <span :class="online ? 'live-dot' : 'offline-dot'" aria-hidden="true" />
    <span :class="online ? 'text-success' : 'text-fg-faint'">
      {{ label ?? (online ? 'online' : 'offline') }}
    </span>
  </span>
</template>
