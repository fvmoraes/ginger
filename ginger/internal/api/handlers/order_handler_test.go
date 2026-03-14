package handlers_test

import (
	"net/http"
	"testing"

	"github.com/ginger-framework/ginger/pkg/testhelper"
)

func TestOrderHandler_List(t *testing.T) {
	// TODO: inject a mock service and create the handler
	// h := NewOrderHandler(mockSvc)
	// r := router.New()
	// h.Register(r)

	rec := testhelper.NewRequest(t, http.NotFoundHandler(), http.MethodGet, "/orders").Do()
	testhelper.AssertStatus(t, rec, http.StatusNotFound) // replace with real handler
}

func TestOrderHandler_Create(t *testing.T) {
	body := map[string]any{
		// TODO: fill with valid fields
	}
	rec := testhelper.NewRequest(t, http.NotFoundHandler(), http.MethodPost, "/orders").
		WithBody(body).
		Do()
	testhelper.AssertStatus(t, rec, http.StatusNotFound) // replace with real handler
}
