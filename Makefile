PROTO_DIR=proto
PB_OUT=pkg/pb

ifdef PROTO_SUBDIR
	PROTO_FILES=$(wildcard $(PROTO_DIR)/$(PROTO_SUBDIR)/*.proto)
else
	PROTO_FILES=$(wildcard $(PROTO_DIR)/*/*.proto)
endif

protoc:
	protoc -I $(PROTO_DIR) \
	  --go_out=$(PB_OUT) --go_opt=paths=source_relative \
	  --go-grpc_out=$(PB_OUT) --go-grpc_opt=paths=source_relative \
	  $(PROTO_FILES)

CREATE=migrate create -ext sql -dir $(path) -seq $(name)
create-migration:
	$(CREATE)
