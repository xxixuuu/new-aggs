module verifier.local

go 1.13

replace verifier.local/pkg/kvs => ./pkg/kvs

replace verifier.local/pkg/tracer => ./pkg/tracer

replace verifier.local/pkg/verifier => ./pkg/verifier

replace verifier.local/pkg/infra => ./pkg/infra

require (
	github.com/xxixuuu/utils v0.0.0-20221118105627-dc446834c9e5 // indirect
	github.com/xxixuuu/netcp v0.0.0-20210920072459-508f436d452b // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/go-redis/redis/v8 v8.11.3 // indirect
	github.com/herumi/bls-go-binary v1.0.0
	github.com/mattn/go-runewidth v0.0.13 // indirect
	go.opentelemetry.io/otel v1.0.0 // indirect
	verifier.local/pkg/infra v0.0.0-00010101000000-000000000000
	verifier.local/pkg/kvs v0.0.0-00010101000000-000000000000
	verifier.local/pkg/tracer v0.0.0-00010101000000-000000000000 // indirect
	verifier.local/pkg/verifier v0.0.0-00010101000000-000000000000
)
