start:
	go build 
	./user start

cdb:
	go build 
	./user createDb
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o user .