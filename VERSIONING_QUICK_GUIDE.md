# API Versioning - Quick Reference

## What Changed?

### Before
```
GET /accounts
POST /login
DELETE /accounts/{id}
```

### After
```
GET /api/v1/accounts
POST /api/v1/auth/login
DELETE /api/v1/accounts/{id}
```

## File Organization

```
routes/
├── routes.go          # Entry point - registers all versions
├── v1.go              # All v1 routes (current stable)
│   ├── Auth endpoints (/api/v1/auth/*)
│   └── Account endpoints (/api/v1/accounts/*)
├── auth_routes.go     # [Deprecated - can be removed]
└── account_routes.go  # [Deprecated - can be removed]
```

## Example: Updating Your Frontend

### Before
```javascript
// Old endpoint
fetch('http://localhost:8081/accounts')
  .then(res => res.json())
```

### After
```javascript
// New versioned endpoint
fetch('http://localhost:8081/api/v1/accounts')
  .then(res => res.json())
```

## All v1 Endpoints Reference

### Authentication Endpoints
```
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
```

### Account Endpoints
```
GET    /api/v1/accounts
GET    /api/v1/accounts/{id}
POST   /api/v1/accounts
DELETE /api/v1/accounts/{id}

POST   /api/v1/accounts/{id}/deposit
POST   /api/v1/accounts/{id}/withdraw
POST   /api/v1/accounts/{id}/transfers

GET    /api/v1/accounts/{id}/entries
GET    /api/v1/accounts/{id}/transactions/{transaction_id}
PUT    /api/v1/accounts/{id}/reconcile
```

## Why This Matters

### For Backend Development
- ✅ Can run multiple API versions simultaneously
- ✅ Easy to add v2, v3 later
- ✅ Clear separation of concerns
- ✅ Easier code maintenance

### For Frontend/Clients
- ✅ Clear which API version you're using
- ✅ No surprises when APIs change
- ✅ Can migrate at your own pace
- ✅ Better error messages with version info

### For Operations
- ✅ Easy to track usage per version
- ✅ Can deprecate versions gracefully
- ✅ CDN/Proxy friendly (different URLs = separate caches)
- ✅ Analytics show which versions are in use

## Future: Adding v2

When you need to make breaking changes:

1. Create `routes/v2.go`
2. Copy routes from v1 with your changes
3. Update `routes.go` to register both
4. Keep both running for 3-6 months
5. Deprecate v1 with clear timeline
6. Clients have time to migrate

## Testing Commands

```bash
# Test v1 login
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user","password":"pass"}'

# Test v1 accounts
curl -X GET http://localhost:8081/api/v1/accounts \
  -H "Authorization: Bearer <token>"

# Test v1 create account
curl -X POST http://localhost:8081/api/v1/accounts \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Savings","currency":"USD","is_system":false}'
```

## Summary

| Aspect | Benefit |
|--------|---------|
| **URL Path** | Clear, cacheable, explicit |
| **v1, v2, v3** | Easy numbering scheme |
| **Grouped files** | Better organization |
| **Single router** | Easy to manage all versions |
| **Future-proof** | Scale to 10+ versions if needed |

This approach is used by industry leaders: GitHub API, Stripe, AWS, Google Cloud.
