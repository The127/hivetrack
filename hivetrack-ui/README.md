# Hivetrack UI

Vue 3 frontend for Hivetrack. See the [root README](../README.md) for project overview.

## Requirements

- Node.js 20+
- Backend running on `:8080` (proxied automatically in dev)

## Running locally

```bash
# From repo root
just ui-dev

# Or directly
npm install
npm run dev
```

Dev server runs on `http://localhost:5173`. API calls are proxied to `http://localhost:8080`.

## Commands

```bash
npm run dev          # dev server with HMR
npm run build        # production build → dist/
npm run lint         # lint and fix
npm run lint:check   # lint without fixing (for CI)
```

## Stack

- **Vue 3** with `<script setup>` Composition API
- **Vite** for building
- **TanStack Query** for server state
- **Vue Router 4** for routing
- **Tailwind CSS v4** for styling
- **oidc-client-ts** for OIDC auth
- **Native fetch** for HTTP (no Axios)
- **Lucide Vue Next** for icons

See [`CLAUDE.md`](./CLAUDE.md) for architecture guide and recipes.
