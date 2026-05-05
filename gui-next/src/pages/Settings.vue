<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import Btn from '@/components/ui/Btn.vue'
import Field from '@/components/ui/Field.vue'
import ExtensionDownload from '@/components/ExtensionDownload.vue'

const auth = useAuthStore()
const newPwd = ref('')
const newPwd2 = ref('')
const error = ref<string | null>(null)
const ok = ref<string | null>(null)
const saving = ref(false)

async function changePwd(): Promise<void> {
  error.value = null
  ok.value = null
  if (newPwd.value.length < 8) {
    error.value = 'Password must be at least 8 characters.'
    return
  }
  if (newPwd.value !== newPwd2.value) {
    error.value = 'Passwords do not match.'
    return
  }
  saving.value = true
  try {
    await auth.changePassword(newPwd.value)
    ok.value = 'Password updated.'
    newPwd.value = ''
    newPwd2.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'change password failed'
  } finally {
    saving.value = false
  }
}

function downloadCA(): void {
  window.location.href = '/api/v1/download_ca'
}
</script>

<template>
  <div class="px-6 pt-5 pb-12 max-w-[720px] mx-auto space-y-6">
    <div>
      <h1 class="text-[20px] font-semibold tracking-tight">Settings</h1>
      <p class="text-[12px] text-fg-faint mt-0.5">
        Account preferences and proxy CA download.
      </p>
    </div>

    <section class="surface p-5 space-y-3">
      <h2 class="text-[13px] font-semibold">Account</h2>
      <p class="text-[12px] text-fg-muted">
        Signed in as <span class="mono">{{ auth.me?.username }}</span>.
        <span v-if="auth.mustChangePassword" class="ml-1 text-warn">
          (rotation required)
        </span>
      </p>

      <Field
        v-model="newPwd"
        label="New password"
        type="password"
        autocomplete="new-password"
        hint="At least 8 characters."
      />
      <Field
        v-model="newPwd2"
        label="Confirm new password"
        type="password"
        autocomplete="new-password"
      />

      <p v-if="error" class="text-danger text-[12px]">{{ error }}</p>
      <p v-if="ok" class="text-success text-[12px]">{{ ok }}</p>

      <Btn variant="primary" :loading="saving" @click="changePwd">
        Change password
      </Btn>
    </section>

    <section class="surface p-5 space-y-3">
      <h2 class="text-[13px] font-semibold">Proxy CA certificate</h2>
      <p class="text-[12px] text-fg-muted">
        Required to use the HTTPS proxy. Install the CA cert in your client's trust store.
      </p>
      <Btn variant="subtle" @click="downloadCA">Download rootCA.crt</Btn>
    </section>

    <ExtensionDownload />
  </div>
</template>
