package aphtest

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/dictyBase/go-middlewares/middlewares/query"
	"github.com/dictyBase/go-middlewares/middlewares/router"
	"github.com/julienschmidt/httprouter"
)

// RequestBuilder interface is for incremental building of RequestBuilder to receive
// a ResponseBuilder object
type RequestBuilder interface {
	AddRouterParam(string, string) RequestBuilder
	AddIncludes(...string) RequestBuilder
	AddFieldSets(string, bool, ...string) RequestBuilder
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

// AddIncludes adds relationships includes of JSONAPI  in the http request context
func (b *HTTPRequestBuilder) AddIncludes(relationships ...string) RequestBuilder {
	p, ok := b.req.Context().Value(query.ContextKeyQueryParams).(*query.Params)
	if ok {
		if p.HasIncludes {
			p.Includes = append(p.Includes, relationships...)
		} else {
			p.Includes = relationships
			p.HasIncludes = true
		}
	} else {
		p = &query.Params{
			HasIncludes: true,
			Includes:    relationships,
		}
	}
	ctx := context.WithValue(b.req.Context(), query.ContextKeyQueryParams, p)
	b.req = b.req.WithContext(ctx)
	return b
}

// AddFieldSets adds sparse fieldsets of JSONAPI in the http request context
func (b *HTTPRequestBuilder) AddFieldSets(resource string, relationship bool, fields ...string) RequestBuilder {
	p, ok := b.req.Context().Value(query.ContextKeyQueryParams).(*query.Params)
	f := &query.Fields{Relationship: relationship}
	f.Append(fields...)
	if ok {
		if p.HasSparseFields {
			p.SparseFields[resource] = f
		} else {
			p.HasSparseFields = true
			p.SparseFields = map[string]*query.Fields{
				resource: f,
			}
		}
	} else {
		p = &query.Params{
			HasSparseFields: true,
			SparseFields: map[string]*query.Fields{
				resource: f,
			},
		}
	}
	ctx := context.WithValue(b.req.Context(), query.ContextKeyQueryParams, p)
	b.req = b.req.WithContext(ctx)
	return b
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
		ctx := context.WithValue(b.req.Context(), router.ContextKeyParams, b.params)
		req = b.req.WithContext(ctx)
	}
	w := httptest.NewRecorder()
	b.handlerFn(w, req)
	return NewHTTPResponseBuilder(b.reporter, w)
}
