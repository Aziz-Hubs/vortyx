# Troubleshooting Guide

This document provides resolutions for common issues encountered in Vortyx development and production.

## 1. Port Conflicts

**Symptom**: `bind: address already in use` or services failing to start.

**Resolution**:
-   **Check Ports**:
    -   Backend: `8081` (default)
    -   Zitadel: `8080`
    -   Postgres: `5432`
    -   Frontend: `3000`
-   **Fix**: Update `.env` or kill the conflicting process (`lsof -i :<port>`).
-   **Docker**: Check `docker ps` for container port mappings.

## 2. Database Connection Errors

**Symptom**: `dial tcp [::1]:5432: connect: connection refused` or `pq: password authentication failed`.

**Resolution**:
-   **Check DB**: Is the `db` container running? (`docker-compose ps`)
-   **Check Credentials**: Do `POSTGRES_USER`/`PASSWORD` in `.env` match `docker-compose.yml`?
-   **Check Host**: In Docker, use `db` hostname. Local dev uses `localhost`.

## 3. Code Generation Issues

**Symptom**: `import cycle not allowed` or `undefined: PulseService`.

**Resolution**:
-   **Regenerate**: Run `task gen` (or `buf generate`).
-   **Check Proto**: Ensure `.proto` files are valid (`buf lint`).
-   **Update Buf**: `buf update` to pull latest dependencies.

## 4. Authentication Failures

**Symptom**: `401 Unauthorized` or infinite redirect loop on login.

**Resolution**:
-   **Check Token**: Inspect the `Authorization` header in browser DevTools.
-   **Check Issuer**: Ensure `ZITADEL_ISSUER` matches exactly (e.g., `http` vs `https`).
-   **Check Client ID**: Verify `NEXT_PUBLIC_ZITADEL_CLIENT_ID` in Zitadel Console.
-   **Check Clock**: Ensure server time is synced (NTP). JWT `exp`/`nbf` claims rely on accurate clocks.

## 5. Docker Issues

**Symptom**: `context canceled` or container exiting immediately.

**Resolution**:
-   **Check Logs**: `docker-compose logs <service>` to see the crash reason.
-   **Prune**: `docker system prune` to clear stale images/networks.
-   **Rebuild**: `docker-compose up --build` to force a rebuild.
