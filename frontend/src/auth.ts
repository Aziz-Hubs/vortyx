import NextAuth from "next-auth"
import Zitadel from "next-auth/providers/zitadel"
import Credentials from "next-auth/providers/credentials"

export const { handlers, auth, signIn, signOut } = NextAuth({
  providers: [
    Zitadel({
      issuer: process.env.ZITADEL_ISSUER || "http://localhost:8080",
      clientId: process.env.ZITADEL_CLIENT_ID || "vortyx-frontend",
      clientSecret: process.env.ZITADEL_CLIENT_SECRET || "",
      authorization: { params: { scope: "openid email profile offline_access" } },
    }),
    Credentials({
      name: "Credentials",
      credentials: {
        username: { label: "Username", type: "text" },
        password: { label: "Password", type: "password" },
      },
      async authorize(credentials) {
        console.log("[CREDENTIALS] Authorize called with:", credentials?.username);

        if (!credentials?.username || !credentials?.password) {
          console.error("[CREDENTIALS] Missing username or password");
          throw new Error("Missing credentials");
        }

        try {
          const issuer = process.env.ZITADEL_ISSUER || "http://localhost:8080";
          const tokenEndpoint = `${issuer}/oauth/v2/token`;
          
          console.log("[CREDENTIALS] Token endpoint:", tokenEndpoint);
          console.log("[CREDENTIALS] Client ID:", process.env.ZITADEL_CLIENT_ID);

          const params = new URLSearchParams();
          params.append("grant_type", "password");
          params.append("client_id", process.env.ZITADEL_CLIENT_ID || "");
          params.append("client_secret", process.env.ZITADEL_CLIENT_SECRET || "");
          params.append("username", credentials.username as string);
          params.append("password", credentials.password as string);
          params.append("scope", "openid email profile offline_access");

          const res = await fetch(tokenEndpoint, {
            method: "POST",
            headers: { "Content-Type": "application/x-www-form-urlencoded" },
            body: params,
          });

          const data = await res.json();
          console.log("[CREDENTIALS] Zitadel response status:", res.status);
          console.log("[CREDENTIALS] Zitadel response:", JSON.stringify(data));

          if (!res.ok) {
            console.error("[CREDENTIALS] Zitadel login failed:", data);
            // Throw a proper error with details instead of returning null
            throw new Error(`Zitadel login failed: ${JSON.stringify(data)}`);
          }

          // Get user info
          const userRes = await fetch(`${issuer}/oidc/v1/userinfo`, {
            headers: { Authorization: `Bearer ${data.access_token}` },
          });
          const user = await userRes.json();
          console.log("[CREDENTIALS] User info:", JSON.stringify(user));

          return {
            id: user.sub,
            name: user.name,
            email: user.email,
            accessToken: data.access_token,
            idToken: data.id_token,
          };
        } catch (error) {
          console.error("[CREDENTIALS] Auth error:", error);
          // Throw to get proper error message
          throw error;
        }
      },
    }),
  ],
  callbacks: {
    async jwt({ token, account, user }) {
      if (account) {
        token.accessToken = account.access_token
        token.idToken = account.id_token
      }
      if (user) {
        // @ts-expect-error extending token
        if (user.accessToken) token.accessToken = user.accessToken;
        // @ts-expect-error extending token
        if (user.idToken) token.idToken = user.idToken;
      }
      return token
    },
    async session({ session, token }) {
      // @ts-expect-error extending session type
      session.accessToken = token.accessToken
      // @ts-expect-error extending session type
      session.idToken = token.idToken
      return session
    },
  },
})
