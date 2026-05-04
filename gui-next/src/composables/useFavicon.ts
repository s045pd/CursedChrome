export function faviconUrl(pageUrl: string | undefined | null, size = 16): string {
  if (!pageUrl) return ''
  try {
    const { origin } = new URL(pageUrl)
    return `https://www.google.com/s2/favicons?domain=${encodeURIComponent(origin)}&sz=${size}`
  } catch {
    return ''
  }
}
