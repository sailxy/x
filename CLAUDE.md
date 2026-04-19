# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go library (github.com/sailxy/x) that provides utilities and integrations for common cloud services, databases, and developer tools. It's designed to be used as a dependency in Go applications.

## Common Commands

### Testing
```bash
# Run all tests
go test ./...

# Run tests in a specific package
go test ./jwt

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...
```

### Dependency Management
```bash
# Tidy up dependencies
go mod tidy

# Download dependencies
go mod download

# Verify dependencies
go mod verify
```

### Security Checks
```bash
# Run gitleaks to check for secrets
gitleaks detect --source .

# Run pre-commit hooks
pre-commit run --all-files
```

### Build
```bash
# Build the module
go build ./...
```

## Architecture Patterns

### Package Structure
The codebase is organized into functional packages, each with a specific purpose:

- **id/**: Unique ID generation (UUID, XID, Snowflake, NanoID, Random)
- **config/**: Configuration file loading using Viper
- **logger/**: Zap-based structured logging with file rotation
- **database/gorm/**: GORM ORM wrapper with MySQL/PostgreSQL support and transaction context management
- **database/redis/**: Redis client wrapper
- **cache/**: Redis-based caching using gocache
- **locker/**: Distributed locking using Redis
- **tracer/**: OpenTelemetry tracing (stdout and HTTP exporters)
- **jwt/**: JWT token generation and parsing with json.Number support for precision
- **aws/s3**: S3 client with presigned URL support
- **aws/sqs**: SQS client wrapper
- **aliyun/oss**: Aliyun OSS client
- **aliyun/sms**: Aliyun SMS service
- **aliyun/pns**: Aliyun PNS (Phone Number Service)
- **apple/pay**: Apple Pay and In-App Purchase verification
- **wechat/**: Wechat Open Platform integration
- **qq/**: QQ Connect Platform integration
- **rest/**: HTTP client using resty.dev/v3
- **email/smtp**: SMTP email sending
- **password**: Password hashing utilities
- **env/**: Environment variable loading with .env support
- **errtrace/**: Error stack trace utilities
- **faker/**: Fake data generation using gofakeit
- **printer/**: Pretty printing utilities
- **cast/**: Type casting utilities
- **util/arrutil**: Array utilities
- **util/fsutil**: File system utilities

### Design Patterns

**Configuration Pattern**: Each package typically has a `Config` struct and a `New()` constructor:

```go
type Config struct {
    // Configuration fields
}

func New(c Config) *Type {
    // Initialize and return instance
}
```

**Context Usage**: Database and cache operations accept `context.Context` for request-scoped values:

- `database/gorm` uses context to pass transactions (see `transaction.go`)
- `logger` uses context to pass trace IDs (see `logger.SetTraceID()`)
- Cache operations require context

**Error Handling**: Consistent error wrapping with `%w`:

```go
return nil, fmt.Errorf("operation failed: %w", err)
```

**Transaction Pattern**: GORM transactions are managed through context:

```go
tx := gorm.NewTx(db)
err := tx.Exec(ctx, func(ctx context.Context) error {
    // Use ctx in database operations to participate in transaction
    db := tx.GetTx(ctx)
    return db.Create(&model).Error
})
```

**Type Aliases**: Packages often export type aliases for underlying types to simplify imports:

```go
type DB = gorm.DB
type Client = redis.Client
```

**Implementation Style**: Prefer direct, local logic over extra indirection when the abstraction does not clearly improve reuse or readability:

- Prefer direct field assignment and inline argument passing when a value is only used once
- Avoid package-level indirection added only to make tests easier
- Avoid trivial helper methods like `x.params()` when writing the values inline is clearer
- Keep wrappers thin and obvious; if a helper exists, it should carry real semantic value
- When boolean flags affect behavior, prefer explicit caller-provided values over implicit defaults or pointer-bool config fields

### JWT Implementation Notes

The JWT package uses `jwt.WithJSONNumber()` to preserve numeric precision when parsing tokens. This prevents Go's default JSON parser from converting numbers to float64. Use `json.Number` type assertions when extracting numeric values from claims.

### Testing

- Tests use `github.com/stretchr/testify/assert` for assertions
- Test files follow the pattern `*_test.go` in the same package
- Use table-driven tests for multiple test cases (see `jwt_test.go`)
- For gitleaks compliance, use `// gitleaks:allow` comments on lines with test data that resembles secrets

### Pre-commit Hooks

The project uses pre-commit with gitleaks for secret detection. Configure hooks in `.pre-commit-config.yaml`. Hooks run automatically before commits.

### Go Version

This project requires Go 1.23.0+ with toolchain go1.24.1.

### Dependencies

Key third-party libraries used:
- `github.com/golang-jwt/jwt/v5` - JWT handling
- `github.com/spf13/viper` - Configuration management
- `go.uber.org/zap` - Structured logging
- `gorm.io/gorm` - ORM
- `github.com/redis/go-redis/v9` - Redis client
- `github.com/aws/aws-sdk-go-v2/*` - AWS SDK
- `go.opentelemetry.io/*` - OpenTelemetry tracing
- `resty.dev/v3` - HTTP client
- `github.com/stretchr/testify` - Testing assertions
