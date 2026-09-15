// Every campaign the org has, with the roll-up above it.

import { useState } from 'react'
import { Link, useNavigate } from 'react-router'
import { Text, XStack, YStack } from '@hanzo/ui'

import { Act, Failed, Nothing, Ready, Screen } from '~/page'
import { Mark, tone } from '~/status'
import { STATUSES, named, useCampaigns, useSummary } from '~/ads'
import { money, since } from '~/units'
import { useHttp } from '~/client'

function Count({ of, says }: { of: string; says: string }) {
  return (
    <YStack gap="$1" p="$3" rounded="$3" borderWidth={1} borderColor="$borderColor" flex={1} minW={150}>
      <Text fontSize="$7" fontWeight="600" color="$ink">
        {of}
      </Text>
      <Text fontSize="$1" color="$soft">
        {says}
      </Text>
    </YStack>
  )
}

export function Campaigns() {
  const http = useHttp()
  const go = useNavigate()
  const [status, setStatus] = useState('')
  const summary = useSummary(http)
  const campaigns = useCampaigns(http, status || undefined)
  const rows = campaigns.it ?? []
  const s = summary.it

  return (
    <Ready>
      <Screen
        title="Campaigns"
        says="Every paid campaign in this organization."
        beside={
          <Act onPress={() => go('/new')} loud>
            New campaign
          </Act>
        }
      >
        {summary.failed ? (
          <Failed what="the totals" why={summary.failed} />
        ) : (
          <XStack gap="$3" flexWrap="wrap">
            <Count of={s ? String(s.campaigns) : '—'} says="Campaigns" />
            <Count of={s ? String(s.active) : '—'} says="Active" />
            <Count of={s ? money(s.budget) : '—'} says="Budget" />
            <Count of={s ? money(s.spend) : '—'} says="Spent" />
          </XStack>
        )}

        <XStack gap="$2" flexWrap="wrap">
          {['', ...STATUSES].map((one) => (
            <Act key={one || 'all'} onPress={() => setStatus(one)} on={status === one}>
              {one || 'everything'}
            </Act>
          ))}
        </XStack>

        {campaigns.failed ? (
          <Failed what="the campaigns" why={campaigns.failed} />
        ) : campaigns.it && rows.length === 0 ? (
          <Nothing says={status ? `No ${status} campaigns.` : 'No campaigns yet. Start one and it appears here.'} />
        ) : (
          <YStack>
            {rows.map((c) => (
              <Link key={c.id} to={`/campaigns/${c.id}`} style={{ textDecoration: 'none' }}>
                <XStack
                  items="center"
                  gap="$3"
                  py="$3"
                  px="$2"
                  borderBottomWidth={1}
                  borderColor="$borderColor"
                  hoverStyle={{ bg: '$hover' }}
                  flexWrap="wrap"
                >
                  <YStack flex={1} minW={200} gap="$1">
                    <Text fontSize="$2" color="$ink" numberOfLines={1}>
                      {c.name}
                    </Text>
                    <Text fontSize="$1" color="$quiet">
                      {named(c.platform)}
                      {c.objective ? ` · ${c.objective}` : ''} · changed {since(c.updatedAt)}
                    </Text>
                  </YStack>
                  <Text fontSize="$1" color="$soft" minW={170} text="right">
                    {money(c.spend)} of {money(c.budget)}
                  </Text>
                  <YStack minW={80}>
                    <Mark of={tone(c.status)} says={c.status} />
                  </YStack>
                </XStack>
              </Link>
            ))}
          </YStack>
        )}
      </Screen>
    </Ready>
  )
}
