import { hanzo } from '@hanzo/vite'
import react from '@vitejs/plugin-react'

// api.hanzo.ai refuses a credentialed request from a localhost port, so dev and
// preview serve /v1 from their own origin and forward it.
const proxy = { '/v1': { target: 'https://api.hanzo.ai', changeOrigin: true } }

// Pinned: the port is part of the redirect URI registered with hanzo.id.
const PORT = 3310

// Every unknown address must reach the SPA; static hosts serve 404.html for it.
const spaFallback = {
  name: 'spa-fallback',
  async writeBundle(options: { dir?: string }) {
    const { copyFile } = await import('node:fs/promises')
    const { join } = await import('node:path')
    const dir = options.dir ?? 'dist'
    await copyFile(join(dir, 'index.html'), join(dir, '404.html'))
  },
}

const config = hanzo(
  {
    plugins: [react(), spaFallback],
    server: { port: PORT, allowedHosts: true, proxy },
    preview: { port: PORT, allowedHosts: true, proxy },
  },
  {
    root: import.meta.dirname,
    // One copy of the theme context, or components throw "Missing theme."
    dedupe: ['@hanzogui/core', '@hanzogui/web', '@hanzogui/portal', '@hanzogui/toast'],
  },
)

// The dependency optimizer resolves separately; give it the same web-first extensions.
export default {
  ...config,
  optimizeDeps: {
    rollupOptions: {
      resolve: { extensions: (config.resolve as { extensions: string[] }).extensions },
    },
  },
}
