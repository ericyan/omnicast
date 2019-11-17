package upnp

import (
	"encoding/xml"
	"io"
	"net/http"
	"strconv"
	"strings"
	"text/template"
)

type SOAPAction struct {
	Namespace string
	Name      string
}

// An SOAPError represents an UPnP DCP specific error.
type SOAPError struct {
	Code        int
	Description string
}

// Error implements the error interface
func (err *SOAPError) Error() string {
	return err.Description
}

// UPnP defined error codes
var (
	ErrInvalidAction        = &SOAPError{401, "Invalid Action"}
	ErrInvalidArgs          = &SOAPError{402, "Invalid Args"}
	ErrActionFailed         = &SOAPError{501, "Action Failed"}
	ErrArgValueInvalid      = &SOAPError{600, "Argument Value Invalid"}
	ErrArgValueOutOfRange   = &SOAPError{601, "Argument Value Out of Range"}
	ErrActionNotImplemented = &SOAPError{602, "Optional Action Not Implemented"}
	ErrOutOfMemory          = &SOAPError{603, "Out of Memory"}
	ErrInterventionRequired = &SOAPError{604, "Human Intervention Required"}
	ErrArgTooLong           = &SOAPError{605, "String Argument Too Long"}
)

type SOAPRequest struct {
	Action *SOAPAction
	Args   map[string]string
}

func ParseHTTPRequest(r *http.Request) (*SOAPRequest, error) {
	action, err := parseAction(r.Header.Get("SOAPAction"))
	if err != nil {
		return nil, err
	}

	args, err := parseArgs(r.Body, action)
	if err != nil {
		return nil, err
	}

	return &SOAPRequest{action, args}, nil
}

func parseAction(s string) (*SOAPAction, error) {
	action, err := strconv.Unquote(s)
	if err != nil {
		return nil, err
	}

	a := strings.SplitN(action, "#", 2)

	return &SOAPAction{a[0], a[1]}, nil
}

func parseArgs(r io.Reader, action *SOAPAction) (map[string]string, error) {
	actionName := xml.Name{action.Namespace, action.Name}
	args := make(map[string]string)

	d := xml.NewDecoder(r)
	var v string
	for {
		token, err := d.Token()
		if token == nil {
			break
		}
		if err != nil && err != io.EOF {
			return nil, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name != actionName {
				break
			}
		case xml.CharData:
			v = string(t)
		case xml.EndElement:
			if t.Name == actionName {
				return args, nil
			}

			args[t.Name.Local] = v
		}
	}

	return nil, io.EOF
}

type SOAPResponse struct {
	Action *SOAPAction
	Args   map[string]string
	Error  *SOAPError
}

const responseTemplate = `<?xml version="1.0" encoding="utf-8"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
  <s:Body>
    <u:{{.Action.Name}}Response xmlns:u="{{.Action.Namespace}}">
    {{- range $k, $v := .Args }}
      <{{$k}}>{{escape $v}}</{{$k}}>
    {{- end}}
    </u:{{.Action.Name}}Response>
		{{- with .Error }}
		<s:Fault>
			<faultcode>s:Client</faultcode>
			<faultstring>UPnPError</faultstring>
			<detail>
				<UPnPError xmlns="urn:schemas-upnp-org:control-1-0">
					<errorCode>{{.Error.Code}}</errorCode>
					<errorDescription>{{.Error.Description}}</errorDescription>
				</UPnPError>
			</detail>
		</s:Fault>
		{{- end }}
  </s:Body>
</s:Envelope>
`

func (resp *SOAPResponse) WriteTo(w io.Writer) error {
	funcs := template.FuncMap{
		"escape": func(s string) string {
			b := new(strings.Builder)
			err := xml.EscapeText(b, []byte(s))
			if err != nil {
				panic(err)
			}

			return b.String()
		}}

	tpl := template.Must(template.New("response").Funcs(funcs).Parse(responseTemplate))

	return tpl.Execute(w, resp)
}
