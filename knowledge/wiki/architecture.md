# Architecture

## Package boundaries

The repository treats top-level packages as independent modules. Code under `grpc/...` should not import other top-level packages such as `metric`, `log`, `status`, `config`, or `validator` unless an explicit exception is documented.

When behavior from another top-level package is needed:
- define a small local interface
- inject the real implementation at runtime

Current notable example:
- reusable gRPC metrics interceptors live in `metric/grpc` instead of `grpc/...` so `grpc` does not depend on the top-level `metric` package

## Design direction

Preferred style:
- native gRPC APIs
- small option-based helpers
- explicit interceptor construction
- safe defaults

Avoid:
- broad framework abstractions
- hidden globals
- large generic middleware surfaces copied wholesale from framework ecosystems

## Examples and modules

- `examples/helloworld/` is the main end-to-end reference flow
- plugin and example directories are separate modules and need local validation when changed
