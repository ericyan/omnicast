CAST_CHANNEL = internal/castv2/cast_channel

.PHONY: build
build: $(CAST_CHANNEL)/cast_channel.pb.go

.PHONY: clean
clean:
	rm -f $(CAST_CHANNEL)/*.pb.go

$(CAST_CHANNEL)/cast_channel.pb.go:
	protoc --gogo_out=. $(CAST_CHANNEL)/cast_channel.proto
