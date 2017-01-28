package aphtest

import (
	"net/http/httptest"

	"github.com/Jeffail/gabs"
)

// Reporter interface is used for reporting test failures
type Reporter interface {
	Error(...interface{})
	Errorf(string, ...interface{})
	Fatal(...interface{})
	Fatalf(string, ...interface{})
	Log(...interface{})
	Logf(string, ...interface{})
}

// ResponseBuilder interface is for incremental testing of http response and json object.
type ResponseBuilder interface {
	JSON() *gabs.Container
	Status(int) ResponseBuilder
	DumpJSON() string
}

// HTTPResponseBuilder implements ResponseBuilder interface
type HTTPResponseBuilder struct {
	reporter Reporter
	response *httptest.ResponseRecorder
	failed   bool
}

// NewHTTPResponseBuilder is the constructor for HTTPResponseBuilder
func NewHTTPResponseBuilder(rep Reporter, w *httptest.ResponseRecorder) ResponseBuilder {
	b := &HTTPResponseBuilder{
		reporter: rep,
		response: w,
	}
	return b
}

// Status matches the expected and actual http status
func (b *HTTPResponseBuilder) Status(status int) ResponseBuilder {
	if b.response.Code != status {
		b.failed = true
		b.reporter.Errorf(
			"actual http status %d did not match with expected status %d with error body %s\n",
			b.response.Code,
			status,
			string(IndentJSON(b.response.Body.Bytes())),
		)
	}
	return b
}

// JSON return a container type for introspecting json response
func (b *HTTPResponseBuilder) JSON() *gabs.Container {
	if b.failed {
		b.reporter.Fatal("errors stopped any further processing")
	}
	body := b.response.Body.Bytes()
	cont, err := gabs.ParseJSON(body)
	if err != nil {
		b.reporter.Fatalf("unable to parse json from response body %s\n", err)
	}
	return cont
}

// DumpJSON returns the JSON string from the response
func (b *HTTPResponseBuilder) DumpJSON() string {
	return string(IndentJSON(b.response.Body.Bytes()))
}
