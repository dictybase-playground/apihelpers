package aphtest

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/dictyBase/go-middlewares/middlewares/router"
	"github.com/julienschmidt/httprouter"
)

// RequestBuilder interface is for incremental building of RequestBuilder to receive
// a ResponseBuilder object
type RequestBuilder interface {
	AddRouterParam(string, string) RequestBuilder
	Expect() ResponseBuilder
}

// HTTPRequestBuilder implements RequestBuilder interface
type HTTPRequestBuilder struct {
	params    httprouter.Params
	handlerFn http.HandlerFunc
	reporter  Reporter
	req       *http.Request
}

// NewHTTPRequestBuilder is the constructor for HTTPRequestBuilder
func NewHTTPRequestBuilder(rep Reporter, req *http.Request, fn http.HandlerFunc) RequestBuilder {
	return &HTTPRequestBuilder{
		handlerFn: fn,
		reporter:  rep,
		req:       req,
	}
}

// AddRouterParam add key and value to httprouter's parameters
func (b *HTTPRequestBuilder) AddRouterParam(key, value string) RequestBuilder {
	if len(b.params) > 0 {
		b.params = append(b.params, httprouter.Param{Key: key, Value: value})
	} else {
		var p httprouter.Params
		p = append(p, httprouter.Param{Key: key, Value: value})
		b.params = p
	}
	return b
}

// Expect gets the Response object for further testing
func (b *HTTPRequestBuilder) Expect() ResponseBuilder {
	req := b.req
	if len(b.params) > 0 {
		ctx := context.WithValue(context.Background(), router.ContextKeyParams, b.params)
		req = b.req.WithContext(ctx)
	}
	w := httptest.NewRecorder()
	b.handlerFn(w, req)
	return NewHTTPResponseBuilder(b.reporter, w)
}
