module signer.local

go 1.13

replace signer.local/pkg/signer => ./pkg/signer

require (
	github.com/xxixuuu/utils v0.0.0-20221118105627-dc446834c9e5 // indirect
	github.com/xxixuuu/netcp v0.0.0-20210920072459-508f436d452b // indirect
	github.com/herumi/bls-go-binary v1.0.0
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b // indirect
	signer.local/pkg/signer v0.0.0-00010101000000-000000000000
)
