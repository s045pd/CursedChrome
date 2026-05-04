<script setup lang="ts">
import { onMounted, onBeforeUnmount, watch } from 'vue'

interface Props {
  open: boolean
  title?: string
  width?: 'md' | 'lg' | 'xl' | 'full'
}
const props = withDefaults(defineProps<Props>(), { width: 'lg' })
const emit = defineEmits<{ close: [] }>()

const widthClass = {
  md: 'w-[480px]',
  lg: 'w-[720px]',
  xl: 'w-[960px]',
  full: 'w-[min(1280px,100vw)]',
}

function onKey(e: KeyboardEvent): void {
  if (e.key === 'Escape' && props.open) emit('close')
}

onMounted(() => window.addEventListener('keydown', onKey))
onBeforeUnmount(() => window.removeEventListener('keydown', onKey))

watch(
  () => props.open,
  (v) => {
    document.body.style.overflow = v ? 'hidden' : ''
  },
)
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-opacity duration-150"
      leave-active-class="transition-opacity duration-150"
      enter-from-class="opacity-0"
      leave-to-class="opacity-0"
    >
      <div
        v-if="open"
        class="fixed inset-0 z-40 bg-black/60 backdrop-blur-[2px]"
        @click="emit('close')"
      />
    </Transition>
    <Transition
      enter-active-class="transition-transform duration-200"
      leave-active-class="transition-transform duration-150"
      enter-from-class="translate-x-full"
      leave-to-class="translate-x-full"
    >
      <aside
        v-if="open"
        class="fixed top-0 right-0 z-50 h-full bg-bg-raised border-l border-border-subtle shadow-pop flex flex-col"
        :class="widthClass[props.width]"
      >
        <header
          class="flex items-center justify-between h-12 px-4 border-b border-border-subtle shrink-0"
        >
          <h2 class="text-[13px] font-semibold tracking-wide truncate">
            <slot name="title">{{ props.title }}</slot>
          </h2>
          <button
            class="size-7 inline-flex items-center justify-center text-fg-muted hover:text-fg-base hover:bg-bg-hover rounded transition-colors"
            aria-label="Close drawer"
            @click="emit('close')"
          >
            <svg viewBox="0 0 16 16" class="size-3.5" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M3 3l10 10M13 3L3 13" />
            </svg>
          </button>
        </header>
        <div class="flex-1 overflow-auto">
          <slot />
        </div>
        <footer
          v-if="$slots.footer"
          class="border-t border-border-subtle px-4 py-3 shrink-0 flex items-center gap-2 justify-end"
        >
          <slot name="footer" />
        </footer>
      </aside>
    </Transition>
  </Teleport>
</template>
