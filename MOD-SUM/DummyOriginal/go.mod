module dummy.local

go 1.15

replace dummy.local/pkg/dummy => ./pkg/dummy

require (
	dummy.local/pkg/dummy v0.0.0-00010101000000-000000000000
	github.com/Coresummer/netcp v0.0.0-20210920072459-508f436d452b
	github.com/Coresummer/utils v0.0.0-20210920072513-3cb7edbc7a3a
)
