# Coding Standards

Vortyx enforces strict coding standards to ensure maintainability, type safety, and consistency across the team.

## General Principles

-   **Type Safety First**: Avoid `any` in TypeScript or `interface{}` in Go unless absolutely necessary.
-   **Explicit > Implicit**: Favor verbose, descriptive variable names.
-   **No Magic Numbers**: Use constants or configuration values.
-   **Error Handling**: Always handle errors explicitly. In Go, never ignore the error return (`_`).

## Go (Backend)

-   **Formatting**: Use `gofmt` and `goimports`.
-   **Linter**: Use `golangci-lint` with default settings.
-   **Project Layout**: Follow Standard Go Project Layout (`cmd/`, `internal/`, `pkg/`).
-   **Dependency Injection**: Use constructor injection (`NewService(deps)`) rather than global variables.
-   **SQL Access**: Never write raw SQL strings in Go code. Use `sqlc` with `.sql` files.
-   **Testing**: Write table-driven tests (`t.Run`) for all business logic.

### Example (Good):
```go
func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    user, err := s.repo.GetUser(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return user, nil
}
```

## TypeScript (Frontend)

-   **Formatting**: Prettier (standard config).
-   **Linter**: ESLint with `next/core-web-vitals` and TypeScript strict rules.
-   **Naming**: PascalCase for components (`UserProfile.tsx`), camelCase for variables/functions (`getUser`).
-   **Hooks**: Custom hooks should start with `use`.
-   **Data Fetching**: Use generated ConnectRPC clients (`useQuery` integration recommended).
-   **Styles**: Use Tailwind CSS utility classes. Avoid inline styles.

### Example (Good):
```tsx
const UserProfile = ({ userId }: { userId: string }) => {
  const { data, isLoading } = useQuery(['user', userId], () => client.getUser({ id: userId }));
  
  if (isLoading) return <Spinner />;
  return <div className="p-4 bg-white rounded">{data?.name}</div>;
};
```

## Protocol Buffers (API)

-   **Style**: Follow the [Buf Style Guide](https://buf.build/docs/style-guide).
-   **Naming**: `snake_case` for fields, `PascalCase` for messages/services.
-   **Versioning**: All packages must be versioned (e.g., `vortyx.pulse.v1`).
-   **Documentation**: All RPCs and Messages must have comments.

## Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

-   `feat: add new dashboard widget`
-   `fix: resolve null pointer in auth middleware`
-   `docs: update local setup guide`
-   `chore: upgrade dependencies`
