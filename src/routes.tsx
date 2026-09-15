import { createBrowserRouter } from 'react-router'

import { Accounts } from '~/Accounts'
import { Callback } from '~/callback'
import { Campaign } from '~/Campaign'
import { Campaigns } from '~/Campaigns'
import { Frame } from '~/frame'
import { New } from '~/New'
import { Root } from '~/root'
import { Spend } from '~/Spend'

/** The rail's screens. The router reads the same table. */
export const SCREENS = [
  { path: '/', label: 'Campaigns', element: <Campaigns /> },
  { path: '/new', label: 'New campaign', element: <New /> },
  { path: '/accounts', label: 'Ad accounts', element: <Accounts /> },
  { path: '/spend', label: 'Spend', element: <Spend /> },
] as const

export const router = createBrowserRouter([
  {
    element: <Root />,
    children: [
      // Where hanzo.id returns a browser; outside the frame.
      { path: '/auth/callback', element: <Callback /> },
      {
        element: <Frame />,
        children: [
          ...SCREENS.map(({ path, element }) => ({ path, element })),
          { path: '/campaigns/:id', element: <Campaign /> },
          { path: '*', element: <Campaigns /> },
        ],
      },
    ],
  },
])
