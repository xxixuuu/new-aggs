module signer.local

go 1.13

replace signer.local/pkg/signer => ./pkg/signer

require (
	github.com/Coresummer/netcp v0.0.0-20210920072459-508f436d452b // indirect
	github.com/Coresummer/utils v0.0.0-20210920072513-3cb7edbc7a3a // indirect
	github.com/herumi/bls-go-binary v1.0.0
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b // indirect
	signer.local/pkg/signer v0.0.0-00010101000000-000000000000
)
