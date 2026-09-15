import { useIam } from '@hanzo/iam/react'
import { Box, Text, YStack } from '@hanzo/ui'
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router'

/** The return from hanzo.id. The SDK exchanges the code; this only lands the reader. */
export function Callback() {
  const go = useNavigate()
  const { handleCallback } = useIam()
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    handleCallback()
      .then(() => !cancelled && go('/', { replace: true }))
      .catch((err: unknown) => !cancelled && setError(err instanceof Error ? err.message : 'Sign-in failed'))
    return () => {
      cancelled = true
    }
  }, [handleCallback, go])

  return (
    <Box flex={1} items="center" justify="center" bg="$background">
      <YStack gap="$2" items="center">
        <Text fontSize="$5" color="$ink">
          {error ?? 'Completing sign-in…'}
        </Text>
        <Text fontSize="$3" color="$soft">
          {error ? 'Try signing in again.' : 'One moment.'}
        </Text>
      </YStack>
    </Box>
  )
}
