.PHONY: generate proto oapi
generate: proto oapi

oapi:
	@oapi-codegen -package=oapi endpoints.yaml > server.gen.go   

proto:
	@protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	api/proto/raft.proto
    

.PHONY: run
run:
	@go build

	@./kvstorage -id=1 -http_port=8080 -raft_port=6051 localhost:6052 localhost:6053 &
	@./kvstorage -id=2 -http_port=8081 -raft_port=6052 localhost:6051 localhost:6053 &
	@./kvstorage -id=3 -http_port=8082 -raft_port=6053 localhost:6051 localhost:6052 &
	
	@rm kvstorage