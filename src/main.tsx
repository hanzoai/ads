import { Hanzo } from '@hanzo/ui'
import '@hanzo/font/css'
import '@hanzo/ui/theme.css'
import '@hanzo/ui/styles/motion.css'
import '~/ads.css'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'

import { App } from '~/app'

// The shared components' telemetry client has no key here; keep it off.
;(globalThis as { __HANZO_TELEMETRY__?: { enabled?: boolean } }).__HANZO_TELEMETRY__ = { enabled: false }

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <Hanzo theme="dark">
      <App />
    </Hanzo>
  </StrictMode>,
)
