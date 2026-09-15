// One platform client, scoped to one organization. Signed out there is none:
// every /v1/ad route refuses an anonymous caller, so screens show sign-in instead.

import { createContext, useContext, useMemo, useState, type ReactNode } from 'react'
import { useIam } from '@hanzo/iam/react'
import { AiProvider } from '@hanzo/ai/react'
import { createAiClient, type HttpClient } from '@hanzo/ai'
import { api } from '~/api'
import { bearer, hasSession, org, orgs, work } from '~/token'

export interface Tenant {
  /** The org every call is scoped to; null when several and none picked. */
  org: string | null
  orgs: string[]
  choose: (org: string) => void
}

const Who = createContext<Tenant>({ org: null, orgs: [], choose: () => {} })
const Transport = createContext<HttpClient | null>(null)

export function useHttp(): HttpClient | null {
  return useContext(Transport)
}

export function useTenant(): Tenant {
  return useContext(Who)
}

export function Client({ children }: { children: ReactNode }) {
  const { sdk, isAuthenticated } = useIam()
  // State as well as storage: storage events reach other tabs only.
  const [picked, setPicked] = useState<string | null>(() => org())
  const mine = useMemo(() => orgs(), [isAuthenticated, picked])
  const account = Boolean((isAuthenticated || hasSession()) && sdk)

  const client = useMemo(
    () =>
      account && sdk
        ? createAiClient({
            baseUrl: api(),
            headers: picked ? { 'X-Org-Id': picked } : {},
            auth: {
              ...sdk,
              getValidAccessToken: async () => (await sdk.getValidAccessToken?.()) || bearer(),
            },
          })
        : null,
    [account, sdk, picked],
  )

  const tenant = useMemo<Tenant>(
    () => ({
      org: picked,
      orgs: mine,
      choose: (pick) => {
        work(pick)
        setPicked(pick)
      },
    }),
    [picked, mine],
  )

  return (
    <Who.Provider value={tenant}>
      <Transport.Provider value={client ? client.http : null}>
        {client ? <AiProvider client={client}>{children}</AiProvider> : children}
      </Transport.Provider>
    </Who.Provider>
  )
}
