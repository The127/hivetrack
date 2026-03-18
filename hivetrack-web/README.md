# hivetrack-web

Public website for Hivetrack — static Astro site deployed to Vercel.

## Development

```bash
npm install
npm run dev      # dev server at http://localhost:4321
npm run build    # static output → dist/
npm run preview  # preview built output
```

## Deployment (Vercel)

The site deploys automatically via GitHub Actions (`.github/workflows/website.yml`):

- **Pull requests** touching `hivetrack-web/**` → Vercel preview deployment, URL posted as a PR comment
- **Merge to `main`** with changes in `hivetrack-web/**` → Vercel production deployment

### Required GitHub secrets

| Secret | How to get it |
|---|---|
| `VERCEL_TOKEN` | Vercel dashboard → Settings → Tokens |
| `VERCEL_ORG_ID` | `vercel link` → `.vercel/project.json` → `orgId` |
| `VERCEL_PROJECT_ID` | `vercel link` → `.vercel/project.json` → `projectId` |

### One-time Vercel setup

1. Install the Vercel CLI: `npm i -g vercel`
2. Run `vercel link` inside `hivetrack-web/` and follow the prompts to create or link the project
3. Copy `orgId` and `projectId` from `.vercel/project.json` into GitHub repository secrets
4. Add your Vercel API token as the `VERCEL_TOKEN` secret
