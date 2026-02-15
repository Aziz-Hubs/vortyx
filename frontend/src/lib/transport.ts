import { createConnectTransport } from "@connectrpc/connect-web";
import { getSession } from "next-auth/react";

export const transport = createConnectTransport({
  baseUrl: "http://localhost:8081",
  interceptors: [
    (next) => async (req) => {
      if (typeof window !== "undefined") {
        try {
          const session = await getSession();
          // @ts-expect-error extending session type
          if (session?.accessToken) {
            // @ts-expect-error extending session type
            req.header.set("Authorization", `Bearer ${session.accessToken}`);
          }
        } catch (e) {
          console.error("Auth error", e);
        }
      }
      return next(req);
    },
  ],
});
