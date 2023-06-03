module aggregator.local

go 1.13

replace aggregator.local/pkg/tracer => ./pkg/tracer

replace aggregator.local/pkg/aggregator => ./pkg/aggregator

replace aggregator.local/pkg/infra => ./pkg/infra

//import original repo then replace it to the fork @ here
// replace github.com/herumi/bls-go-binary => github.com/Coresummer/bls-go-binary v1.0.1-0.20201111053131-cfa39c0d3aab

require (
	aggregator.local/pkg/aggregator v0.0.0-00010101000000-000000000000
	aggregator.local/pkg/infra v0.0.0-00010101000000-000000000000
	aggregator.local/pkg/tracer v0.0.0-00010101000000-000000000000
	github.com/xxixuuu/utils v0.0.0-20221118105627-dc446834c9e5 // indirect
	github.com/xxixuuu/netcp v0.0.0-20210920072459-508f436d452b // indirect
	github.com/herumi/bls-go-binary v1.0.0
	github.com/mattn/go-runewidth v0.0.13 // indirect

)
