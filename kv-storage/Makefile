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

	@./kvstorage -id=1 -http_port=8080 -raft_port=5051 localhost:5052 localhost:5053 &
	@./kvstorage -id=2 -http_port=8081 -raft_port=5052 localhost:5051 localhost:5053 &
	@./kvstorage -id=3 -http_port=8082 -raft_port=5053 localhost:5051 localhost:5052 &
	
	@rm kvstorage