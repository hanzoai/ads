# Hanzo Ads

Plan a paid campaign, launch it on Meta, and keep its budget and spend beside
every other campaign your organization runs.

This repository holds two things:

- the web app (`src/`, a Vite SPA on `@hanzo/ui`), which runs campaigns through
  `api.hanzo.ai/v1/ad` and connects ad accounts through `/v1/integration`;
- the Go module `github.com/hanzoai/ads`, with network connectors and `cmd/ads`.

## Run the app

```sh
pnpm install
pnpm dev        # http://localhost:3310, /v1 forwarded to api.hanzo.ai
```

Sign-in is Hanzo IAM (`hanzo.id`, app `hanzo-ads`).

## Build

```sh
pnpm build      # typecheck, then dist/
```

The image is built by CI from `Dockerfile` and `hanzo.yml` and served by
`ghcr.io/hanzoai/spa`.

## Screens

- **Campaigns** — every campaign, filtered by status, with totals.
- **New campaign** — store a campaign; it reaches its network only on launch.
- **Campaign** — launch, pause, resume, edit, delete.
- **Ad accounts** — connect Meta, Google or TikTok ad accounts.
- **Spend** — spend against budget, per network and per campaign.

## License

Apache-2.0 OR MIT.
