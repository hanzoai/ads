import { YStack } from '@hanzo/ui'
import { ThemeProvider } from 'next-themes'
import { RouterProvider } from 'react-router/dom'

import { router } from '~/routes'

/** The app fills the window and clips; each pane scrolls inside it. */
export const App = () => (
  <ThemeProvider attribute="class" defaultTheme="dark" enableSystem>
    <YStack flex={1} minW={0} bg="$background" style={{ height: '100dvh', overflow: 'hidden' }}>
      <RouterProvider router={router} />
    </YStack>
  </ThemeProvider>
)
