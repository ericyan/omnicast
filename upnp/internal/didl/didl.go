package didl

import (
	"encoding/xml"
)

// String represents a string value
type String struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

// Type returns value type, as the local part of the element name.
func (s *String) Type() string {
	return s.XMLName.Local
}

// String returns the value as a string.
func (s *String) String() string {
	return s.Value
}

// Item represents an item element.
type Item struct {
	XMLName    xml.Name
	ID         string    `xml:"id,attr"`
	ParentID   string    `xml:"parentID,attr"`
	Restricted bool      `xml:"restricted,attr"`
	Values     []*String `xml:",any"`
}

// Document represents a DIDL-Lite document.
type Document struct {
	Items []Item `xml:"item"`
}
