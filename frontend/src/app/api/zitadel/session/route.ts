import { NextResponse } from "next/server";

const ZITADEL_API = process.env.ZITADEL_ISSUER || "http://localhost:8080";

function getZitadelPAT(): string {
  const pat = process.env.ZITADEL_PAT?.trim();
  if (!pat) {
    throw new Error("ZITADEL_PAT environment variable is not set. Authentication proxy unavailable.");
  }
  if (pat.startsWith("ZaUt")) {
    throw new Error("ZITADEL_PAT is invalid or expired. Please provide a valid Personal Access Token.");
  }
  return pat;
}

export async function POST(req: Request) {
  let ZITADEL_PAT: string;
  try {
    ZITADEL_PAT = getZitadelPAT();
  } catch (err: any) {
    console.error("[PROXY] Configuration error:", err.message);
    return NextResponse.json(
      { error: "Authentication service misconfigured", code: "CONFIG_ERROR" },
      { status: 503 }
    );
  }

  try {
    const body = await req.json();
    const { action, sessionId, sessionToken, authRequestId, checks, password } = body;
    
    if (!action) {
      return NextResponse.json({ error: "Missing action" }, { status: 400 });
    }

    const cleanSessionToken = sessionToken?.trim();

    let url = "";
    let method = "POST";
    let headers: Record<string, string> = {
      "Content-Type": "application/json",
      "Accept": "application/json",
    };
    let requestBody: any = {};

    // Validate requirements for each action
    switch (action) {
      case "create":
        if (!checks?.user?.loginName) {
          return NextResponse.json({ error: "Missing login name for session creation" }, { status: 400 });
        }
        url = `${ZITADEL_API}/v2/sessions`;
        headers["Authorization"] = `Bearer ${ZITADEL_PAT}`;
        requestBody = { checks };
        break;
      case "checkPassword":
        if (!sessionId || !cleanSessionToken || !password) {
          return NextResponse.json({ error: "Missing required fields for password check" }, { status: 400 });
        }
        url = `${ZITADEL_API}/v2/sessions/${sessionId}`;
        method = "PATCH";
        headers["Authorization"] = `Bearer ${ZITADEL_PAT}`;
        headers["x-zitadel-session"] = cleanSessionToken;
        requestBody = { 
          checks: {
            password: { password }
          }
        };
        break;
      case "finalize":
        if (!authRequestId || !sessionId || !cleanSessionToken) {
          return NextResponse.json({ error: "Missing required fields for finalization" }, { status: 400 });
        }
        url = `${ZITADEL_API}/v2/oidc/auth_requests/${authRequestId}`;
        headers["Authorization"] = `Bearer ${ZITADEL_PAT}`;
        requestBody = {
          session: {
            sessionId,
            sessionToken: cleanSessionToken
          }
        };
        break;
      default:
        return NextResponse.json({ error: `Invalid action: ${action}` }, { status: 400 });
    }

    console.log(`[PROXY] Calling Zitadel ${action} (${method} ${url})`);

    const response = await fetch(url, {
      method,
      headers,
      body: JSON.stringify(requestBody),
    });

    const data = await response.json();
    
    if (!response.ok) {
      console.error(`[PROXY] Zitadel ${action} failed [${response.status}]:`, JSON.stringify(data, null, 2));
      
      // Map Zitadel errors to more descriptive internal ones if possible
      const errorMessage = data.message || data.error || "Zitadel API error";
      const errorCode = data.code || "UNKNOWN_ERROR";
      
      return NextResponse.json({ 
        error: errorMessage, 
        code: errorCode,
        details: data.details || [],
        step: action
      }, { status: response.status });
    }

    return NextResponse.json(data);
  } catch (error: any) {
    console.error(`[PROXY] Internal Error during ${req.method} /api/zitadel/session:`, error);
    return NextResponse.json({ 
      error: "Internal server error in auth proxy",
      details: error.message
    }, { status: 500 });
  }
}
