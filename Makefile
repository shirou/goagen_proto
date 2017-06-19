gen_proto:
	cd example && goagen gen --pkg-path=github.com/shirou/goagen_proto -d github.com/shirou/goagen_proto/example/design
	cd example && protoc --go_out=plugins=grpc:. *.proto
	cd example/server && go build
	cd example/client && go build
