# Vortyx Documentation

Welcome to the official documentation for **Vortyx**, a next-generation "Unified Monolith" platform designed to revolutionize MSP (Managed Service Provider) and MSSP (Managed Security Service Provider) operations.

## Mission & Vision

**Vortyx** aims to consolidate disparate IT management and security tools into a single, cohesive ecosystem. By leveraging a high-performance, type-safe architecture, we provide a unified control plane that eliminates context switching, reduces operational overhead, and enhances security posture for IT professionals.

Our vision is to become the operating system for modern IT service delivery, where monitoring, automation, security, and documentation live in harmony.

## Documentation Structure

This documentation is organized to guide you from high-level concepts to deep technical implementation details.

### 🏗️ [Architecture](./architecture/system_overview.md)
Deep dive into the "Unified Monolith" design, service boundaries, data flow, and technology choices.
- **[System Overview](./architecture/system_overview.md)**: The definitive map of the Vortyx ecosystem.

### 💻 [Development](./development/local_setup.md)
Everything you need to start contributing code.
- **[Local Setup](./development/local_setup.md)**: Step-by-step guide to running Vortyx locally with Docker and Go.
- **[Coding Standards](./development/coding_standards.md)**: Guidelines for Go, TypeScript, and Protobuf to maintain code quality.

### 🔌 [API & Integration](./api/authentication.md)
Technical details on how our services communicate and how to integrate with them.
- **[Authentication](./api/authentication.md)**: Comprehensive guide to OIDC/JWT auth flow with Zitadel.

### 🚀 [Deployment & Ops](./deployment/infrastructure.md)
Guides for running Vortyx in production environments.
- **[Infrastructure](./deployment/infrastructure.md)**: Server configuration and scaling strategies.
- **[CI/CD Pipeline](./deployment/ci_cd_pipeline.md)**: Workflows for testing and deployment.
- **[Environment Variables](./deployment/environment_variables.md)**: Reference for all configuration knobs.

---

## Quick Links

- **Repository**: [github.com/abdul/vortyx](https://github.com/abdul/vortyx)
- **Frontend**: [http://localhost:3000](http://localhost:3000) (Local Dev)
- **Backend**: [http://localhost:8081](http://localhost:8081) (Local Dev)
- **Identity Provider**: [http://localhost:8080](http://localhost:8080) (Zitadel Console)
