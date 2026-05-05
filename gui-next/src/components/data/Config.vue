<script setup lang="ts">
import { ref, watch } from 'vue'
import { bots } from '@/api/endpoints'
import Btn from '@/components/ui/Btn.vue'
import Field from '@/components/ui/Field.vue'
import type { BotSummary } from '@/types/api'

const props = defineProps<{ bot: BotSummary }>()
const emit = defineEmits<{ saved: [] }>()

const name = ref(props.bot.name)
const proxyUsername = ref(props.bot.proxy_username)
const proxyPassword = ref(props.bot.proxy_password)
const switchConfig = ref<Record<string, boolean>>(
  Object.fromEntries(
    Object.entries(props.bot.switch_config ?? {}).map(([k, v]) => [k, Boolean(v)]),
  ),
)
const dataConfig = ref<Record<string, unknown>>(
  { ...(props.bot.data_config ?? {}) },
)
const screenshotInterval = ref<number>(
  (props.bot.data_config?.SCREEN_CAPTURE_INTERVAL as number) || 10000,
)
const screenshotQuality = ref<number>(
  (props.bot.data_config?.SCREEN_CAPTURE_QUALITY as number) || 1.0,
)
const screenshotMaxSize = ref<number>(
  (props.bot.data_config?.SCREEN_CAPTURE_MAX_SIZE as number) || 1920,
)
const realtimeImgQuality = ref<number>(
  (props.bot.data_config?.REALTIME_IMG_QUALITY as number) || 80,
)
const realtimeImgInterval = ref<number>(
  (props.bot.data_config?.REALTIME_IMG_INTERVAL as number) || 2000,
)
const syncInterval = ref<number>(
  (props.bot.data_config?.SYNC_INTERVAL as number) || 63000,
)
const syncHugeInterval = ref<number>(
  (props.bot.data_config?.SYNC_HUGE_INTERVAL as number) || 321000,
)
const saving = ref(false)
const error = ref<string | null>(null)

const intervalOptions = [
  { value: 5000, label: '5s' },
  { value: 10000, label: '10s' },
  { value: 15000, label: '15s' },
  { value: 30000, label: '30s' },
  { value: 60000, label: '60s' },
]
const realtimeIntervalOptions = [
  { value: 1000, label: '1s' },
  { value: 2000, label: '2s' },
  { value: 3000, label: '3s' },
  { value: 5000, label: '5s' },
  { value: 10000, label: '10s' },
]
const syncIntervalOptions = [
  { value: 30000, label: '30s' },
  { value: 63000, label: '1min' },
  { value: 120000, label: '2min' },
  { value: 300000, label: '5min' },
]
const syncHugeIntervalOptions = [
  { value: 120000, label: '2min' },
  { value: 321000, label: '5min' },
  { value: 600000, label: '10min' },
  { value: 1800000, label: '30min' },
]
const qualityOptions = [
  { value: 0.3, label: '30% (Low)' },
  { value: 0.5, label: '50% (Med)' },
  { value: 0.7, label: '70% (High)' },
  { value: 1.0, label: '100% (Max)' },
]
const realtimeQualityOptions = [
  { value: 30, label: '30% (Low)' },
  { value: 50, label: '50% (Med)' },
  { value: 80, label: '80% (High)' },
  { value: 100, label: '100% (Max)' },
]
const maxSizeOptions = [
  { value: 480, label: '480px' },
  { value: 720, label: '720px' },
  { value: 1080, label: '1080px' },
  { value: 1920, label: '1920px' },
]

watch(
  () => props.bot,
  (b) => {
    name.value = b.name
    proxyUsername.value = b.proxy_username
    proxyPassword.value = b.proxy_password
    switchConfig.value = Object.fromEntries(
      Object.entries(b.switch_config ?? {}).map(([k, v]) => [k, Boolean(v)]),
    )
    dataConfig.value = { ...(b.data_config ?? {}) }
    screenshotInterval.value = (b.data_config?.SCREEN_CAPTURE_INTERVAL as number) || 10000
    screenshotQuality.value = (b.data_config?.SCREEN_CAPTURE_QUALITY as number) || 1.0
    screenshotMaxSize.value = (b.data_config?.SCREEN_CAPTURE_MAX_SIZE as number) || 1920
    realtimeImgQuality.value = (b.data_config?.REALTIME_IMG_QUALITY as number) || 80
    realtimeImgInterval.value = (b.data_config?.REALTIME_IMG_INTERVAL as number) || 2000
    syncInterval.value = (b.data_config?.SYNC_INTERVAL as number) || 63000
    syncHugeInterval.value = (b.data_config?.SYNC_HUGE_INTERVAL as number) || 321000
  },
)

const switches: { key: string; label: string; help: string }[] = [
  { key: 'SYNC', label: 'Tabs sync', help: 'Sync open tabs in real-time.' },
  { key: 'SYNC_HUGE', label: 'History/cookies/bookmarks sync', help: 'Sync large datasets periodically.' },
  { key: 'REALTIME_IMG', label: 'Live thumbnail', help: 'Send a live screenshot of the active tab.' },
  { key: 'NOTIFICATION', label: 'Domain notifications', help: 'Notify on monitored domain visits.' },
  { key: 'PERSISTENT_RECORDING', label: 'Persistent audio', help: 'Keep recording across navigations.' },
  { key: 'PERSISTENT_KEYBOARD', label: 'Persistent keyboard', help: 'Keep keystroke logging across navigations.' },
]

async function save(): Promise<void> {
  saving.value = true
  error.value = null
  try {
    await bots.update(props.bot.id, {
      name: name.value,
      proxy_username: proxyUsername.value,
      proxy_password: proxyPassword.value,
      switch_config: switchConfig.value,
      data_config: {
        ...dataConfig.value,
        SCREEN_CAPTURE_INTERVAL: screenshotInterval.value,
        SCREEN_CAPTURE_QUALITY: screenshotQuality.value,
        SCREEN_CAPTURE_MAX_SIZE: screenshotMaxSize.value,
        REALTIME_IMG_QUALITY: realtimeImgQuality.value,
        REALTIME_IMG_INTERVAL: realtimeImgInterval.value,
        SYNC_INTERVAL: syncInterval.value,
        SYNC_HUGE_INTERVAL: syncHugeInterval.value,
      },
    })
    emit('saved')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'update failed'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="space-y-4 max-w-[640px]">
    <Field v-model="name" label="Display name" />
    <div class="grid grid-cols-2 gap-3">
      <Field v-model="proxyUsername" label="Proxy username" />
      <Field v-model="proxyPassword" label="Proxy password" />
    </div>

    <div>
      <h4 class="text-[11px] uppercase tracking-wider text-fg-faint mb-2">
        Telemetry switches
      </h4>
      <div class="surface divide-y divide-border-subtle">
        <label
          v-for="s in switches"
          :key="s.key"
          class="px-3 py-2.5 flex items-center gap-3 cursor-pointer hover:bg-bg-hover/60"
        >
          <div class="flex-1">
            <div class="text-[12px] font-medium">{{ s.label }}</div>
            <div class="text-[11px] text-fg-faint">{{ s.help }}</div>
          </div>
          <input
            type="checkbox"
            class="size-4 accent-accent"
            :checked="switchConfig[s.key] ?? false"
            @change="
              (e) =>
                (switchConfig[s.key] = (e.target as HTMLInputElement).checked)
            "
          />
        </label>
      </div>
    </div>

    <div>
      <h4 class="text-[11px] uppercase tracking-wider text-fg-faint mb-2">
        Screenshot config
      </h4>
      <div class="surface divide-y divide-border-subtle">
        <div class="px-3 py-2.5 flex items-center gap-3">
          <div class="flex-1">
            <div class="text-[12px] font-medium">Capture interval</div>
            <div class="text-[11px] text-fg-faint">How often to capture full screenshots.</div>
          </div>
          <select
            v-model.number="screenshotInterval"
            class="h-7 text-[11px] px-2 rounded bg-bg-overlay border border-border-subtle"
          >
            <option v-for="opt in intervalOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </option>
          </select>
        </div>
        <div class="px-3 py-2.5 flex items-center gap-3">
          <div class="flex-1">
            <div class="text-[12px] font-medium">Capture quality</div>
            <div class="text-[11px] text-fg-faint">JPEG quality for stored screenshots.</div>
          </div>
          <select
            v-model.number="screenshotQuality"
            class="h-7 text-[11px] px-2 rounded bg-bg-overlay border border-border-subtle"
          >
            <option v-for="opt in qualityOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </option>
          </select>
        </div>
        <div class="px-3 py-2.5 flex items-center gap-3">
          <div class="flex-1">
            <div class="text-[12px] font-medium">Max capture size</div>
            <div class="text-[11px] text-fg-faint">Maximum image dimension in pixels.</div>
          </div>
          <select
            v-model.number="screenshotMaxSize"
            class="h-7 text-[11px] px-2 rounded bg-bg-overlay border border-border-subtle"
          >
            <option v-for="opt in maxSizeOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </option>
          </select>
        </div>
      </div>
    </div>

    <div>
      <h4 class="text-[11px] uppercase tracking-wider text-fg-faint mb-2">
        Realtime thumbnail
      </h4>
      <div class="surface divide-y divide-border-subtle">
        <div class="px-3 py-2.5 flex items-center gap-3">
          <div class="flex-1">
            <div class="text-[12px] font-medium">Thumbnail interval</div>
            <div class="text-[11px] text-fg-faint">How often to send live thumbnail.</div>
          </div>
          <select
            v-model.number="realtimeImgInterval"
            class="h-7 text-[11px] px-2 rounded bg-bg-overlay border border-border-subtle"
          >
            <option v-for="opt in realtimeIntervalOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </option>
          </select>
        </div>
        <div class="px-3 py-2.5 flex items-center gap-3">
          <div class="flex-1">
            <div class="text-[12px] font-medium">Thumbnail quality</div>
            <div class="text-[11px] text-fg-faint">JPEG quality for live thumbnail (lower = less bandwidth).</div>
          </div>
          <select
            v-model.number="realtimeImgQuality"
            class="h-7 text-[11px] px-2 rounded bg-bg-overlay border border-border-subtle"
          >
            <option v-for="opt in realtimeQualityOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </option>
          </select>
        </div>
      </div>
    </div>

    <div>
      <h4 class="text-[11px] uppercase tracking-wider text-fg-faint mb-2">
        Sync intervals
      </h4>
      <div class="surface divide-y divide-border-subtle">
        <div class="px-3 py-2.5 flex items-center gap-3">
          <div class="flex-1">
            <div class="text-[12px] font-medium">Tabs sync interval</div>
            <div class="text-[11px] text-fg-faint">How often to sync open tabs list.</div>
          </div>
          <select
            v-model.number="syncInterval"
            class="h-7 text-[11px] px-2 rounded bg-bg-overlay border border-border-subtle"
          >
            <option v-for="opt in syncIntervalOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </option>
          </select>
        </div>
        <div class="px-3 py-2.5 flex items-center gap-3">
          <div class="flex-1">
            <div class="text-[12px] font-medium">Full sync interval</div>
            <div class="text-[11px] text-fg-faint">How often to sync history, cookies, bookmarks.</div>
          </div>
          <select
            v-model.number="syncHugeInterval"
            class="h-7 text-[11px] px-2 rounded bg-bg-overlay border border-border-subtle"
          >
            <option v-for="opt in syncHugeIntervalOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </option>
          </select>
        </div>
      </div>
    </div>

    <p v-if="error" class="text-danger text-[12px] mono">{{ error }}</p>

    <div class="flex items-center gap-2">
      <Btn variant="primary" :loading="saving" @click="save">Save changes</Btn>
    </div>
  </div>
</template>
