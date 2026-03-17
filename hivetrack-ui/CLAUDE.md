# Hivetrack UI — Claude Instructions

Vue 3 frontend. Read the root `CLAUDE.md` and `docs/` before making changes here.

**Run:** `just ui-dev` from repo root, or `npm run dev` from this directory.
**Build:** `just ui-build` — output goes to `dist/`, embedded in the Go binary.

---

## Core rules

- **Composition API with `<script setup>` only.** No Options API, no `defineComponent`.
- **TanStack Query for all server state.** No manual `ref` + `fetch` loops.
- **No Pinia/Vuex.** Local state in `ref`/`reactive`. Auth state in `useAuth` composable.
- **Native `fetch` API — no Axios.** Use the `apiFetch` wrapper from `useApi.js` which adds auth headers and throws on non-2xx.
- **Tailwind v4 for styling.** No custom CSS except in `style.css` for global resets.
- **`@` alias resolves to `src/`.** Use it everywhere: `import Foo from '@/components/Foo.vue'`

---

## Adding a new page

### 1. Create the view
`src/views/ProjectBoardView.vue`:
```vue
<script setup>
import { useRoute } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'
import { fetchBoard } from '@/api/issues'

const route = useRoute()
const slug = computed(() => route.params.slug)

const { data: board, isLoading } = useQuery({
  queryKey: ['board', slug],
  queryFn: () => fetchBoard(slug.value),
})
</script>

<template>
  <MainLayout>
    <!-- content -->
  </MainLayout>
</template>
```

### 2. Register the route
`src/router/index.js`:
```js
{
  path: '/projects/:slug/board',
  component: () => import('@/views/ProjectBoardView.vue'),
  meta: {
    requiresAuth: true,
    layout: 'main',
  },
}
```

### 3. Add the API function
`src/api/issues.js`:
```js
import { apiFetch } from '@/composables/useApi'

export const fetchBoard = (slug) =>
  apiFetch(`/api/v1/projects/${slug}/board`)
```

---

## TanStack Query patterns

```js
// Read (query)
const { data, isLoading, error } = useQuery({
  queryKey: ['issues', projectSlug, filters],  // include all deps in key
  queryFn: () => fetchIssues(projectSlug, filters),
})

// Write (mutation) with optimistic update
const queryClient = useQueryClient()
const { mutate: updateStatus } = useMutation({
  mutationFn: ({ issueId, status }) => patchIssueStatus(issueId, status),
  onMutate: async ({ issueId, status }) => {
    // Cancel in-flight queries for this key
    await queryClient.cancelQueries({ queryKey: ['issues', projectSlug] })
    // Snapshot for rollback
    const previous = queryClient.getQueryData(['issues', projectSlug])
    // Optimistically update
    queryClient.setQueryData(['issues', projectSlug], old => /* update */)
    return { previous }
  },
  onError: (_err, _vars, context) => {
    queryClient.setQueryData(['issues', projectSlug], context.previous)
  },
  onSettled: () => {
    queryClient.invalidateQueries({ queryKey: ['issues', projectSlug] })
  },
})
```

Always use optimistic updates for actions the user takes on the board (drag, status change, assign). Never show a spinner for these — they should feel instant.

---

## Auth

```js
// In any component or composable
import { useAuth } from '@/composables/useAuth'

const { user, isAuthenticated, signIn, signOut } = useAuth()
```

The router guard in `src/router/index.js` handles redirecting unauthenticated users. Components do not need to check auth themselves — they only render if the guard passes.

OIDC config is fetched from the backend at startup (`GET /api/v1/auth/oidc-config`). Do not hardcode the authority URL anywhere in the frontend.

## API fetch wrapper

All API calls go through `apiFetch` from `@/composables/useApi.js`. It:
- Attaches the OIDC access token as `Authorization: Bearer <token>`
- Throws a typed error on non-2xx responses (so TanStack Query's `error` state works correctly)
- Returns parsed JSON

```js
import { apiFetch } from '@/composables/useApi'

// GET
const data = await apiFetch('/api/v1/projects')

// POST
const issue = await apiFetch('/api/v1/projects/ht/issues', {
  method: 'POST',
  body: JSON.stringify({ title: 'Fix login bug', type: 'task' }),
})

// DELETE
await apiFetch(`/api/v1/projects/ht/issues/42`, { method: 'DELETE' })
```

---

## File structure

```
src/
├── api/              # One file per resource (issues.js, projects.js, sprints.js)
│                     # Functions return promises. No state here.
├── composables/      # Reusable reactive logic
│   ├── useAuth.js    # OIDC user manager, current user
│   └── useApi.js     # Axios instance with auth header
├── views/            # Page components (one per route)
│   # Named: [Resource][Action]View.vue e.g. ProjectBoardView, IssueDetailView
├── components/
│   ├── ui/           # Base components: Button, Badge, Input, Modal, etc.
│   └── [feature]/    # Feature components: IssueCard, SprintHeader, etc.
├── layouts/
│   ├── MainLayout.vue    # Full app chrome (nav, sidebar) — requires auth
│   ├── MinimalLayout.vue # Centered, minimal — settings/onboarding
│   └── PublicLayout.vue  # No auth — customer portal, support tracking
├── router/index.js   # All routes with meta
├── App.vue           # Root — just <RouterView />
├── main.js           # App bootstrap
└── style.css         # @import "tailwindcss" only
```

---

## Naming conventions

| Type | Convention | Example |
|---|---|---|
| View components | `[Resource][Action]View.vue` | `ProjectBoardView.vue` |
| Feature components | `[Resource][Description].vue` | `IssueCard.vue`, `SprintHeader.vue` |
| Base UI components | `[Name].vue` in `components/ui/` | `Button.vue`, `Badge.vue` |
| Composables | `use[Name].js` | `useAuth.js`, `useIssueFilters.js` |
| API functions | `fetch[Resource]`, `create[Resource]`, `update[Resource]`, `delete[Resource]` | `fetchIssues`, `createIssue` |
| Query keys | `[resource, ...ids, ...filters]` | `['issues', slug, { status: 'open' }]` |
