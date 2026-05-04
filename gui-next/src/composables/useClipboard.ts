import { ref } from 'vue'

// Clipboard helper that gracefully falls back to a textarea on
// non-secure contexts (the original GUI hit this exact bug).
export function useClipboard() {
  const copied = ref<string | null>(null)

  async function copy(text: string, label = ''): Promise<void> {
    try {
      if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(text)
      } else {
        const ta = document.createElement('textarea')
        ta.value = text
        ta.style.position = 'fixed'
        ta.style.opacity = '0'
        document.body.appendChild(ta)
        ta.select()
        document.execCommand('copy')
        document.body.removeChild(ta)
      }
      copied.value = label || text
      window.setTimeout(() => {
        if (copied.value === (label || text)) copied.value = null
      }, 1200)
    } catch {
      copied.value = null
    }
  }

  return { copied, copy }
}
