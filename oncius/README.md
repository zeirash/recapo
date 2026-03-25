# Oncius

Next.js 14 frontend for Recapo. Runs on port 3000.

## Tech Stack

- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript
- **Styling**: MUI (Material UI) + Emotion + SASS
- **Data fetching**: React Query v3 (`react-query`)
- **i18n**: next-intl (EN / ID)
- **Icons**: lucide-react

## Development

```bash
cp .env.example .env.local  # fill in values
npm install
npm run dev
```

App runs on http://localhost:3000. Requires arion running on port 4000.

## Scripts

```bash
npm run dev          # development server
npm run build        # production build
npm run lint         # ESLint
npm run type-check   # TypeScript check without emit
```

## Project Structure

```
src/
├── app/             # Next.js App Router pages
├── components/      # Shared UI components
├── hooks/           # useAuth, useLocale
├── providers/       # React context providers
├── types/           # TypeScript type definitions
└── utils/api.ts     # Centralized API client
```

## Environment Variables

Copy `.env.example` to `.env.local` and fill in:

| Variable | Description |
|---|---|
| `NEXT_PUBLIC_API_BASE_URL` | Arion backend URL (default `http://localhost:4000`) |
| `NEXT_PUBLIC_SENTRY_DSN` | Sentry error tracking (optional) |

## Docker

```bash
docker build -t oncius .
docker run -p 3000:3000 oncius
```
