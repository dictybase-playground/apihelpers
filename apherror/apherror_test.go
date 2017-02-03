package apherror

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSONAPIError(t *testing.T) {
	w := httptest.NewRecorder()
	sferr := ErrSparseFieldSets.New("sparse field errors")
	JSONAPIError(w, sferr)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expecting %d actual %d", http.StatusBadRequest, w.Code)
	}
}
