package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdmin(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(auth.RequireTokenAuthentication(GetBadMapList)))
	defer ts.Close()

	testGet(ts, nil, "1", t)
}
