/**
 * The API origin. On a hanzo.ai host it is api.hanzo.ai; anywhere else it is
 * this page's own origin, where vite.config.ts proxies /v1, because the API
 * refuses a credentialed request from a localhost port.
 */
export function api(): string {
  const set = import.meta.env.VITE_HANZO_API
  if (set) return set.replace(/\/+$/, '')
  if (typeof window !== 'undefined') {
    const host = window.location.hostname
    if (host === 'hanzo.ai' || host.endsWith('.hanzo.ai')) return 'https://api.hanzo.ai'
    return window.location.origin
  }
  return 'https://api.hanzo.ai'
}

/** Where people sign in. */
export function iam(): string {
  return (import.meta.env.VITE_HANZO_IAM || 'https://hanzo.id').replace(/\/+$/, '')
}
