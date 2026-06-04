# API Versioning Strategy

## Overview
This project uses **URL Path Versioning** as the primary strategy for API versioning. All endpoints are prefixed with `/api/v{version}/`.

## Current Structure
- **v1** - Current stable version
- **v2, v3, etc.** - Future versions (when needed)

## Why URL Path Versioning?

### ✅ Advantages
1. **Explicit & Clear**: Version is visible in every request URL
2. **Cacheable**: Each version has unique URLs, allowing HTTP caching
3. **Proxy-Friendly**: Works seamlessly with CDNs and reverse proxies
4. **CORS Compatible**: No header manipulation needed
5. **Easy Deprecation**: Can set sunset dates and notify clients
6. **Documentation**: Version is clear in API docs and browser URLs
7. **Analytics**: Easy to track usage per API version
8. **Testing**: Simple to test multiple versions simultaneously

### ❌ Alternatives (and why we didn't use them)
- **Query Parameter** (`?version=1`): Makes URLs longer, harder to cache
- **Header-based** (`Accept: application/vnd.api.v1+json`): Hidden from URLs, harder to debug
- **Subdomain** (`v1.api.example.com`): More infrastructure overhead

## File Structure

```
routes/
├── routes.go          # Main router orchestrator
├── v1.go              # All v1 endpoints
├── auth_routes.go     # (Legacy - kept for reference)
└── account_routes.go  # (Legacy - kept for reference)
```

## Adding New API Versions

### When to Create a New Version
- Breaking changes to existing endpoints
- Removing deprecated endpoints
- Major refactoring of response format
- Major feature additions

### How to Add a New Version

1. **Create new version file** (`routes/v2.go`):
```go
package routes

func RegisterV2Routes(srv *handlers.Server, authSrv *handlers.AuthServer) {
    // All v2 endpoints with /api/v2/ prefix
    http.HandleFunc("GET /api/v2/accounts", ...)
}
```

2. **Register in routes.go**:
```go
func RegisterAllRoutes(srv *handlers.Server, authSrv *handlers.AuthServer) {
    RegisterV1Routes(srv, authSrv)
    RegisterV2Routes(srv, authSrv)  // Add this
}
```

3. **Support both versions** during transition period
4. **Set deprecation date** for old version
5. **Update documentation** with migration guide

## Current API Endpoints (v1)

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login user
- `POST /api/v1/auth/refresh` - Refresh access token

### Accounts
- `GET /api/v1/accounts` - List all accounts
- `GET /api/v1/accounts/{id}` - Get account details
- `POST /api/v1/accounts` - Create account
- `DELETE /api/v1/accounts/{id}` - Delete account
- `POST /api/v1/accounts/{id}/deposit` - Deposit money
- `POST /api/v1/accounts/{id}/withdraw` - Withdraw money
- `POST /api/v1/accounts/{id}/transfers` - Transfer between accounts
- `GET /api/v1/accounts/{id}/entries` - Get account entries
- `GET /api/v1/accounts/{id}/transactions/{transaction_id}` - Get transaction
- `PUT /api/v1/accounts/{id}/reconcile` - Reconcile account

## Deprecation Strategy

### Deprecation Timeline
1. **Release v2** with backward compatibility to v1
2. **Announce deprecation** (3-6 months notice minimum)
3. **Provide migration guide** with examples
4. **Monitor usage** - v1 endpoint hit rates
5. **Sunset v1** on announced date
6. **Return 410 Gone** for v1 endpoints after sunset

### Example Response Header (Deprecation Warning)
```
Deprecation: true
Sunset: Sun, 01 Jan 2025 00:00:00 GMT
Link: <https://docs.api.com/v2>; rel="successor-version"
```

## Response Versioning

All response formats are tied to the API version. When making breaking changes:

```go
// v1 response format
GET /api/v1/accounts/{id}
{
    "id": "uuid",
    "name": "Savings Account",
    "balance": "1000.00"
}

// v2 response format (hypothetical)
GET /api/v2/accounts/{id}
{
    "id": "uuid",
    "name": "Savings Account",
    "balance": {
        "amount": "1000.00",
        "currency": "USD"
    },
    "metadata": {
        "created_at": "2024-01-01T00:00:00Z"
    }
}
```

## Best Practices Applied

✅ **Version in URL path** - Explicit and cacheable
✅ **Semantic versioning** - v1, v2, v3 (simple and clear)
✅ **Grouped by version file** - Easy maintenance
✅ **Single entry point** - `RegisterAllRoutes()` handles all versions
✅ **Backward compatibility** - Old versions work during transition
✅ **Clear documentation** - Each version clearly separated
✅ **Future-proof** - Easy to add v2, v3 without refactoring

## Testing API Versions

```bash
# Test v1
curl -X GET http://localhost:8081/api/v1/accounts \
  -H "Authorization: Bearer <token>"

# Future: Test v2
curl -X GET http://localhost:8081/api/v2/accounts \
  -H "Authorization: Bearer <token>"
```

## Migration Checklist for New Version

- [ ] Create new version file (e.g., `routes/v2.go`)
- [ ] Copy routes from v1 with updates/changes
- [ ] Register new version in `routes.go`
- [ ] Update response models if needed
- [ ] Create migration guide in documentation
- [ ] Test both versions work simultaneously
- [ ] Add deprecation headers to old version
- [ ] Announce sunset date to users
- [ ] Monitor usage statistics
- [ ] Decommission old version after sunset date
