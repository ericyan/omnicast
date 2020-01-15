CASTV2 = gcast/internal/castv2

bin/omnicastd:
	go build -o bin/omnicastd cmd/omnicastd/main.go

upnp/scpd/statik.go:
	statik -src upnp/scpd/ -dest upnp/ -p scpd -c "Package scpd contains service definitions."

$(CASTV2)/cast_channel/cast_channel.pb.go:
	$(MAKE) -C $(CASTV2)/cast_channel
