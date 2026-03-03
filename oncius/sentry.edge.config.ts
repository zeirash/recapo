import * as Sentry from '@sentry/nextjs'

const dsn = process.env.NEXT_PUBLIC_SENTRY_DSN

Sentry.init({
  dsn,
  environment: process.env.NODE_ENV,
  enabled: !!dsn,
  tracesSampleRate: 0,
})
