package types

import "testing"

const metadataTestCase = `<?xml version="1.0" encoding="UTF-8"?>
<DIDL-Lite xmlns="urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:upnp="urn:schemas-upnp-org:metadata-1-0/upnp/">
  <item id="1" parentID="-1" restricted="1">
    <dc:title>WALL-E</dc:title>
    <upnp:class>object.item.videoItem.movie</upnp:class>
    <upnp:genre>Unknown</upnp:genre>
    <upnp:storageMedium>UNKNOWN</upnp:storageMedium>
    <upnp:writeStatus>UNKNOWN</upnp:writeStatus>
  </item>
</DIDL-Lite>`

func TestMetadata(t *testing.T) {
	m := make(Metadata)
	if err := m.UnmarshalText([]byte(metadataTestCase)); err != nil {
		t.Error(err)
	}

	if m.Title() != "WALL-E" {
		t.Errorf("Unexpected title: '%s'", m.Title())
	}
}
