import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

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

const publicPathPrefixes = [
  "/api/auth",
  "/api/z1/ping",
  "/_next",
  "/favicon.ico",
  "/assets",
];

function isPublicPath(pathname: string): boolean {
  if (publicPaths.includes(pathname)) {
    return true;
  }

  for (const prefix of publicPathPrefixes) {
    if (pathname.startsWith(prefix)) {
      return true;
    }
  }

  return false;
}

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  if (isPublicPath(pathname)) {
    return NextResponse.next();
  }

  const sessionToken = request.cookies.get("next-auth.session-token");
  const altSessionToken = request.cookies.get("__Secure-next-auth.session-token");

  if (!sessionToken && !altSessionToken) {
    const loginUrl = new URL("/login", request.url);
    loginUrl.searchParams.set("redirect", pathname);
    return NextResponse.redirect(loginUrl);
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    "/((?!api/zort/v1/register|api/zort/v1/authenticate|api/zort/v1/heartbeat|_vercel|.*\\..*).*)",
  ],
};
