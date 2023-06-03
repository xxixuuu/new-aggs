module dummy.local

go 1.13

replace dummy.local/pkg/dummy => ./pkg/dummy

require (
	dummy.local/pkg/dummy v0.0.0-00010101000000-000000000000
	github.com/xxixuuu/netcp v0.0.0-20210920072459-508f436d452b
	github.com/xxixuuu/utils v0.0.0-20221118105627-dc446834c9e5
)
