/**
 * Authentication Middleware
 *
 * This middleware protects routes by checking for valid authentication sessions.
 * It redirects unauthenticated users to the login page while allowing access
 * to public routes.
 *
 * How it works:
 *
 * 1. Check if the requested path is a public route
 *    - Public routes are defined in publicPaths and publicPathPrefixes
 *    - If public, allow access without authentication
 *
 * 2. Check for valid session
 *    - Look for NextAuth session cookies
 *    - Supports both "next-auth.session-token" and "__Secure-next-auth.session-token"
 *
 * 3. Redirect to login if no valid session
 *    - Include the original URL as a redirect parameter
 *    - This allows redirecting back after successful login
 *
 * Security Considerations:
 *
 * - Session cookies are HTTP-only and cannot be accessed via JavaScript
 * - The matcher excludes VORT agent API routes (they use JWT, not sessions)
 * - Static assets and Next.js internals are always accessible
 *
 * Public Routes:
 *
 * The following paths are accessible without authentication:
 * - /, /login, /auth/login - Login pages
 * - /api/auth/* - NextAuth API routes
 * - /api/zitadel/session - Zitadel session proxy
 * - /health, /healthz, /ping - Health checks
 * - /_next/*, /favicon.ico, /assets/* - Static assets
 *
 * VORT Agent Routes:
 *
 * Agent API routes (/api/zort/*) are excluded from this middleware because:
 * - They use JWT authentication, not session cookies
 * - They have their own authentication logic in the backend
 *
 * Environment:
 *
 * This middleware runs on the Edge, so it must be lightweight and
 * cannot use Node.js-specific APIs.
 */

import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

/**
 * Exact paths that do not require authentication.
 * These are typically pages that must be publicly accessible.
 */
const publicPaths = [
  "/",
  "/login",
  "/auth/login",
  "/api/auth",
  "/api/zitadel/session",
  "/health",
  "/healthz",
  "/ping",
  "/api/v1/ping",
];

/**
 * Path prefixes that do not require authentication.
 * All paths starting with these prefixes are publicly accessible.
 */
const publicPathPrefixes = [
  "/api/auth",        // NextAuth API endpoints
  "/api/z1/ping",    // Internal ping endpoint
  "/_next",          // Next.js static assets
  "/favicon.ico",   // Favicon
  "/assets",         // Static asset folders
];

/**
 * Determines if a given pathname is a public route that doesn't require authentication.
 *
 * Checks both exact matches (publicPaths) and prefix matches (publicPathPrefixes).
 * This allows fine-grained control over which routes are publicly accessible.
 *
 * @param pathname - The request path to check
 * @returns true if the path is public (no auth required), false otherwise
 *
 * @example
 * isPublicPath("/login")     // true - exact match
 * isPublicPath("/api/auth") // true - prefix match
 * isPublicPath("/admin")    // false - requires auth
 */
function isPublicPath(pathname: string): boolean {
  // Check exact path matches first (O(1) lookup)
  if (publicPaths.includes(pathname)) {
    return true;
  }

  // Check prefix matches (O(n) where n = number of prefixes)
  for (const prefix of publicPathPrefixes) {
    if (pathname.startsWith(prefix)) {
      return true;
    }
  }

  // Path requires authentication
  return false;
}

/**
 * Next.js middleware function that runs before each request.
 *
 * This middleware:
 * 1. Skips authentication for public paths
 * 2. Checks for valid NextAuth session cookies
 * 3. Redirects unauthenticated users to login
 *
 * @param request - The incoming Next.js request
 * @returns NextResponse to either allow or redirect the request
 */
export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Step 1: Allow public paths without authentication
  // This includes health checks, login pages, and static assets
  if (isPublicPath(pathname)) {
    return NextResponse.next();
  }

  // Step 2: Check for valid session cookie
  // NextAuth stores the session in an HTTP-only cookie
  // We check both cookie names for compatibility
  const sessionToken = request.cookies.get("next-auth.session-token");
  const altSessionToken = request.cookies.get("__Secure-next-auth.session-token");

  // Step 3: Redirect to login if no valid session
  if (!sessionToken && !altSessionToken) {
    const loginUrl = new URL("/login", request.url);
    // Save the original URL to redirect back after login
    loginUrl.searchParams.set("redirect", pathname);
    return NextResponse.redirect(loginUrl);
  }

  // Session is valid - allow the request
  return NextResponse.next();
}

/**
 * Middleware configuration for Next.js.
 *
 * The matcher defines which routes the middleware applies to.
 * It uses negative lookahead to exclude certain patterns:
 *
 * - api/zort/v1/register - Agent registration (public)
 * - api/zort/v1/authenticate - Agent auth (public)
 * - api/zort/v1/heartbeat - Agent heartbeat (has its own auth)
 * - _vercel - Vercel internals
 * - .* - Files with extensions (static assets)
 */
export const config = {
  matcher: [
    // Match all paths except:
    // - VORT agent API routes (they have their own JWT auth)
    // - Vercel internal paths
    // - Files with extensions (static files)
    "/((?!api/zort/v1/register|api/zort/v1/authenticate|api/zort/v1/heartbeat|_vercel|.*\\..*).*)",
  ],
};
