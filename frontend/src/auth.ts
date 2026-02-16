import NextAuth from "next-auth"
import Zitadel from "next-auth/providers/zitadel"
import Credentials from "next-auth/providers/credentials"

const baseScope = "openid email profile offline_access"
const apiProjectID = process.env.ZITADEL_API_PROJECT_ID
const audienceScope = apiProjectID
  ? `urn:zitadel:iam:org:project:id:${apiProjectID}:aud`
  : ""
const scope = [baseScope, audienceScope].filter(Boolean).join(" ")

export const { handlers, auth, signIn, signOut } = NextAuth({
  providers: [
    Zitadel({
      issuer: process.env.ZITADEL_ISSUER || "http://localhost:8080",
      clientId: process.env.ZITADEL_CLIENT_ID || "vortyx-frontend",
      clientSecret: process.env.ZITADEL_CLIENT_SECRET || "",
      authorization: { params: { scope } },
    }),
    ...(process.env.ZITADEL_ENABLE_PASSWORD_GRANT === "true"
      ? [
          Credentials({
            name: "Credentials",
            credentials: {
              username: { label: "Username", type: "text" },
              password: { label: "Password", type: "password" },
            },
            async authorize(credentials) {
              if (!credentials?.username || !credentials?.password) {
                throw new Error("Missing credentials")
              }

              const issuer = process.env.ZITADEL_ISSUER || "http://localhost:8080"
              const tokenEndpoint = `${issuer}/oauth/v2/token`

              const params = new URLSearchParams()
              params.append("grant_type", "password")
              params.append("client_id", process.env.ZITADEL_CLIENT_ID || "")
              params.append("client_secret", process.env.ZITADEL_CLIENT_SECRET || "")
              params.append("username", credentials.username as string)
              params.append("password", credentials.password as string)
              params.append("scope", scope)

              const res = await fetch(tokenEndpoint, {
                method: "POST",
                headers: { "Content-Type": "application/x-www-form-urlencoded" },
                body: params,
              })

              const data = await res.json()
              if (!res.ok) {
                throw new Error("Zitadel login failed")
              }

              const userRes = await fetch(`${issuer}/oidc/v1/userinfo`, {
                headers: { Authorization: `Bearer ${data.access_token}` },
              })
              const user = await userRes.json()

              return {
                id: user.sub,
                name: user.name,
                email: user.email,
                accessToken: data.access_token,
                idToken: data.id_token,
              }
            },
          }),
        ]
      : []),
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
