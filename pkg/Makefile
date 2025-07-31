PROTO_DIR=api/proto
GEN_DIR=pkg/fin-aggregator

.PHONY: proto-gen proto-clean

proto-gen:
	mkdir -p $(GEN_DIR)
	buf generate $(PROTO_DIR) --output $(GEN_DIR)

proto-clean:
	rm -rf $(GEN_DIR)