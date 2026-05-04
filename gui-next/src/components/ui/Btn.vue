<script setup lang="ts">
// Single button component. Variant + size knobs cover every place
// in the app — no rogue className strings.
interface Props {
  variant?: 'primary' | 'ghost' | 'subtle' | 'danger' | 'success'
  size?: 'sm' | 'md' | 'lg'
  loading?: boolean
  disabled?: boolean
  block?: boolean
  type?: 'button' | 'submit' | 'reset'
}
const props = withDefaults(defineProps<Props>(), {
  variant: 'subtle',
  size: 'md',
  loading: false,
  disabled: false,
  block: false,
  type: 'button',
})

const variantClass = {
  primary:
    'bg-accent text-bg-base hover:bg-accent-strong active:bg-accent disabled:bg-accent/40',
  ghost:
    'bg-transparent text-fg-muted hover:bg-bg-hover hover:text-fg-base disabled:text-fg-faint',
  subtle:
    'bg-bg-overlay text-fg-base border border-border-subtle hover:border-border-strong hover:bg-bg-hover',
  danger:
    'bg-danger-soft text-danger border border-danger/30 hover:bg-danger/20',
  success:
    'bg-success-soft text-success border border-success/30 hover:bg-success/20',
}
const sizeClass = {
  sm: 'h-7 px-2.5 text-[11px] gap-1.5 rounded',
  md: 'h-8 px-3 text-[12px] gap-2 rounded',
  lg: 'h-10 px-4 text-[14px] gap-2 rounded-md',
}
</script>

<template>
  <button
    :type="props.type"
    :disabled="props.disabled || props.loading"
    class="inline-flex items-center justify-center font-medium transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-accent/50 disabled:cursor-not-allowed disabled:opacity-60 select-none"
    :class="[
      variantClass[props.variant],
      sizeClass[props.size],
      props.block ? 'w-full' : '',
    ]"
  >
    <span
      v-if="props.loading"
      class="size-3 rounded-full border-2 border-current border-t-transparent animate-spin"
    />
    <slot />
  </button>
</template>
