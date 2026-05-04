<script setup lang="ts">
import { computed, useId } from 'vue'

interface Props {
  modelValue?: string | number
  label?: string
  type?: string
  placeholder?: string
  required?: boolean
  autocomplete?: string
  autofocus?: boolean
  hint?: string
  error?: string
  block?: boolean
  size?: 'sm' | 'md' | 'lg'
}
const props = withDefaults(defineProps<Props>(), {
  type: 'text',
  size: 'md',
  block: true,
})
defineEmits<{ 'update:modelValue': [v: string] }>()
const id = useId()
const sizeClass = computed(
  () =>
    ({
      sm: 'h-7 text-[11px] px-2',
      md: 'h-9 text-[13px] px-3',
      lg: 'h-11 text-[14px] px-3.5',
    })[props.size],
)
</script>

<template>
  <label class="block" :for="id">
    <span
      v-if="props.label"
      class="block text-[11px] uppercase tracking-wider text-fg-muted mb-1.5"
    >
      {{ props.label }}
      <span v-if="props.required" class="text-danger ml-0.5">*</span>
    </span>
    <input
      :id="id"
      :type="props.type"
      :placeholder="props.placeholder"
      :required="props.required"
      :autocomplete="props.autocomplete"
      :autofocus="props.autofocus"
      :value="props.modelValue"
      @input="
        (e) => $emit('update:modelValue', (e.target as HTMLInputElement).value)
      "
      class="bg-bg-base border border-border-subtle text-fg-base rounded-md focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/40 placeholder:text-fg-faint transition-colors"
      :class="[
        sizeClass,
        props.block ? 'w-full' : '',
        props.error ? '!border-danger !ring-danger/40' : '',
      ]"
    />
    <p v-if="props.error" class="text-danger text-[11px] mt-1">{{ props.error }}</p>
    <p v-else-if="props.hint" class="text-fg-faint text-[11px] mt-1">{{ props.hint }}</p>
  </label>
</template>
