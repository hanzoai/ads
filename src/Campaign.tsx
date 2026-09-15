// One campaign: launch it, pause it, change it, or delete it.

import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router'
import { Text, XStack, YStack } from '@hanzo/ui'

import { Act, Failed, Ready, Screen } from '~/page'
import { Mark, tone } from '~/status'
import { Fields, toDraft, toPlan, type Draft } from '~/form'
import {
  CONNECTOR,
  drop,
  launch,
  named,
  plan,
  update,
  useCampaign,
  useConnections,
  why,
  type Campaign as Row,
} from '~/ads'
import { money, since } from '~/units'
import { useHttp } from '~/client'

function Fact({ label, children }: { label: string; children: string }) {
  return (
    <YStack gap="$1" minW={160} flex={1}>
      <Text fontSize="$1" color="$quiet">
        {label}
      </Text>
      <Text fontSize="$2" color="$ink">
        {children}
      </Text>
    </YStack>
  )
}

function Body({ c, again }: { c: Row; again: () => void }) {
  const http = useHttp()
  const go = useNavigate()
  const connections = useConnections(http)
  const [draft, setDraft] = useState<Draft>(() => toDraft(plan(c)))
  const [busy, setBusy] = useState('')
  const [failed, setFailed] = useState<string | null>(null)
  const [sure, setSure] = useState(false)

  useEffect(() => setDraft(toDraft(plan(c))), [c])

  const run = async (what: string, act: () => Promise<unknown>) => {
    if (!http || busy) return
    setBusy(what)
    setFailed(null)
    try {
      await act()
      if (what === 'drop') return go('/')
      again()
    } catch (e) {
      setFailed(why(e))
    } finally {
      setBusy('')
    }
  }

  const connector = CONNECTOR[c.platform]
  const link = (connections.it ?? []).find((one) => one.id === connector)
  const launched = Boolean(c.externalId)

  return (
    <>
      <XStack gap="$4" flexWrap="wrap" p="$4" rounded="$3" borderWidth={1} borderColor="$borderColor">
        <Fact label="Budget">{money(c.budget)}</Fact>
        <Fact label="Spent so far">{money(c.spend)}</Fact>
        <Fact label="Objective">{c.objective || 'Network default'}</Fact>
        <Fact label="Ad account">{c.account || 'None yet'}</Fact>
        <Fact label="On the network">{c.externalId || 'Not launched'}</Fact>
        <Fact label="Changed">{since(c.updatedAt)}</Fact>
      </XStack>

      <YStack gap="$2">
        <Text fontSize="$2" color="$soft">
          {!connector
            ? `${named(c.platform)} has no ad account to connect, so this campaign cannot launch.`
            : link?.connected
              ? `${named(c.platform)} ad account connected${link.connection?.account ? `: ${link.connection.account}` : ''}.`
              : `No ${named(c.platform)} ad account is connected.`}
          {connector && !link?.connected ? (
            <>
              {' '}
              <Link to="/accounts" style={{ color: 'inherit', textDecoration: 'underline' }}>
                Connect one.
              </Link>
            </>
          ) : null}
        </Text>
        {connections.failed ? <Failed what="the ad accounts" why={connections.failed} /> : null}

        <XStack gap="$2" flexWrap="wrap" items="center">
          {launched ? null : (
            <Act onPress={() => void run('launch', () => launch(http!, c.id))} disabled={busy !== ''} loud>
              {busy === 'launch' ? 'Launching…' : 'Launch'}
            </Act>
          )}
          {c.status === 'active' ? (
            <Act
              onPress={() => void run('status', () => update(http!, c.id, { ...plan(c), status: 'paused' }))}
              disabled={busy !== ''}
            >
              {busy === 'status' ? 'Pausing…' : 'Pause'}
            </Act>
          ) : c.status === 'paused' ? (
            <Act
              onPress={() => void run('status', () => update(http!, c.id, { ...plan(c), status: 'active' }))}
              disabled={busy !== ''}
            >
              {busy === 'status' ? 'Resuming…' : 'Resume'}
            </Act>
          ) : null}
          <Act
            onPress={() => (sure ? void run('drop', () => drop(http!, c.id)) : setSure(true))}
            disabled={busy !== ''}
          >
            {busy === 'drop' ? 'Deleting…' : sure ? 'Delete for good' : 'Delete'}
          </Act>
        </XStack>
        {failed ? (
          <Text fontSize="$2" color="$ink" fontWeight="500">
            {failed}
          </Text>
        ) : null}
        <Text fontSize="$1" color="$quiet">
          Launching creates the campaign on {named(c.platform)}. It does not deliver or spend until it has ad sets
          there. Pause, resume and delete change this page only; a launched campaign keeps its state on the network.
        </Text>
      </YStack>

      <YStack gap="$3" pt="$4" borderTopWidth={1} borderColor="$borderColor">
        <Text fontSize="$3" fontWeight="500" color="$ink">
          Details
        </Text>
        <Fields value={draft} onChange={setDraft} />
        <XStack>
          <Act
            onPress={() => void run('save', () => update(http!, c.id, toPlan(draft)))}
            disabled={busy !== '' || !draft.name.trim()}
          >
            {busy === 'save' ? 'Saving…' : 'Save changes'}
          </Act>
        </XStack>
      </YStack>
    </>
  )
}

export function Campaign() {
  const http = useHttp()
  const { id = '' } = useParams()
  const campaign = useCampaign(http, id)
  const c = campaign.it

  return (
    <Ready>
      <Screen
        title={c?.name ?? 'Campaign'}
        says={c ? `${named(c.platform)} campaign, started ${since(c.createdAt)}.` : 'One campaign.'}
        beside={c ? <Mark of={tone(c.status)} says={c.status} /> : null}
      >
        {campaign.failed ? <Failed what="this campaign" why={campaign.failed} /> : null}
        {c ? <Body c={c} again={campaign.again} /> : null}
      </Screen>
    </Ready>
  )
}
