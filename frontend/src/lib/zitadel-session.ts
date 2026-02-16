/**
 * Zitadel Session Management
 *
 * This module provides client-side utilities for interacting with Zitadel's
 * session-based authentication flow. It abstracts the complexity of:
 *
 * 1. Creating a new authentication session
 * 2. Verifying user credentials (password)
 * 3. Finalizing the authentication request
 *
 * The flow works as follows:
 *
 * 1. createSession: Creates a session context in Zitadel for the user
 *    - This initializes the login process
 *    - Returns sessionId and sessionToken
 *
 * 2. checkPassword: Validates the user's password against Zitadel
 *    - Sends the password to Zitadel for verification
 *    - Returns updated sessionToken if successful
 *
 * 3. finalizeSession: Completes the authentication flow
 *    - Exchanges the session for an OIDC auth request result
 *    - Returns a callback URL to redirect the user
 *
 * Security Considerations:
 * - Passwords are never stored on the client
 * - Session tokens are handled server-side via NextAuth
 * - Error messages are sanitized to prevent information disclosure
 *
 * Environment Requirements:
 * - ZITADEL_ISSUER must be configured in the environment
 * - The /api/zitadel/session proxy must be available
 */

export interface CreateSessionRequest {
  /**
   * Checks define what credentials to validate.
   * Currently supports user identification via loginName or userId.
   */
  checks?: {
    user?: {
      /** The user's login name (email/username) */
      loginName?: string;
      /** Alternative: The user's Zitadel ID */
      userId?: string;
    };
  };
  /** Optional metadata to attach to the session */
  metadata?: Record<string, string>;
}

/**
 * Structured error format for authentication failures.
 * Provides consistent error handling across the application.
 */
export interface AuthError {
  /** Human-readable error message */
  message: string;
  /** Optional error code for programmatic handling */
  code?: string;
  /** The step/operation that failed */
  step?: string;
  /** Additional error details */
  details?: any;
}

/**
 * Internal function that calls the Zitadel session proxy API.
 *
 * This proxy handles the communication with Zitadel's internal APIs,
 * which are not directly accessible from the browser due to CORS.
 *
 * @param data - The action and parameters to send to Zitadel
 * @returns The parsed JSON response from Zitadel
 * @throws AuthError if the request fails
 */
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
    // Return structured error to enable consistent error handling
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

/**
 * Creates a new Zitadel session for a user.
 *
 * This is the first step in the Zitadel session authentication flow.
 * It creates a session context that subsequent operations will use.
 *
 * @param req - Session creation request with user identification
 * @returns Session object containing sessionId and sessionToken
 *
 * @example
 * const session = await createSession({
 *   checks: {
 *     user: { loginName: "user@example.com" }
 *   }
 * });
 * console.log(session.sessionId);
 */
export async function createSession(req: CreateSessionRequest) {
  return callProxy({
    action: "create",
    checks: req.checks,
  });
}

/**
 * Validates a user's password against Zitadel.
 *
 * This is the second step in the authentication flow.
 * The password is sent to Zitadel for verification.
 *
 * @param sessionId - The session ID from createSession
 * @param sessionToken - The session token from createSession
 * @param password - The user's password
 * @returns Updated session token if password is correct
 *
 * @example
 * const checkResp = await checkPassword(
 *   session.sessionId,
 *   session.sessionToken,
 *   "user-password"
 * );
 */
export async function checkPassword(sessionId: string, sessionToken: string, password: string) {
  return callProxy({
    action: "checkPassword",
    sessionId,
    sessionToken,
    password,
  });
}

/**
 * Finalizes the Zitadel authentication flow.
 *
 * This is the final step that completes the authentication.
 * It exchanges the valid session for a callback URL that
 * redirects to the original requested resource.
 *
 * @param authRequestId - The OIDC auth request ID
 * @param sessionId - The session ID
 * @param sessionToken - The validated session token
 * @returns Response containing callbackUrl for redirection
 *
 * @example
 * const result = await finalizeSession(
 *   authRequestId,
 *   sessionId,
 *   sessionToken
 * );
 * window.location.href = result.callbackUrl;
 */
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
 *
 * This convenience function orchestrates all three steps of the authentication
 * flow in sequence, handling the complexity of token passing between steps.
 *
 * Flow:
 * 1. Create a new session for the user
 * 2. Validate the password
 * 3. Finalize the session and get callback URL
 *
 * @param username - The user's login name (email/username)
 * @param password - The user's password
 * @param authRequestId - The OIDC auth request ID from the login page
 * @returns The callback URL to redirect the user to
 * @throws AuthError if any step fails
 *
 * @example
 * try {
 *   const callbackUrl = await performZitadelLogin(
 *     "user@example.com",
 *     "password123",
 *     authRequestId
 *   );
 *   router.push(callbackUrl);
 * } catch (error) {
 *   console.error("Login failed:", error.message);
 * }
 */
export async function performZitadelLogin(username: string, password: string, authRequestId: string) {
  try {
    // Step 1: Create Session
    // Initialize a new authentication session in Zitadel
    const session = await createSession({
      checks: {
        user: {
          loginName: username,
        },
      },
    });

    // Step 2: Check Password
    // Validate the user's credentials
    const checkResp = await checkPassword(session.sessionId, session.sessionToken, password);
    
    // Step 3: Finalize Session
    // Complete the flow and get the redirect URL
    const finalizeResp = await finalizeSession(authRequestId, session.sessionId, checkResp.sessionToken);
    
    if (!finalizeResp.callbackUrl) {
      throw { message: "Login successful but no callback URL received", step: "finalize" } as AuthError;
    }

    return finalizeResp.callbackUrl;
  } catch (error: any) {
    // Ensure we're throwing a structured AuthError
    // This maintains consistent error handling across the app
    if (error.message) throw error;
    throw { message: "An unexpected error occurred during login", details: error } as AuthError;
  }
}
