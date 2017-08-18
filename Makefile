build:
	go build --ldflags '-linkmode external -extldflags "-static"' -o gopwsafe ./gopwsafe.go
