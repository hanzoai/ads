// The rail of screens on the left, the open screen beside it.

import { useState } from 'react'
import { useIam } from '@hanzo/iam/react'
import { Check, ChevronDown } from 'lucide-react'
import { Box, Text, XStack, YStack } from '@hanzo/ui'
import { HanzoMark } from '@hanzo/ui/product'
import { NavLink, Outlet } from 'react-router'

import { enter } from '~/enter'
import { SCREENS } from '~/routes'
import { useTenant } from '~/client'

/** The organization every campaign on screen belongs to. One is a name, not a choice. */
function Tenant() {
  const { org, orgs, choose } = useTenant()
  const [open, setOpen] = useState(false)

  if (orgs.length === 0) return null
  if (orgs.length === 1) {
    return (
      <Text fontSize="$1" color="$quiet" px="$3" numberOfLines={1}>
        {orgs[0]}
      </Text>
    )
  }

  return (
    <YStack gap="$1">
      <Box render="button" onClick={() => setOpen(!open)} aria-expanded={open} width="100%" px="$3" py="$2" rounded="$2" hoverStyle={{ bg: '$hover' }}>
        <XStack items="center" gap="$2">
          <Text fontSize="$2" color="$ink" flex={1} numberOfLines={1}>
            {org ?? 'Pick an organization'}
          </Text>
          <ChevronDown size={13} aria-hidden />
        </XStack>
      </Box>
      {open
        ? orgs.map((one) => (
            <Box
              key={one}
              render="button"
              onClick={() => {
                choose(one)
                setOpen(false)
              }}
              width="100%"
              px="$3"
              py="$2"
              rounded="$2"
              hoverStyle={{ bg: '$hover' }}
            >
              <XStack items="center" gap="$2">
                <Text fontSize="$2" color={one === org ? '$ink' : '$soft'} flex={1} numberOfLines={1}>
                  {one}
                </Text>
                {one === org ? <Check size={12} aria-hidden /> : null}
              </XStack>
            </Box>
          ))
        : null}
    </YStack>
  )
}

function Account() {
  const door = useIam()
  const { user, isAuthenticated, logout } = door
  const who = user as { displayName?: string; name?: string; email?: string } | null
  const name = who?.displayName || who?.name || who?.email || ''

  if (!isAuthenticated) {
    return (
      <Box render="button" onClick={() => void enter(door)} px="$3" py="$2" rounded="$3" borderWidth={1} borderColor="$borderColor" hoverStyle={{ bg: '$hover' }}>
        <Text fontSize="$2" color="$ink">
          Sign in
        </Text>
      </Box>
    )
  }

  return (
    <YStack gap="$1">
      <Text fontSize="$1" color="$quiet" px="$3" numberOfLines={1}>
        {name}
      </Text>
      <Box render="button" onClick={() => void logout()} px="$3" py="$2" rounded="$2" hoverStyle={{ bg: '$hover' }}>
        <Text fontSize="$2" color="$soft">
          Sign out
        </Text>
      </Box>
    </YStack>
  )
}

export function Frame() {
  return (
    <XStack className="shell" flex={1} minW={0} minH={0}>
      <YStack className="rail" width={232} shrink={0} minH={0} borderRightWidth={1} borderColor="$borderColor" py="$3" gap="$3">
        <XStack items="center" gap="$2" px="$3">
          <HanzoMark size={18} />
          <Text fontSize="$3" fontWeight="500" color="$ink">
            Ads
          </Text>
        </XStack>

        <YStack className="ways" gap="$1" px="$2" flex={1} minH={0} overflow="scroll">
          {SCREENS.map(({ path, label }) => (
            <NavLink key={path} to={path} end style={{ textDecoration: 'none' }}>
              {({ isActive }) => (
                <Box px="$3" py="$2" rounded="$2" bg={isActive ? '$hover' : 'transparent'} hoverStyle={{ bg: '$hover' }}>
                  <Text fontSize="$2" color={isActive ? '$ink' : '$soft'}>
                    {label}
                  </Text>
                </Box>
              )}
            </NavLink>
          ))}
        </YStack>

        <YStack className="whose" gap="$2" px="$1" shrink={0}>
          <Tenant />
          <Account />
        </YStack>
      </YStack>

      <YStack flex={1} minW={0} minH={0} overflow="scroll">
        <Outlet />
      </YStack>
    </XStack>
  )
}
