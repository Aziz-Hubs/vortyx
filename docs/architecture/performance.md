# Performance Optimization

This document outlines guidelines, benchmarks, and tuning parameters for high-performance operation.

## 1. Benchmarking Targets

| Component | Metric | Target | Notes |
| :--- | :--- | :--- | :--- |
| **API Latency** | P95 | < 50ms | Excl. external integrations |
| **Throughput** | RPS | 10k | Single Backend instance |
| **Ingestion** | Events/sec | 100k | TimescaleDB bulk insert |
| **Startup** | Time | < 2s | Backend cold start |

## 2. Go (Backend) Tuning

-   **Memory**: Set `GOMEMLIMIT` to 90% of container memory limit.
-   **GC**: Set `GOGC` based on memory pressure (default 100).
-   **Concurrency**: Use `errgroup` or worker pools for parallel tasks (e.g., scanning 1000 IPs). Avoid unbounded goroutines.
-   **ConnectRPC**: Enable compression (`gzip`, `brotli`) for large payloads (e.g., log batches).

## 3. Database (Postgres/TimescaleDB) Tuning

-   **Connections**: Use `pgxpool` with `MinConns=5`, `MaxConns=50` (adjust based on CPU cores).
-   **Bulk Inserts**: Always use `COPY` protocol (via `pgx.CopyFrom`) for high-volume ingestion (Pulse/Radar logs).
-   **Indexing**: Use BRIN indexes for time-series data (smaller, faster scans).
-   **Vacuum**: Tune autovacuum for high-churn tables (`log_entries`).

## 4. Frontend (Next.js) Tuning

-   **Image Optimization**: Use `next/image` with WebP/AVIF.
-   **Code Splitting**: Dynamic imports (`next/dynamic`) for heavy components (e.g., RustDesk player, Charts).
-   **Caching**: Implement `stale-while-revalidate` using React Query.
-   **Bundle Size**: Keep initial JS load < 100KB (gzipped).

## 5. Network Optimization

-   **HTTP/2**: Enabled end-to-end (Frontend -> LB -> Backend).
-   **CDN**: Cache static assets (JS, CSS, Images) at the edge.
-   **Compression**: Enable Brotli compression on the Load Balancer/Reverse Proxy.
