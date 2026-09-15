// What every screen shares: the heading, a control, a refusal, an empty state,
// and the signed-out state.

import type { ReactNode } from 'react'
import { useIam } from '@hanzo/iam/react'
import { Box, Text, XStack, YStack } from '@hanzo/ui'

import { enter } from '~/enter'
import { useTenant } from '~/client'

export function Screen({
  title,
  says,
  beside,
  children,
}: {
  title: string
  says: string
  beside?: ReactNode
  children: ReactNode
}) {
  return (
    <YStack flex={1} minH={0} gap="$4" p="$5">
      <XStack items="flex-start" gap="$3" flexWrap="wrap">
        <YStack gap="$1" flex={1} minW={220}>
          <Text render="h1" fontSize="$6" fontWeight="600" color="$ink">
            {title}
          </Text>
          <Text fontSize="$2" color="$soft">
            {says}
          </Text>
        </YStack>
        {beside}
      </XStack>
      {children}
    </YStack>
  )
}

/**
 * A control. `loud` inverts it and is spent on one control per screen; `on`
 * marks the picked option in a set.
 */
export function Act({
  onPress,
  children,
  disabled,
  loud,
  on,
}: {
  onPress: () => void
  children: ReactNode
  disabled?: boolean
  loud?: boolean
  on?: boolean
}) {
  return (
    <Box
      render="button"
      onClick={disabled ? undefined : onPress}
      aria-disabled={disabled}
      aria-pressed={on}
      px="$3"
      py="$2"
      rounded="$3"
      borderWidth={1}
      borderColor={loud || on ? '$ink' : '$borderColor'}
      bg={loud ? '$ink' : on ? '$hover' : 'transparent'}
      opacity={disabled ? 0.4 : 1}
      hoverStyle={disabled ? {} : { bg: loud ? '$ink' : '$hover' }}
    >
      <Text fontSize="$2" color={loud ? '$background' : '$ink'} fontWeight={loud ? '500' : '400'}>
        {children}
      </Text>
    </Box>
  )
}

/** A refusal, in the platform's own words. */
export function Failed({ what, why }: { what: string; why: string }) {
  return (
    <YStack gap="$1" p="$3" rounded="$3" borderWidth={1} borderColor="$borderColor">
      <Text fontSize="$2" color="$ink">
        Could not load {what}.
      </Text>
      <Text fontSize="$1" color="$soft">
        {why}
      </Text>
    </YStack>
  )
}

export function Nothing({ says }: { says: string }) {
  return (
    <YStack items="center" justify="center" p="$6">
      <Text fontSize="$2" color="$quiet" text="center">
        {says}
      </Text>
    </YStack>
  )
}

/** The screen, once someone is signed in and working in one organization. */
export function Ready({ children }: { children: ReactNode }) {
  const door = useIam()
  const { org, orgs } = useTenant()
  const { isAuthenticated, isLoading } = door

  if (isLoading) return <YStack flex={1} />

  if (!isAuthenticated) {
    return (
      <YStack flex={1} items="center" justify="center" gap="$3" p="$6">
        <Text render="h1" fontSize="$6" fontWeight="600" color="$ink">
          Hanzo Ads
        </Text>
        <Text fontSize="$2" color="$soft" text="center" maxW={420}>
          Plan paid campaigns, launch them on their network, and keep budget and spend in one place.
        </Text>
        <Act onPress={() => void enter(door)} loud>
          Sign in
        </Act>
      </YStack>
    )
  }

  if (!org && orgs.length > 1) {
    return (
      <YStack flex={1} items="center" justify="center" gap="$2" p="$6">
        <Text fontSize="$3" color="$ink">
          Pick an organization
        </Text>
        <Text fontSize="$2" color="$soft" text="center" maxW={420}>
          Campaigns belong to one organization. Choose yours at the bottom of the column on the left.
        </Text>
      </YStack>
    )
  }

  return <>{children}</>
}
