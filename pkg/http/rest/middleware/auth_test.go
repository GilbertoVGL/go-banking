package middleware

import (
	"net/http"
	"net/http/httptest"

	"testing"
)

func TestAuthIsOk(t *testing.T) {
	a := func(w http.ResponseWriter, r *http.Request) {}
	req, err := http.NewRequest(http.MethodPost, "", nil)
	req.Header.Set("Authorization", "Bearer hahahahaha")

	if err != nil {
		t.Fatal(err)
	}

	dullHandler := http.HandlerFunc(a)
	authHandler := Auth(dullHandler)
	rr := httptest.NewRecorder()

	authHandler.ServeHTTP(rr, req)
	t.Errorf("RUIM")
}
