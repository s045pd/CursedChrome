<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import Btn from '@/components/ui/Btn.vue'
import Field from '@/components/ui/Field.vue'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const username = ref('')
const password = ref('')
const localError = ref<string | null>(null)

async function submit(): Promise<void> {
  localError.value = null
  if (!username.value || !password.value) {
    localError.value = 'username and password required'
    return
  }
  try {
    await auth.login(username.value.trim(), password.value)
    const target = (route.query.redirect as string) || '/'
    await router.replace(target)
  } catch (e) {
    localError.value = e instanceof Error ? e.message : 'login failed'
  }
}
</script>

<template>
  <div class="min-h-screen grid place-items-center bg-bg-base px-6">
    <!-- subtle grid background -->
    <div
      aria-hidden="true"
      class="pointer-events-none fixed inset-0 opacity-[0.04] [background-image:linear-gradient(var(--color-fg-base)_1px,transparent_1px),linear-gradient(90deg,var(--color-fg-base)_1px,transparent_1px)] [background-size:32px_32px]"
    />

    <div class="relative w-full max-w-[380px] surface-raised p-8">
      <div class="flex items-center gap-2.5 mb-7">
        <div
          class="size-8 rounded-md bg-accent-soft border border-accent/30 grid place-items-center"
        >
          <span class="text-accent font-bold text-[13px]">C</span>
        </div>
        <div>
          <h1 class="text-[15px] font-semibold tracking-tight">CursedChrome</h1>
          <p class="text-[11px] text-fg-faint mono">edr.console</p>
        </div>
      </div>

      <form class="space-y-3.5" @submit.prevent="submit">
        <Field
          v-model="username"
          label="Username"
          autocomplete="username"
          autofocus
          required
        />
        <Field
          v-model="password"
          label="Password"
          type="password"
          autocomplete="current-password"
          required
        />

        <p
          v-if="localError || auth.error"
          class="text-[12px] text-danger bg-danger-soft border border-danger/30 rounded px-3 py-2 mono"
        >
          {{ localError ?? auth.error }}
        </p>

        <Btn
          type="submit"
          variant="primary"
          size="lg"
          block
          :loading="auth.loading"
        >
          Sign in
        </Btn>
      </form>

      <p class="mt-6 text-[11px] text-fg-faint mono leading-relaxed">
        First-time login: use the credentials printed to the server console.
        You'll be prompted to rotate the password.
      </p>
    </div>
  </div>
</template>
