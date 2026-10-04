module backend

go 1.27.1

require (
	github.com/Masterminds/squirrel v1.5.4 // sql builder
	github.com/impl0x/go-utils v0.8.2 // used for /cache.TTLCache short lived caches.
	github.com/impl0x/mo v1.6.0 // web api framework
	github.com/jackc/pgx/v5 v5.10.0 // postgresql driver
	github.com/mileusna/useragent v1.3.5 // user agent parsing
	golang.org/x/crypto v0.55.0 // password hashing
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
)
