// Identity and the platform client, as the layout route every screen sits under.
// The callback screen needs the router above the session, so this is a route.

import { IamProvider, useIam } from '@hanzo/iam/react'
import { useEffect } from 'react'
import { Outlet } from 'react-router'

import { iam } from '~/api'
import { Client } from '~/client'
import { own, subject } from '~/token'

/** The IAM app is `hanzo-ads`; its org is the name's first word. */
export function origin() {
  const clientId = import.meta.env.VITE_HANZO_CLIENT_ID || 'hanzo-ads'
  return {
    serverUrl: iam(),
    clientId,
    redirectUri: `${window.location.origin}/auth/callback`,
    organization: clientId.split('-')[0]!,
  }
}

/** Once the session settles, the stored token's subject owns local state. */
function Identity() {
  const { isLoading, isAuthenticated } = useIam()
  useEffect(() => {
    if (!isLoading) own(subject())
  }, [isLoading, isAuthenticated])
  return null
}

export function Root() {
  // A browser another person left signed in forgets their organization first.
  own(subject())
  return (
    <IamProvider config={{ ...origin(), postLogoutRedirectUri: `${window.location.origin}/` }}>
      <Identity />
      <Client>
        <Outlet />
      </Client>
    </IamProvider>
  )
}
