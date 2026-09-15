// The ad accounts a campaign launches through. Consent happens on the network's
// own site; the token it grants stays in KMS.

import { useState } from 'react'
import { Text, XStack, YStack } from '@hanzo/ui'

import { Act, Failed, Nothing, Ready, Screen } from '~/page'
import { Mark } from '~/status'
import { connect, useConnections, why, type Connection } from '~/ads'
import { useHttp } from '~/client'

function One({ it }: { it: Connection }) {
  const http = useHttp()
  const [busy, setBusy] = useState(false)
  const [failed, setFailed] = useState<string | null>(null)

  const start = async () => {
    if (!http || busy) return
    setBusy(true)
    setFailed(null)
    try {
      const { authorizeUrl } = await connect(http, it.id)
      if (authorizeUrl) window.location.assign(authorizeUrl)
      else setFailed('The network gave no sign-in address.')
    } catch (e) {
      setFailed(why(e))
    } finally {
      setBusy(false)
    }
  }

  return (
    <YStack gap="$2" p="$4" rounded="$3" borderWidth={1} borderColor="$borderColor">
      <XStack items="center" gap="$3">
        <Text fontSize="$3" fontWeight="500" color="$ink" flex={1}>
          {it.name}
        </Text>
        <Mark of={it.connected ? 'up' : 'quiet'} says={it.connected ? 'connected' : 'not connected'} />
      </XStack>
      <Text fontSize="$1" color="$soft">
        {it.description}
      </Text>
      {it.connected && it.connection ? (
        <Text fontSize="$1" color="$quiet">
          {it.connection.account || 'Account connected'}
          {it.connection.scopes.length ? ` · ${it.connection.scopes.join(', ')}` : ''}
        </Text>
      ) : it.available ? (
        <XStack>
          <Act onPress={() => void start()} disabled={busy}>
            {busy ? 'Opening…' : `Connect ${it.name}`}
          </Act>
        </XStack>
      ) : (
        <Text fontSize="$1" color="$quiet">
          Connecting {it.name} is not available here yet.
        </Text>
      )}
      {failed ? (
        <Text fontSize="$1" color="$ink" fontWeight="500">
          {failed}
        </Text>
      ) : null}
    </YStack>
  )
}

export function Accounts() {
  const http = useHttp()
  const connections = useConnections(http)
  const rows = connections.it ?? []

  return (
    <Ready>
      <Screen title="Ad accounts" says="Connect a network's ad account once. Every campaign on that network launches through it.">
        {connections.failed ? <Failed what="the ad accounts" why={connections.failed} /> : null}
        {connections.it && rows.length === 0 ? (
          <Nothing says="No ad account connectors are offered here." />
        ) : (
          <YStack gap="$3">
            {rows.map((it) => (
              <One key={it.id} it={it} />
            ))}
          </YStack>
        )}
        <Text fontSize="$1" color="$quiet">
          X campaigns have no ad account to connect.
        </Text>
      </Screen>
    </Ready>
  )
}
