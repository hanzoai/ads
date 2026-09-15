// Starting a campaign. It is stored here; launching it is a separate step.

import { useState } from 'react'
import { useNavigate } from 'react-router'
import { Text, XStack } from '@hanzo/ui'

import { Act, Ready, Screen } from '~/page'
import { Fields, blank, toPlan } from '~/form'
import { create, why } from '~/ads'
import { useHttp } from '~/client'

export function New() {
  const http = useHttp()
  const go = useNavigate()
  const [draft, setDraft] = useState(blank)
  const [busy, setBusy] = useState(false)
  const [failed, setFailed] = useState<string | null>(null)

  const save = async () => {
    if (!http || busy || !draft.name.trim()) return
    setBusy(true)
    setFailed(null)
    try {
      const saved = await create(http, toPlan(draft))
      go(`/campaigns/${saved.id}`)
    } catch (e) {
      setFailed(why(e))
      setBusy(false)
    }
  }

  return (
    <Ready>
      <Screen title="New campaign" says="Save it first. It reaches its network only when you launch it.">
        <Fields value={draft} onChange={setDraft} />
        <XStack gap="$3" items="center" flexWrap="wrap">
          <Act onPress={() => void save()} disabled={busy || !draft.name.trim()} loud>
            {busy ? 'Saving…' : 'Save campaign'}
          </Act>
          {failed ? (
            <Text fontSize="$1" color="$ink" fontWeight="500">
              {failed}
            </Text>
          ) : null}
        </XStack>
      </Screen>
    </Ready>
  )
}
