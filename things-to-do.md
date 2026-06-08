- swagger docs
- should create indexes for accountids   
- async DB operation
- Prevent double login for the same account
- secret key rotation
- distributed redis and ledger using raft
- write tests to simulate 5000 users    
- will have to check whether the response time of redis requests in less than 1ms
- pagination
- add HTTPS support


Things-we-done
--------------
- Double Entry Accounting
    - 2 separate entries for single from-n-to transaction
    - reconcile whenever needed
- async api behaviour in default
- Jwt authentication for every request
- Atomic operation for implementing 2 entries in DB and Redis using lua scripting for preventing race conditions
- Token Bucket Algo for rate limiting all endpoints
- Request and Response Validation
- Idempotency behaviour for /withdraw , /deposit , /transfer

OWASP
- Broken Object Level Authorization 
    - /accounts/{my-id} -> /accounts/{other-id} cannot happen bcuz we only use the account-id if its from the user-id from the jwt using sql query
- Broken Authentication
    - Weak JWT secret
    - Must validate : signature , exp , issuer
    - always enforce jwt.SigningMethodHS256
- Unrestricted Resource Consumption
    - Rate Limiting


CHECK-LIST
----------
AUTHORIZATION

□ object ownership checks
□ admin role checks

AUTHENTICATION

□ JWT expiry
□ signature validation
□ secret rotation

TRANSACTIONAL SAFETY

□ DB transactions
□ FOR UPDATE
□ idempotency

RATE LIMITING

□ Redis token bucket
□ Lua atomic updates

INPUT VALIDATION

□ UUID validation
□ amount validation
□ JSON unknown fields

INJECTION

□ parameterized SQL everywhere

CONFIGURATION

□ HTTPS
□ secure Redis
□ secure env vars
