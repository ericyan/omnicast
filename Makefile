CASTV2 = gcast/internal/castv2

omnicastd:
	go build -o omnicastd cmd/omnicastd/main.go

$(CASTV2)/cast_channel/cast_channel.pb.go:
	$(MAKE) -C $(CASTV2)/cast_channel
