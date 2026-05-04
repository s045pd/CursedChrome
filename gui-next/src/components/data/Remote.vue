<script setup lang="ts">
import { ref } from 'vue'
import { remote } from '@/api/endpoints'
import Btn from '@/components/ui/Btn.vue'
import Field from '@/components/ui/Field.vue'

const props = defineProps<{ botId: string; userAgent: string }>()
const url = ref('https://example.com')
const lastResult = ref<string | null>(null)
const error = ref<string | null>(null)
const sending = ref(false)

async function go(): Promise<void> {
  if (!url.value) return
  sending.value = true
  error.value = null
  lastResult.value = null
  try {
    const r = await remote.navigate(props.botId, url.value)
    lastResult.value = JSON.stringify(r, null, 2)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'navigate failed'
  } finally {
    sending.value = false
  }
}

async function stop(): Promise<void> {
  await remote.stop(props.botId)
  lastResult.value = '{ "stopped": true }'
}
</script>

<template>
  <div class="space-y-3">
    <div class="surface p-3">
      <div class="text-[11px] uppercase tracking-wider text-fg-faint mb-1">User-Agent</div>
      <pre class="mono text-[11px] whitespace-pre-wrap break-all leading-relaxed">{{ userAgent || '(unknown)' }}</pre>
    </div>

    <div class="surface p-3 space-y-3">
      <div class="text-[12px] font-medium">Navigate the bot's active tab</div>
      <Field
        v-model="url"
        label="URL"
        placeholder="https://example.com"
      />
      <div class="flex items-center gap-2">
        <Btn variant="primary" :loading="sending" @click="go">Navigate</Btn>
        <Btn variant="subtle" @click="stop">Stop navigate</Btn>
      </div>
      <p v-if="error" class="text-danger text-[12px] mono">{{ error }}</p>
      <pre
        v-if="lastResult"
        class="mono text-[11px] bg-bg-base border border-border-subtle p-2 rounded whitespace-pre-wrap break-all"
      >{{ lastResult }}</pre>
    </div>
  </div>
</template>
