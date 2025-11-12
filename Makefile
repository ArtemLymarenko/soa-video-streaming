PROTO_DIR=proto
PB_OUT=pkg/pb

protoc:
	protoc -I $(PROTO_DIR) \
	  --go_out=$(PB_OUT) --go_opt=paths=source_relative \
	  --go-grpc_out=$(PB_OUT) --go-grpc_opt=paths=source_relative \
	  $(shell find $(PROTO_DIR)/$(PROTO_SUBDIR) -name "*.proto")