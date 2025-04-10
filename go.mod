module github.com/huyshop/user

go 1.23.7

replace github.com/huyshop/header v0.0.23 => ../header

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-redis/redis/v8 v8.11.5
	github.com/go-sql-driver/mysql v1.9.0
	github.com/joho/godotenv v1.5.1
	github.com/rs/xid v1.6.0
	github.com/urfave/cli/v2 v2.27.6
	golang.org/x/crypto v0.36.0
	xorm.io/xorm v1.3.9
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.6 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/syndtr/goleveldb v1.0.0 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250227231956-55c901821b1e // indirect
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/huyshop/header v0.0.25
	golang.org/x/net v0.37.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/grpc v1.71.0
	google.golang.org/protobuf v1.36.5 // indirect
	xorm.io/builder v0.3.13 // indirect
)
