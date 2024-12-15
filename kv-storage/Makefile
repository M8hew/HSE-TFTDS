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
	@./kvstorage -id=1 -http_port=8080 -raft_port=9091 localhost:9092 localhost:9093
	@./kvstorage -id=2 -http_port=8081 -raft_port=9092 localhost:9091 localhost:9093
	@./kvstorage -id=3 -http_port=8082 -raft_port=9093 localhost:9091 localhost:9092
	@rm kvstorage