// A state on a palette with no hues. Luminance carries degree: running is
// brightest, paused is the middle, draft and completed are dim. The word is
// always beside the dot.

import { Text, XStack, YStack } from '@hanzo/ui'

export type Tone = 'up' | 'moving' | 'quiet'

const step = {
  up: { ink: '$ink', dot: '$ink', ring: '$ink' },
  moving: { ink: '$soft', dot: '$soft', ring: '$soft' },
  quiet: { ink: '$quiet', dot: 'transparent', ring: '$quiet' },
} as const

/** A campaign status as a step. */
export function tone(status: string): Tone {
  if (status === 'active') return 'up'
  if (status === 'paused') return 'moving'
  return 'quiet'
}

export function Mark({ of, says }: { of: Tone; says: string }) {
  const it = step[of]
  return (
    <XStack items="center" gap="$2">
      <YStack width={6} height={6} rounded={999} bg={it.dot} borderWidth={1} borderColor={it.ring} />
      <Text fontSize="$1" color={it.ink}>
        {says}
      </Text>
    </XStack>
  )
}
