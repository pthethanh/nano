module github.com/pthethanh/nano/plugins/cache/redis

go 1.24.5

require (
	github.com/alicebob/miniredis/v2 v2.37.0
	github.com/pthethanh/nano v0.0.2
	github.com/redis/go-redis/v9 v9.18.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
)

replace github.com/pthethanh/nano v0.0.2 => ../../..
