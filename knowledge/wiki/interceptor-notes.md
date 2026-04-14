# Interceptor Notes

Current guidance for interceptor design in `nano`:

## General
- keep interceptor APIs small
- prefer option-based construction
- avoid hidden global behavior
- default behavior should be safe for production

## Retry
- retries should not be enabled implicitly for all methods
- explicit retry classifiers such as gRPC status codes are safer than broad defaults
- backoff should use `google.golang.org/grpc/backoff.Config` semantics instead of custom retry-specific backoff APIs
- exponential backoff is appropriate, but only when retry eligibility is explicit

## Circuit breaker
- breaker failures should represent dependency health, not every application error
- default trip codes should stay narrow

## Rate limiting and concurrency limiting
- `ratelimit` should provide at least one usable concrete limiter
- `concurrencylimit` should provide a concrete in-memory limiter such as a semaphore

## Tracing and metrics
- client stream tracing should reflect stream lifetime, not just stream creation
- client stream metrics should record terminal stream duration and terminal status, not setup latency under generic request-duration names
- prefer local test doubles over pulling tracing SDK dependencies into the root module when API-level tests are sufficient

## File layout
- interceptor wiring and interfaces stay in the package entry file
- concrete implementations live in separate files in the same package
