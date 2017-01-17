package aphtest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gocraft/dbr"
)

// Resource is the interface that every http handler have to implement
type Resource interface {
	// Gets the database handler
	GetDbh() *dbr.Connection
	// Handles the http GET for singular resource
	Get(http.ResponseWriter, *http.Request)
	// Handles the http GET for collection resource
	GetAll(http.ResponseWriter, *http.Request)
	// Handles the http POST
	Create(http.ResponseWriter, *http.Request)
	// Handles the http PATCH
	Update(http.ResponseWriter, *http.Request)
	// Handles the http DELETE
	Delete(http.ResponseWriter, *http.Request)
}

// ExpectBuilder interface is for incremental building of http configuration
type ExpectBuilder interface {
	Get(string) RequestBuilder
}

// HTTPExpectBuilder implements ExpectBuilder interface
type HTTPExpectBuilder struct {
	reporter Reporter
	host     string
	resource Resource
}

// NewHTTPExpectBuilder is the constructor for HTTPExpectBuilder
func NewHTTPExpectBuilder(rep Reporter, host string, rs Resource) ExpectBuilder {
	return &HTTPExpectBuilder{
		reporter: rep,
		host:     host,
		resource: rs,
	}
}

// Get configures Request to execute a http GET request
func (b *HTTPExpectBuilder) Get(path string) RequestBuilder {
	req := httptest.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/%s",
			b.host,
			strings.Trim(path, "/"),
		),
		nil,
	)
	return NewHTTPRequestBuilder(b.reporter, req, b.resource.Get)
}
