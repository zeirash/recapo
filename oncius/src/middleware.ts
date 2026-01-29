import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

// Bypass next-intl middleware - using client-side locale from localStorage instead.
// Restore createMiddleware from 'next-intl/middleware' if you add [locale] routing.
export function middleware(request: NextRequest) {
  return NextResponse.next();
}

export const config = {
  matcher: ['/((?!api|_next|.*\\..*).*)']
};
