<script setup lang="ts">
import { computed } from 'vue'
import type { BotSummary } from '@/types/api'
import { bots as botsApi } from '@/api/endpoints'
import StatusDot from '@/components/ui/StatusDot.vue'
import { timeAgo } from '@/composables/useTime'
import { useClipboard } from '@/composables/useClipboard'

interface Props {
  bot: BotSummary
  selected: boolean
  globalProxy: boolean
}
const props = defineProps<Props>()
defineEmits<{
  toggle: [id: string]
  open: [id: string]
  setProxy: [id: string]
  delete: [id: string]
}>()

const { copy, copied } = useClipboard()
const imageURL = computed(() => botsApi.imageURL(props.bot.id))
const tabTitle = computed(
  () => props.bot.current_tab?.title ?? props.bot.current_tab?.url ?? '—',
)
const tabHref = computed(() => props.bot.current_tab?.url ?? null)
</script>

<template>
  <tr
    class="group border-b border-border-subtle hover:bg-bg-hover/60 transition-colors"
    :class="selected ? 'bg-accent-soft/30' : ''"
  >
    <!-- selection -->
    <td class="px-3 py-2.5 align-middle w-8">
      <input
        type="checkbox"
        :checked="selected"
        class="size-3.5 accent-accent cursor-pointer"
        @click.stop="$emit('toggle', bot.id)"
      />
    </td>

    <!-- thumbnail -->
    <td class="px-2 py-2 align-middle w-[88px]">
      <div
        class="w-[80px] h-[44px] rounded overflow-hidden border border-border-subtle bg-bg-base relative cursor-pointer"
        @click="$emit('open', bot.id)"
      >
        <img
          v-if="bot.current_tab_image"
          :src="imageURL"
          alt=""
          loading="lazy"
          class="w-full h-full object-cover"
        />
        <div
          v-else
          class="w-full h-full grid place-items-center text-fg-faint text-[10px] mono"
        >
          no img
        </div>
      </div>
    </td>

    <!-- name + browser id -->
    <td class="px-3 py-2 align-middle min-w-[180px]">
      <button
        class="text-left block w-full hover:text-accent transition-colors"
        @click="$emit('open', bot.id)"
      >
        <div class="font-medium text-[13px] truncate">{{ bot.name || 'Untitled' }}</div>
        <div
          class="mono text-[10px] text-fg-faint truncate cursor-pointer"
          :title="bot.browser_id"
          @click.stop="copy(bot.browser_id, 'browser_id')"
        >
          {{ bot.browser_id.slice(0, 8) }}…{{ bot.browser_id.slice(-4) }}
          <span v-if="copied === 'browser_id'" class="text-success ml-1">copied</span>
        </div>
      </button>
    </td>

    <!-- online -->
    <td class="px-3 py-2 align-middle w-[120px]">
      <StatusDot :online="bot.is_online" />
    </td>

    <!-- current tab -->
    <td class="px-3 py-2 align-middle min-w-[260px] max-w-[320px]">
      <a
        v-if="tabHref"
        :href="tabHref"
        target="_blank"
        rel="noopener"
        class="text-[12px] truncate block hover:text-accent transition-colors"
        :title="tabHref"
      >
        {{ tabTitle }}
      </a>
      <span v-else class="text-fg-faint text-[12px]">—</span>
    </td>

    <!-- counts -->
    <td class="px-3 py-2 align-middle w-[100px] mono text-[12px]">
      <span class="text-fg-base">{{ bot.tabs }}</span>
      <span class="text-fg-faint mx-1">/</span>
      <span class="text-fg-faint">{{ bot.history }}</span>
    </td>

    <!-- proxy creds -->
    <td class="px-3 py-2 align-middle w-[200px]">
      <div class="space-y-0.5">
        <button
          class="block mono text-[11px] text-fg-base hover:text-accent text-left truncate w-full"
          :title="bot.proxy_username"
          @click.stop="copy(bot.proxy_username, 'user-' + bot.id)"
        >
          {{ bot.proxy_username }}
          <span v-if="copied === 'user-' + bot.id" class="text-success ml-1">✓</span>
        </button>
        <button
          class="block mono text-[11px] text-fg-faint hover:text-accent text-left truncate w-full"
          :title="bot.proxy_password"
          @click.stop="copy(bot.proxy_password, 'pwd-' + bot.id)"
        >
          •••• {{ bot.proxy_password.slice(-4) }}
          <span v-if="copied === 'pwd-' + bot.id" class="text-success ml-1">✓</span>
        </button>
      </div>
    </td>

    <!-- last active -->
    <td class="px-3 py-2 align-middle w-[120px] mono text-[11px] text-fg-muted">
      {{ timeAgo(bot.last_active_at ?? bot.last_online) }}
    </td>

    <!-- actions -->
    <td class="px-3 py-2 align-middle w-[180px]">
      <div class="flex items-center gap-1 opacity-60 group-hover:opacity-100 transition-opacity">
        <button
          class="h-7 px-2 text-[11px] rounded border border-border-subtle hover:border-border-strong hover:bg-bg-hover transition-colors"
          @click="$emit('open', bot.id)"
        >
          Open
        </button>
        <button
          class="h-7 px-2 text-[11px] rounded transition-colors"
          :class="
            globalProxy
              ? 'bg-accent-soft text-accent border border-accent/30'
              : 'border border-border-subtle hover:border-accent hover:text-accent'
          "
          :title="globalProxy ? 'Default proxy bot' : 'Use as default proxy'"
          @click="$emit('setProxy', bot.id)"
        >
          {{ globalProxy ? '★' : '☆' }}
        </button>
        <button
          class="h-7 px-2 text-[11px] rounded border border-border-subtle text-fg-faint hover:border-danger hover:text-danger transition-colors"
          @click="$emit('delete', bot.id)"
        >
          Delete
        </button>
      </div>
    </td>
  </tr>
</template>
