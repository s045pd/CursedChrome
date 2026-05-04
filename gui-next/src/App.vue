<script setup lang="ts">
import { onMounted } from 'vue'
import { RouterView } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()

// Hydrate auth on first paint so refresh on a deep link doesn't bounce
// to login when a session cookie is already present.
onMounted(async () => {
  if (!auth.isAuthenticated) {
    try {
      await auth.refresh()
    } catch {
      /* surfaced by router guard */
    }
  }
})
</script>

<template>
  <RouterView />
</template>
