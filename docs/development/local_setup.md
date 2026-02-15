# Local Development Setup

This guide covers setting up the Vortyx development environment from scratch.

## Prerequisites

Ensure you have the following installed:
-   **Go**: v1.22+ ([Download](https://go.dev/dl/))
-   **Node.js**: v20+ (LTS) ([Download](https://nodejs.org/))
-   **Docker & Docker Compose**: ([Download](https://www.docker.com/products/docker-desktop/))
-   **Task**: Task runner ([Install](https://taskfile.dev/installation/))
-   **Buf**: Protobuf tool ([Install](https://buf.build/docs/installation))
-   **SQLC**: SQL Compiler ([Install](https://docs.sqlc.dev/en/latest/overview/install.html))

## 1. Clone & Configure

```bash
git clone https://github.com/abdul/vortyx.git
cd vortyx

# Create .env file from example (ensure this file exists or create it)
cp .env.example .env 
# OR manually create .env with:
# PORT=8081
# DATABASE_URL=postgres://postgres:postgres@localhost:5432/vortyx?sslmode=disable
```

## 2. Infrastructure Startup

Start the supporting services (PostgreSQL, Zitadel, RustDesk):

```bash
task up
# OR
docker-compose up -d
```

> **Note**: Wait a few moments for Zitadel and Postgres to become healthy.

## 3. Backend Setup

Generate code and start the Go server:

```bash
# 1. Generate Protobuf and SQL code
task gen

# 2. Start the Backend Server
task dev:backend
# OR
cd backend && go run .
```

The backend should be running at `http://localhost:8081`.

## 4. Frontend Setup

Install dependencies and start the Next.js server:

```bash
cd frontend
npm install
npm run dev
```

The frontend should be running at `http://localhost:3000`.

## 5. Verification

1.  Open your browser to `http://localhost:3000`.
2.  Navigate to the **Pulse** dashboard.
3.  If you see "System Status: Operational", the connection to the backend is successful.

## Troubleshooting

-   **Port Conflicts**: Ensure ports `8080` (Zitadel), `8081` (Backend), and `5432` (Postgres) are free.
-   **Database Errors**: Check `docker-compose logs db` to ensure the database initialized correctly.
-   **Code Gen Errors**: Run `buf update` in the `proto` directory if you encounter Protobuf dependency issues.
