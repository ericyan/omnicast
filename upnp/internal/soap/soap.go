package soap

import (
	"encoding/xml"
	"io"
	"net/http"
	"strconv"
	"strings"
	"text/template"
)

type Action struct {
	Namespace string
	Name      string
}

// An Error represents an UPnP DCP specific error.
type Error struct {
	Code        int
	Description string
}

// Error implements the error interface
func (err *Error) Error() string {
	return err.Description
}

// UPnP defined error codes
var (
	ErrInvalidAction        = &Error{401, "Invalid Action"}
	ErrInvalidArgs          = &Error{402, "Invalid Args"}
	ErrActionFailed         = &Error{501, "Action Failed"}
	ErrArgValueInvalid      = &Error{600, "Argument Value Invalid"}
	ErrArgValueOutOfRange   = &Error{601, "Argument Value Out of Range"}
	ErrActionNotImplemented = &Error{602, "Optional Action Not Implemented"}
	ErrOutOfMemory          = &Error{603, "Out of Memory"}
	ErrInterventionRequired = &Error{604, "Human Intervention Required"}
	ErrArgTooLong           = &Error{605, "String Argument Too Long"}
)

type Request struct {
	Action *Action
	Args   map[string]string
}

func ParseHTTPRequest(r *http.Request) (*Request, error) {
	action, err := parseAction(r.Header.Get("SOAPAction"))
	if err != nil {
		return nil, err
	}

	args, err := parseArgs(r.Body, action)
	if err != nil {
		return nil, err
	}

	return &Request{action, args}, nil
}

func parseAction(s string) (*Action, error) {
	action, err := strconv.Unquote(s)
	if err != nil {
		return nil, err
	}

	a := strings.SplitN(action, "#", 2)

	return &Action{a[0], a[1]}, nil
}

func parseArgs(r io.Reader, action *Action) (map[string]string, error) {
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

type Response struct {
	Action *Action
	Args   map[string]string
	Error  *Error
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
					<errorCode>{{.Code}}</errorCode>
					<errorDescription>{{.Description}}</errorDescription>
				</UPnPError>
			</detail>
		</s:Fault>
		{{- end }}
  </s:Body>
</s:Envelope>
`

func (resp *Response) WriteTo(w io.Writer) error {
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
