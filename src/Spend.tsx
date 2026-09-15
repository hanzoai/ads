// Budget against spend, overall and per network. Only the two amounts each
// campaign carries; the platform keeps no delivery numbers.

import { Link } from 'react-router'
import { Text, XStack, YStack } from '@hanzo/ui'

import { Failed, Nothing, Ready, Screen } from '~/page'
import { named, useCampaigns, useSummary, type Campaign } from '~/ads'
import { money } from '~/units'
import { useHttp } from '~/client'

/** Spend as a share of budget, drawn in the ramp. */
function Share({ spend, budget }: { spend: number; budget: number }) {
  const part = budget > 0 ? Math.min(100, Math.round((spend / budget) * 100)) : 0
  return (
    <XStack items="center" gap="$2" flex={1} minW={160}>
      <YStack flex={1} height={4} rounded={999} bg="$panel" overflow="hidden">
        <YStack width={`${part}%`} height={4} bg="$ink" />
      </YStack>
      <Text fontSize="$1" color="$quiet" minW={40} text="right">
        {budget > 0 ? `${part}%` : '—'}
      </Text>
    </XStack>
  )
}

interface Tally {
  platform: string
  count: number
  budget: number
  spend: number
}

function byNetwork(rows: Campaign[]): Tally[] {
  const by = new Map<string, Tally>()
  for (const c of rows) {
    const t = by.get(c.platform) ?? { platform: c.platform, count: 0, budget: 0, spend: 0 }
    t.count += 1
    t.budget += c.budget
    t.spend += c.spend
    by.set(c.platform, t)
  }
  return [...by.values()].sort((a, b) => b.spend - a.spend)
}

function Line({ label, sub, spend, budget, to }: { label: string; sub: string; spend: number; budget: number; to?: string }) {
  const row = (
    <XStack items="center" gap="$3" py="$3" px="$2" borderBottomWidth={1} borderColor="$borderColor" flexWrap="wrap" hoverStyle={to ? { bg: '$hover' } : {}}>
      <YStack minW={200} flex={1} gap="$1">
        <Text fontSize="$2" color="$ink" numberOfLines={1}>
          {label}
        </Text>
        <Text fontSize="$1" color="$quiet">
          {sub}
        </Text>
      </YStack>
      <Text fontSize="$1" color="$soft" minW={170} text="right">
        {money(spend)} of {money(budget)}
      </Text>
      <Share spend={spend} budget={budget} />
    </XStack>
  )
  return to ? (
    <Link to={to} style={{ textDecoration: 'none' }}>
      {row}
    </Link>
  ) : (
    row
  )
}

export function Spend() {
  const http = useHttp()
  const summary = useSummary(http)
  const campaigns = useCampaigns(http)
  const rows = campaigns.it ?? []
  const s = summary.it
  const ranked = [...rows].sort((a, b) => b.spend - a.spend)

  return (
    <Ready>
      <Screen title="Spend" says="What has been spent against each budget.">
        {summary.failed ? (
          <Failed what="the totals" why={summary.failed} />
        ) : (
          <YStack gap="$2" p="$4" rounded="$3" borderWidth={1} borderColor="$borderColor">
            <XStack items="baseline" gap="$2" flexWrap="wrap">
              <Text fontSize="$7" fontWeight="600" color="$ink">
                {s ? money(s.spend) : '—'}
              </Text>
              <Text fontSize="$2" color="$soft">
                spent of {s ? money(s.budget) : '—'} budgeted, across {s?.campaigns ?? 0} campaigns,{' '}
                {s?.active ?? 0} active
              </Text>
            </XStack>
            <Share spend={s?.spend ?? 0} budget={s?.budget ?? 0} />
          </YStack>
        )}

        {campaigns.failed ? <Failed what="the campaigns" why={campaigns.failed} /> : null}
        {campaigns.it && rows.length === 0 ? (
          <Nothing says="No campaigns yet, so nothing has been spent." />
        ) : (
          <>
            <YStack>
              <Text fontSize="$3" fontWeight="500" color="$ink" pb="$1">
                By network
              </Text>
              {byNetwork(rows).map((t) => (
                <Line
                  key={t.platform}
                  label={named(t.platform)}
                  sub={`${t.count} campaign${t.count === 1 ? '' : 's'}`}
                  spend={t.spend}
                  budget={t.budget}
                />
              ))}
            </YStack>
            <YStack>
              <Text fontSize="$3" fontWeight="500" color="$ink" pb="$1">
                By campaign
              </Text>
              {ranked.map((c) => (
                <Line
                  key={c.id}
                  label={c.name}
                  sub={`${named(c.platform)} · ${c.status}`}
                  spend={c.spend}
                  budget={c.budget}
                  to={`/campaigns/${c.id}`}
                />
              ))}
            </YStack>
          </>
        )}

        <Text fontSize="$1" color="$quiet">
          Spend is the amount recorded on each campaign. It is not read back from the ad network, and impressions,
          clicks and conversions are not kept here.
        </Text>
      </Screen>
    </Ready>
  )
}
