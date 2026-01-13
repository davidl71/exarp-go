# Database Abstraction Layer

## Overview

The database layer has been abstracted to support multiple database backends:
- **SQLite** (default) - File-based, no server required
- **MySQL** - Popular relational database
- **PostgreSQL** - Advanced open-source database
- **ODBC** - Generic ODBC support (optional, via cgo)

## Architecture

### Driver Interface

All database drivers implement the `Driver` interface:

```go
type Driver interface {
    Type() DriverType
    Open(dsn string) (*sql.DB, error)
    Configure(db *sql.DB) error
    Dialect() Dialect
    Close() error
}
```

### SQL Dialect Abstraction

Each driver provides a `Dialect` that handles SQL differences:

- **Timestamp functions**: `strftime()` vs `UNIX_TIMESTAMP()` vs `EXTRACT(EPOCH FROM NOW())`
- **Placeholders**: `?` vs `$1, $2, ...`
- **Data types**: `INTEGER` vs `BIGINT`, `TEXT` vs `JSONB`
- **Auto-increment**: `AUTOINCREMENT` vs `AUTO_INCREMENT` vs `SERIAL`
- **Boolean types**: `INTEGER` (0/1) vs `TINYINT(1)` vs `BOOLEAN`
- **Row-level locking**: Support for `SELECT ... FOR UPDATE`

## Configuration

### Environment Variables

```bash
# Database driver (sqlite, mysql, postgres, odbc)
export DB_DRIVER=postgres

# Database connection string (DSN)
export DB_DSN="postgres://user:password@localhost:5432/todo2?sslmode=disable"

# Auto-migrate on startup (default: true)
export DB_AUTO_MIGRATE=true
```

### Programmatic Configuration

```go
import "github.com/davidl71/exarp-go/internal/database"

// Load config from environment
cfg, err := database.LoadConfig(projectRoot)

// Or create custom config
cfg := &database.Config{
    Driver:      database.DriverPostgres,
    DSN:         "postgres://user:pass@localhost/todo2",
    AutoMigrate: true,
}

// Initialize database
err := database.InitWithConfig(cfg)
```

## Driver-Specific DSN Formats

### SQLite
```
# File path
.todo2/todo2.db

# URI format
file:todo2.db?mode=rwc
```

### MySQL
```
user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=UTC
```

### PostgreSQL
```
postgres://user:password@host:port/dbname?sslmode=disable
```

### ODBC
```
# DSN name (configured in odbc.ini)
DSN=todo2

# Full connection string
Driver={PostgreSQL};Server=localhost;Database=todo2;UID=user;PWD=password

# MySQL ODBC
Driver={MySQL ODBC 8.0 Driver};Server=localhost;Database=todo2;UID=user;PWD=password

# SQL Server ODBC
Driver={ODBC Driver 17 for SQL Server};Server=localhost;Database=todo2;UID=user;PWD=password
```

**Note:** ODBC requires cgo to be enabled. Build with:
```bash
CGO_ENABLED=1 go build
# or
go build -tags cgo
```

## Usage Examples

### Using SQLite (Default)

```go
// No configuration needed - uses default SQLite
err := database.Init(projectRoot)
```

### Using MySQL

```go
// Set environment variables
os.Setenv("DB_DRIVER", "mysql")
os.Setenv("DB_DSN", "user:password@tcp(localhost:3306)/todo2")

// Or programmatically
cfg := &database.Config{
    Driver: database.DriverMySQL,
    DSN:    "user:password@tcp(localhost:3306)/todo2",
}
err := database.InitWithConfig(cfg)
```

### Using PostgreSQL

```go
cfg := &database.Config{
    Driver: database.DriverPostgres,
    DSN:    "postgres://user:password@localhost:5432/todo2?sslmode=disable",
}
err := database.InitWithConfig(cfg)
```

### Using Dialect for SQL Differences

```go
// Get current dialect
dialect, err := database.GetDialect()
if err != nil {
    return err
}

// Use dialect-specific functions
timestampSQL := dialect.CurrentTimestamp()
// SQLite: "strftime('%s', 'now')"
// MySQL: "UNIX_TIMESTAMP()"
// PostgreSQL: "EXTRACT(EPOCH FROM NOW())"

// Use dialect-specific placeholders
placeholder1 := dialect.Placeholder(1)
placeholder2 := dialect.Placeholder(2)
// SQLite/MySQL: "?", "?"
// PostgreSQL: "$1", "$2"
```

## SQL Dialect Differences

### Timestamps

| Database | Unix Timestamp | ISO Timestamp |
|----------|---------------|---------------|
| SQLite   | `strftime('%s', 'now')` | `datetime('now')` |
| MySQL    | `UNIX_TIMESTAMP()` | `NOW()` |
| PostgreSQL | `EXTRACT(EPOCH FROM NOW())` | `NOW()` |
| ODBC     | `UNIX_TIMESTAMP()` (generic) | `CURRENT_TIMESTAMP` (standard SQL) |

### Placeholders

| Database | Placeholder Style |
|----------|------------------|
| SQLite   | `?` |
| MySQL    | `?` |
| PostgreSQL | `$1, $2, $3, ...` |
| ODBC     | `?` (most drivers) |

### Boolean Types

| Database | Boolean Type | Storage |
|----------|-------------|---------|
| SQLite   | `INTEGER` | 0 or 1 |
| MySQL    | `TINYINT(1)` | 0 or 1 |
| PostgreSQL | `BOOLEAN` | true/false |
| ODBC     | `BOOLEAN` (standard SQL, may need conversion) | Database-dependent |

### JSON Types

| Database | JSON Type | Notes |
|----------|-----------|-------|
| SQLite   | `TEXT` | Stored as text |
| MySQL    | `JSON` | Native JSON type |
| PostgreSQL | `JSONB` | Binary JSON (faster) |
| ODBC     | `TEXT` | Generic TEXT for maximum compatibility |

## Migration Support

Migrations are database-agnostic and use standard SQL. However, some SQL functions may need to be abstracted:

```sql
-- Migration file: 1_initial_schema.sql
-- Uses dialect-aware functions via Go code, not raw SQL
```

The migration system automatically handles:
- Database-specific syntax differences
- Timestamp functions
- Auto-increment syntax
- Data type mappings

## Adding New Drivers

To add a new database driver:

1. **Implement the Driver interface**:
```go
type MyDriver struct {
    dialect *MyDialect
}

func (d *MyDriver) Type() DriverType {
    return DriverMyDB
}

func (d *MyDriver) Open(dsn string) (*sql.DB, error) {
    // Open connection
}

func (d *MyDriver) Configure(db *sql.DB) error {
    // Configure database
}

func (d *MyDriver) Dialect() Dialect {
    return d.dialect
}

func (d *MyDriver) Close() error {
    return nil
}
```

2. **Implement the Dialect interface**:
```go
type MyDialect struct{}

func (d *MyDialect) CurrentTimestamp() string {
    return "MY_TIMESTAMP_FUNCTION()"
}
// ... implement all Dialect methods
```

3. **Register the driver**:
```go
database.RegisterDriver(NewMyDriver())
```

## Dependencies

### Required (SQLite)
- `modernc.org/sqlite` - Pure Go SQLite driver

### Optional (MySQL)
- `github.com/go-sql-driver/mysql` - MySQL driver

### Optional (PostgreSQL)
- `github.com/lib/pq` - PostgreSQL driver

### Optional (ODBC)
- `github.com/alexbrainman/odbc` - ODBC driver (requires cgo)
- Requires ODBC driver manager:
  - Linux: `unixodbc` and `unixodbc-dev`
  - macOS: `unixodbc` (via Homebrew)
  - Windows: Built-in ODBC
- Database-specific ODBC drivers must be installed separately

## Backward Compatibility

The existing `Init(projectRoot)` function is maintained for backward compatibility. It defaults to SQLite and uses the same file path as before.

## Performance Considerations

- **SQLite**: Best for single-user or small deployments
- **MySQL**: Good for medium-scale deployments
- **PostgreSQL**: Best for large-scale or complex queries
- **Connection pooling**: All drivers use appropriate connection pool settings

## Testing

Each driver should be tested independently:

```go
func TestMySQLDriver(t *testing.T) {
    driver := NewMySQLDriver()
    db, err := driver.Open("test_dsn")
    // ... test driver
}
```

## Future Enhancements

- [x] ODBC driver implementation ✅
- [ ] Connection pooling configuration
- [ ] Read replica support
- [ ] Database-specific optimizations
- [ ] Migration rollback support
- [ ] Database health checks
