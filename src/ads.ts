// Every /v1 address this app calls, written once.
//
// Shapes follow the platform's own types: AdCampaign and adSummary in
// hanzoai/cloud apps/ad, providerView in apps/integrations. Money is in cents.
// No ad network token ever reaches the browser; launch reads it server-side.

import { useCallback, useEffect, useState } from 'react'
import type { HttpClient } from '@hanzo/ai'

/** One campaign, as the platform stores it. Times are unix seconds. */
export interface Campaign {
  id: string
  name: string
  /** meta, google, tiktok or x. */
  platform: string
  /** The network's ad account, e.g. Meta's act_<id>. */
  account?: string
  /** The network's own campaign id. Set only by a successful launch. */
  externalId?: string
  /** draft, active, paused or completed. */
  status: string
  objective: string
  budget: number
  /** Spend as entered here. Not read back from the network. */
  spend: number
  createdAt: number
  updatedAt: number
}

/** The org's roll-up over every campaign. */
export interface Summary {
  campaigns: number
  active: number
  budget: number
  spend: number
}

/** A connector, and whether this org has connected it. */
export interface Connection {
  id: string
  name: string
  description: string
  category: string
  /** The deployment can start a connection. */
  available: boolean
  connected: boolean
  connection?: {
    account: string
    externalId: string
    scopes: string[]
    connectedAt: string
  }
}

/** What a create or an update sends. An update replaces every field. */
export interface Plan {
  name: string
  platform: string
  account: string
  status: string
  objective: string
  budget: number
  spend: number
}

export const PLATFORMS = ['meta', 'google', 'tiktok', 'x'] as const
export const STATUSES = ['draft', 'active', 'paused', 'completed'] as const

const NAMES: Record<string, string> = { meta: 'Meta', google: 'Google', tiktok: 'TikTok', x: 'X' }

/** A network's name on screen. */
export const named = (platform: string): string => NAMES[platform] ?? platform

/** The ad account connector a network launches through. X has none. */
export const CONNECTOR: Record<string, string> = {
  meta: 'meta_ads',
  google: 'google_ads',
  tiktok: 'tiktok_ads',
}

/** A stored campaign as a plan, for a full-replace update. */
export const plan = (c: Campaign): Plan => ({
  name: c.name,
  platform: c.platform,
  account: c.account ?? '',
  status: c.status,
  objective: c.objective,
  budget: c.budget,
  spend: c.spend,
})

/** The platform's refusal sentence (RFC 7807 `detail`), else the error's message. */
export function why(e: unknown): string {
  const detail = (e as { body?: { detail?: unknown } } | null)?.body?.detail
  if (typeof detail === 'string' && detail.trim()) return detail
  return e instanceof Error ? e.message : String(e)
}

export interface Read<T> {
  it: T | null
  failed: string | null
  again: () => void
}

/** One read with its failure. Re-reads when the client (the org) changes. */
function useRead<T>(http: HttpClient | null, get: (h: HttpClient) => Promise<T>, deps: unknown[]): Read<T> {
  const [it, setIt] = useState<T | null>(null)
  const [failed, setFailed] = useState<string | null>(null)
  const [turn, setTurn] = useState(0)
  // eslint-disable-next-line react-hooks/exhaustive-deps
  const fetchIt = useCallback(get, deps)

  useEffect(() => {
    if (!http) return
    let live = true
    setFailed(null)
    fetchIt(http)
      .then((got) => live && setIt(got))
      .catch((e: unknown) => live && setFailed(why(e)))
    return () => {
      live = false
    }
  }, [http, fetchIt, turn])

  return { it, failed, again: () => setTurn((t) => t + 1) }
}

const at = (id: string) => `/v1/ad/campaigns/${encodeURIComponent(id)}`

/** GET /v1/ad/summary */
export function useSummary(http: HttpClient | null): Read<Summary> {
  return useRead(http, (h) => h.json<Summary>({ path: '/v1/ad/summary' }), [])
}

/** GET /v1/ad/campaigns, newest-changed first, optionally one status. */
export function useCampaigns(http: HttpClient | null, status?: string): Read<Campaign[]> {
  return useRead(
    http,
    (h) =>
      h.collection<Campaign>('data', {
        path: '/v1/ad/campaigns',
        query: { limit: 1000, ...(status ? { status } : {}) },
      }),
    [status],
  )
}

/** GET /v1/ad/campaigns/{id} */
export function useCampaign(http: HttpClient | null, id: string): Read<Campaign> {
  return useRead(http, (h) => h.json<Campaign>({ path: at(id) }), [id])
}

/** GET /v1/integration, narrowed to the ad account connectors. */
export function useConnections(http: HttpClient | null): Read<Connection[]> {
  return useRead(
    http,
    async (h) => {
      const ads = new Set(Object.values(CONNECTOR))
      const all = await h.collection<Connection>('providers', { path: '/v1/integration' })
      return all.filter((c) => ads.has(c.id))
    },
    [],
  )
}

/** POST /v1/ad/campaigns. Stores it; nothing reaches the network. */
export function create(http: HttpClient, it: Plan): Promise<Campaign> {
  return http.json<Campaign>({ method: 'POST', path: '/v1/ad/campaigns', body: it })
}

/** PUT /v1/ad/campaigns/{id}. Replaces every field. */
export function update(http: HttpClient, id: string, it: Plan): Promise<Campaign> {
  return http.json<Campaign>({ method: 'PUT', path: at(id), body: it })
}

/** DELETE /v1/ad/campaigns/{id}. Answers 204 with no body. */
export async function drop(http: HttpClient, id: string): Promise<void> {
  await http.raw({ method: 'DELETE', path: at(id) })
}

/** POST /v1/ad/campaigns/{id}/launch. Creates it on its network. */
export function launch(http: HttpClient, id: string): Promise<Campaign> {
  return http.json<Campaign>({ method: 'POST', path: `${at(id)}/launch`, body: {} })
}

/** POST /v1/integration/{provider}/connect. Answers the network's consent address. */
export function connect(http: HttpClient, provider: string): Promise<{ authorizeUrl?: string }> {
  return http.json({ method: 'POST', path: `/v1/integration/${encodeURIComponent(provider)}/connect`, body: {} })
}
