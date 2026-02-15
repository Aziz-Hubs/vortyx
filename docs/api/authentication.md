# Authentication

Vortyx uses **OIDC (OpenID Connect)** and **Zitadel** for authentication and authorization.

## Overview

The authentication flow is based on standard OIDC/OAuth2 protocols:
1.  **Login**: User logs in via the Frontend (redirects to Zitadel).
2.  **Token**: Zitadel issues an `access_token` (JWT) and `id_token`.
3.  **Request**: Frontend sends the `access_token` in the `Authorization: Bearer <token>` header to the Backend.
4.  **Validation**: Backend validates the token signature using Zitadel's public keys (JWKS).
5.  **Context**: Validated user identity (sub, email, roles) is injected into the request context.

## Frontend Implementation

We use `next-auth` (or equivalent OIDC client library) to handle the redirect flow.

```tsx
// Example using NextAuth
import NextAuth from "next-auth";
import ZitadelProvider from "next-auth/providers/zitadel";

export const authOptions = {
  providers: [
    ZitadelProvider({
      clientId: process.env.ZITADEL_CLIENT_ID,
      clientSecret: process.env.ZITADEL_CLIENT_SECRET,
      issuer: process.env.ZITADEL_ISSUER,
    }),
  ],
  callbacks: {
    async jwt({ token, account }) {
      if (account) {
        token.accessToken = account.access_token;
      }
      return token;
    },
    async session({ session, token }) {
      session.accessToken = token.accessToken;
      return session;
    },
  },
};
```

## Backend Validation

The backend uses a middleware to validate tokens on every request.

```go
// internal/auth/middleware.go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tokenString := r.Header.Get("Authorization")
        // ... extract token ...
        
        claims, err := validator.ValidateToken(r.Context(), tokenString)
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        // Add user to context
        ctx := context.WithValue(r.Context(), UserKey, claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## Security Considerations

-   **HTTPS**: All traffic MUST be encrypted in production.
-   **Token Storage**: Tokens should be stored securely (e.g., `httpOnly` cookies or in memory), never in `localStorage` if possible to avoid XSS.
-   **Scopes**: Always request the minimum required scopes (`openid profile email`).
-   **Rotation**: Use refresh tokens to maintain sessions securely without long-lived access tokens.
