# Error Handling in Verdis

A quick guide to creating, handling, and logging errors in Verdis.

## Quick Start

```go
import verr "github.com/ElshadHu/verdis/internal/errors"

// Create a simple error
err := verr.New(verr.ERR, "something went wrong")

// Create with formatting
err := verr.Newf(verr.ERR, "invalid value: %s", value)

// Wrap an existing error
err := verr.Wrapf(cause, verr.ERR, "failed to do X")

// Add context for logging
err := verr.ErrKeyNotFound.With("key", keyName)
```

## Error Codes

| Code | When to Use |
|------|-------------|
| `ERR` | Generic errors, unknown commands, invalid arguments |
| `WRONGTYPE` | Operation on wrong data type |
| `SYNTAX` | Protocol/parsing syntax errors |
| `NOTFOUND` | Key doesn't exist (internal use) |
| `DELETED` | Key was deleted at requested version |
| `VERSIONNOTFOUND` | Requested version doesn't exist |
| `BUSY` | Server temporarily busy (transient) |
| `TIMEOUT` | Operation timed out (transient) |
| `INTERNAL` | Internal server error |

## Transient vs Permanent Errors

**Transient errors** are temporary - retry might succeed:
- Connection timeouts
- Server busy
- Address in use (during startup)

**Permanent errors** won't succeed on retry:
- Wrong number of arguments
- Unknown command
- Key not found
- Syntax errors

### Checking Transient Errors

```go
if verr.IsTransient(err) {
    // Safe to retry
    return retry(operation)
}
// Don't retry - it will fail again
return err
```

### Creating Transient Errors

```go
// Automatically transient based on code
err := verr.New(verr.TIMEOUT, "operation timed out")  // transient=true
err := verr.New(verr.ERR, "bad argument")             // transient=false

// Explicitly transient
err := verr.NewTransient(verr.ERR, "temporary failure")
```

## Common Constructors

### Command Errors

```go
// Unknown command
verr.ErrUnknownCommand("BADCMD")

// Wrong argument count
verr.ErrWrongArity("GET", got, min, max)

// Invalid argument
verr.ErrInvalidArgument("timeout", "must be positive")

// Invalid version number
verr.ErrInvalidVersion("abc")
```

### Storage Errors

```go
// Sentinel errors (can use directly)
verr.ErrKeyNotFound
verr.ErrKeyDeleted
verr.ErrVersionNotFound

// With context
verr.KeyNotFound("mykey")
verr.KeyDeleted("mykey", version)
verr.VersionNotFound("mykey", version)
```

### Server Errors

```go
verr.ErrAddressInUse("localhost:6379", cause)
verr.ErrServerBusy("too many connections")
verr.ErrMaxConnections(current, max)
verr.ErrInternal("unexpected state")
```

### Protocol Errors

```go
verr.ErrSyntax("unexpected token")
verr.ErrProtocol("invalid RESP format")
verr.ErrParseFailed("command", cause)
verr.ErrEmptyCommand()
```

## Wrapping Errors

Preserve the cause chain while adding context:

```go
// Wrap with new message
err := verr.Wrapf(cause, verr.ERR, "failed to process %s", key)

// Add context to existing error
err := verr.ErrKeyNotFound.With("key", key).With("operation", "GET")

// Wrap an existing VerdisError
err := existingErr.Wrap(underlyingCause)
```

## Checking Error Types

```go
// Check specific conditions
if verr.IsNotFound(err) { ... }
if verr.IsDeleted(err) { ... }
if verr.IsTimeout(err) { ... }
if verr.IsBusy(err) { ... }

// Get the error code
code := verr.GetCode(err)

// Standard errors.Is/As
if verr.Is(err, verr.ErrKeyNotFound) { ... }

var ve *verr.VerdisError
if verr.As(err, &ve) {
    // Use ve.Code, ve.Message, etc.
}
```

## Logging with slog

Errors implement `slog.LogValuer` for structured logging:

```go
err := verr.ErrWrongArity("GET", 0, 1, 1)
slog.Error("command failed", "error", err)
// Output: level=ERROR msg="command failed" error.code=ERR error.message="wrong number of arguments..." error.transient=false error.command=GET error.got=0 error.min=1 error.max=1
```

### Log Level by Transient Status

```go
if verr.IsTransient(err) {
    slog.Warn("transient error, will retry", "error", err)
} else {
    slog.Error("permanent error", "error", err)
}
```

## Converting to RESP

When returning errors to Redis clients:

```go
import "github.com/ElshadHu/verdis/internal/protocol"

// VerdisError -> RESP Error
ve := verr.ErrUnknownCommand("FOO")
return protocol.NewError(ve.Error())
// Returns: -ERR unknown command 'FOO'\r\n
```

## Error Hierarchy

```
VerdisError (base)
├── Command Errors (non-transient)
│   ├── ErrUnknownCommand
│   ├── ErrWrongArity
│   ├── ErrWrongType
│   └── ErrInvalidArgument
├── Storage Errors (non-transient)
│   ├── ErrKeyNotFound
│   ├── ErrKeyDeleted
│   └── ErrVersionNotFound
├── Protocol Errors (mixed)
│   ├── ErrSyntax (non-transient)
│   ├── ErrProtocol (non-transient)
│   └── ErrReadTimeout (transient)
└── Server Errors (mixed)
    ├── ErrAddressInUse (transient)
    ├── ErrServerBusy (transient)
    └── ErrInternal (non-transient)
```

