// The session @hanzo/iam keeps in localStorage, read in one place. Storage can
// throw, so every access is guarded.

const PREFIX = 'hanzo_iam_'
const ACCESS = `${PREFIX}access_token`
/** The SDK's org selection, shared with every Hanzo surface. */
const CURRENT = `${PREFIX}current_org`
/** Whose browser state this is. */
const WHO = 'hanzo:who'
/** Per-person selections; the rest of the SDK's keys are the session itself. */
const CHOSEN = [CURRENT, `${PREFIX}current_project`]

const kept = (key: string): boolean => key === WHO || (key.startsWith(PREFIX) && !CHOSEN.includes(key))

export function bearer(): string | null {
  if (typeof window === 'undefined') return null
  try {
    return window.localStorage.getItem(ACCESS)
  } catch {
    return null
  }
}

function claims(): Record<string, unknown> | null {
  const part = bearer()?.split('.')[1]
  if (!part) return null
  try {
    return JSON.parse(atob(part.replace(/-/g, '+').replace(/_/g, '/'))) as Record<string, unknown>
  } catch {
    return null
  }
}

/** The stored token's subject. */
export function subject(): string | undefined {
  const sub = claims()?.sub
  return typeof sub === 'string' && sub ? sub : undefined
}

export function hasSession(): boolean {
  if (typeof window === 'undefined') return false
  try {
    return Boolean(bearer() || window.localStorage.getItem(WHO))
  } catch {
    return false
  }
}

/** When a different person (or nobody) arrives, drop the last person's selections. */
export function own(sub: string | undefined): void {
  if (typeof window === 'undefined') return
  try {
    const store = window.localStorage
    if (store.getItem(WHO) === (sub ?? null)) return
    for (const key of Object.keys(store)) if (key.startsWith('hanzo') && !kept(key)) store.removeItem(key)
    if (sub) store.setItem(WHO, sub)
    else store.removeItem(WHO)
  } catch {
    /* nothing stored, nothing to drop */
  }
}

/** The organizations the token's `orgs` claim names, in order. */
export function orgs(): string[] {
  const set = claims()?.orgs
  const out: string[] = []
  for (const ref of Array.isArray(set) ? set : []) {
    const o = (ref as { org?: unknown } | null)?.org
    if (typeof o === 'string' && o && !out.includes(o)) out.push(o)
  }
  return out
}

/** The stored selection when the token carries it; a sole org otherwise. */
export function org(): string | null {
  try {
    const mine = orgs()
    const chosen = window.localStorage.getItem(CURRENT)
    if (chosen && mine.includes(chosen)) return chosen
    return mine.length === 1 ? mine[0]! : null
  } catch {
    return null
  }
}

/** Record the org to work in. One the token does not carry is ignored. */
export function work(pick: string): void {
  try {
    if (orgs().includes(pick)) window.localStorage.setItem(CURRENT, pick)
  } catch {
    /* keeps the sole org */
  }
}
