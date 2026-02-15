export interface CreateSessionRequest {
  checks?: {
    user?: {
      loginName?: string;
      userId?: string;
    };
  };
  metadata?: Record<string, string>;
}

export interface AuthError {
  message: string;
  code?: string;
  step?: string;
  details?: any;
}

async function callProxy(data: any) {
  const res = await fetch("/api/zitadel/session", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  });

  const result = await res.json();
  if (!res.ok) {
    // Return structured error
    const error: AuthError = {
      message: result.error || result.message || "Request failed",
      code: result.code,
      step: result.step || data.action,
      details: result.details
    };
    throw error;
  }
  return result;
}

export async function createSession(req: CreateSessionRequest) {
  return callProxy({
    action: "create",
    checks: req.checks,
  });
}

export async function checkPassword(sessionId: string, sessionToken: string, password: string) {
  return callProxy({
    action: "checkPassword",
    sessionId,
    sessionToken,
    password,
  });
}

export async function finalizeSession(authRequestId: string, sessionId: string, sessionToken: string) {
  return callProxy({
    action: "finalize",
    sessionId,
    sessionToken,
    authRequestId,
  });
}

/**
 * High-level login function that handles the entire Zitadel session flow.
 * Useful for reusing across different modules.
 */
export async function performZitadelLogin(username: string, password: string, authRequestId: string) {
  try {
    // 1. Create Session
    const session = await createSession({
      checks: {
        user: {
          loginName: username,
        },
      },
    });

    // 2. Check Password
    const checkResp = await checkPassword(session.sessionId, session.sessionToken, password);
    
    // 3. Finalize and get callback URL
    const finalizeResp = await finalizeSession(authRequestId, session.sessionId, checkResp.sessionToken);
    
    if (!finalizeResp.callbackUrl) {
      throw { message: "Login successful but no callback URL received", step: "finalize" } as AuthError;
    }

    return finalizeResp.callbackUrl;
  } catch (error: any) {
    // Ensure we're throwing a structured AuthError
    if (error.message) throw error;
    throw { message: "An unexpected error occurred during login", details: error } as AuthError;
  }
}
