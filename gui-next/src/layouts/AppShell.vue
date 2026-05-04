<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useBotsStore } from '@/stores/bots'

const auth = useAuthStore()
const bots = useBotsStore()
const router = useRouter()

const onlineLabel = computed(() => `${bots.onlineCount} live · ${bots.offlineCount} offline`)

async function logout(): Promise<void> {
  await auth.logout()
  await router.replace('/login')
}

function toggleTheme(): void {
  document.documentElement.classList.toggle('light')
}
</script>

<template>
  <div class="min-h-screen flex flex-col">
    <!-- top bar -->
    <header
      class="h-12 shrink-0 flex items-center px-4 border-b border-border-subtle bg-bg-base/90 backdrop-blur z-10 sticky top-0"
    >
      <RouterLink to="/" class="flex items-center gap-2 mr-6">
        <div
          class="size-6 rounded bg-accent-soft border border-accent/30 grid place-items-center"
        >
          <span class="text-accent font-bold text-[11px]">C</span>
        </div>
        <span class="font-semibold tracking-tight text-[13px]">CursedChrome</span>
        <span class="chip ml-1.5">edr</span>
      </RouterLink>

      <nav class="flex items-center gap-1 text-[12px]">
        <RouterLink
          v-slot="{ isActive }"
          to="/"
          custom
        >
          <RouterLink
            to="/"
            class="px-2.5 h-7 inline-flex items-center rounded transition-colors"
            :class="
              isActive
                ? 'bg-bg-overlay text-fg-base'
                : 'text-fg-muted hover:text-fg-base hover:bg-bg-hover'
            "
          >
            Bots
          </RouterLink>
        </RouterLink>
        <RouterLink
          v-slot="{ isActive }"
          to="/settings"
          custom
        >
          <RouterLink
            to="/settings"
            class="px-2.5 h-7 inline-flex items-center rounded transition-colors"
            :class="
              isActive
                ? 'bg-bg-overlay text-fg-base'
                : 'text-fg-muted hover:text-fg-base hover:bg-bg-hover'
            "
          >
            Settings
          </RouterLink>
        </RouterLink>
      </nav>

      <div class="ml-auto flex items-center gap-3">
        <span class="chip mono">
          <span class="live-dot" /> {{ onlineLabel }}
        </span>
        <button
          class="size-7 rounded text-fg-muted hover:text-fg-base hover:bg-bg-hover grid place-items-center"
          aria-label="Toggle theme"
          title="Toggle theme"
          @click="toggleTheme"
        >
          <svg viewBox="0 0 16 16" class="size-3.5" fill="currentColor">
            <path
              d="M8 1a7 7 0 1 0 7 7c0-.21-.01-.42-.03-.63A6 6 0 0 1 8.63 1.03 6.85 6.85 0 0 0 8 1Z"
            />
          </svg>
        </button>
        <div class="text-[12px] flex items-center gap-2">
          <span class="text-fg-faint mono">{{ auth.me?.username }}</span>
          <button
            class="text-fg-muted hover:text-fg-base text-[11px] uppercase tracking-wider"
            @click="logout"
          >
            Sign out
          </button>
        </div>
      </div>
    </header>

    <main class="flex-1 min-h-0">
      <RouterView />
    </main>
  </div>
</template>
