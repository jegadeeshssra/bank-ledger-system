- swagger docs
- should create indexes for accountids   
- async DB operation
- Prevent double login for the same account
- refresh token
- distributed redis and ledger using raft
- write tests to simulate 5000 users    



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