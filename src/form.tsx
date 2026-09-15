// The fields of a campaign, shared by the new-campaign and campaign screens.

import type { ReactNode } from 'react'
import { Text, XStack, YStack } from '@hanzo/ui'

import { PLATFORMS, STATUSES, named, type Plan } from '~/ads'
import { cents, money } from '~/units'

/** A plan as the inputs hold it: amounts stay typed text until saved. */
export interface Draft {
  name: string
  platform: string
  account: string
  status: string
  objective: string
  budget: string
  spend: string
}

export const blank: Draft = {
  name: '',
  platform: 'meta',
  account: '',
  status: 'draft',
  objective: '',
  budget: '',
  spend: '',
}

export const toDraft = (p: Plan): Draft => ({
  ...p,
  budget: p.budget ? money(p.budget) : '',
  spend: p.spend ? money(p.spend) : '',
})

export const toPlan = (d: Draft): Plan => ({
  ...d,
  name: d.name.trim(),
  account: d.account.trim(),
  objective: d.objective.trim(),
  budget: cents(d.budget),
  spend: cents(d.spend),
})

function Field({ label, hint, children }: { label: string; hint?: string; children: ReactNode }) {
  return (
    <YStack gap="$1" flex={1} minW={220}>
      <Text fontSize="$1" color="$soft">
        {label}
      </Text>
      {children}
      {hint ? (
        <Text fontSize="$1" color="$quiet">
          {hint}
        </Text>
      ) : null}
    </YStack>
  )
}

export function Fields({ value, onChange }: { value: Draft; onChange: (d: Draft) => void }) {
  const set = (key: keyof Draft) => (e: { target: { value: string } }) =>
    onChange({ ...value, [key]: e.target.value })

  return (
    <YStack gap="$3">
      <XStack gap="$3" flexWrap="wrap">
        <Field label="Name">
          <input className="field" value={value.name} onChange={set('name')} placeholder="Spring launch" />
        </Field>
        <Field label="Network">
          <select className="field" value={value.platform} onChange={set('platform')}>
            {PLATFORMS.map((p) => (
              <option key={p} value={p}>
                {named(p)}
              </option>
            ))}
          </select>
        </Field>
      </XStack>
      <XStack gap="$3" flexWrap="wrap">
        <Field label="Ad account" hint="The account on the network, e.g. act_1234 on Meta.">
          <input className="field" value={value.account} onChange={set('account')} />
        </Field>
        <Field label="Objective" hint="As the network names it. Meta uses OUTCOME_TRAFFIC when empty.">
          <input className="field" value={value.objective} onChange={set('objective')} placeholder="OUTCOME_TRAFFIC" />
        </Field>
      </XStack>
      <XStack gap="$3" flexWrap="wrap">
        <Field label="Budget">
          <input className="field" inputMode="decimal" value={value.budget} onChange={set('budget')} placeholder="0.00" />
        </Field>
        <Field label="Spent so far">
          <input className="field" inputMode="decimal" value={value.spend} onChange={set('spend')} placeholder="0.00" />
        </Field>
        <Field label="Status">
          <select className="field" value={value.status} onChange={set('status')}>
            {STATUSES.map((s) => (
              <option key={s} value={s}>
                {s}
              </option>
            ))}
          </select>
        </Field>
      </XStack>
    </YStack>
  )
}
