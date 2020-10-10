package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMain(t *testing.T) {
	srv := httptest.NewUnstartedServer(nil)
	srvAddr := srv.URL

	srv1 := httptest.NewUnstartedServer(nil)
	srv1Addr := srv1.URL

	srv2 := httptest.NewUnstartedServer(nil)
	srv2Addr := srv2.URL

	srv.Config = &http.Server{Handler: wptHandler(srvAddr, srv1Addr, srv2Addr)}
	srv1.Config = &http.Server{Handler: wptHandler(srvAddr, srv1Addr, srv2Addr)}
	srv2.Config = &http.Server{Handler: wptHandler(srvAddr, srv1Addr, srv2Addr)}

	srv.Start()
	srv1.Start()
	srv2.Start()

	defer srv.Close()
	defer srv1.Close()
	defer srv2.Close()

	_, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJS(t *testing.T) {
	srv := httptest.NewUnstartedServer(nil)
	srvAddr := srv.URL

	srv1 := httptest.NewUnstartedServer(nil)
	srv1Addr := srv1.URL

	srv2 := httptest.NewUnstartedServer(nil)
	srv2Addr := srv2.URL

	srv.Config = &http.Server{Handler: wptHandler(srvAddr, srv1Addr, srv2Addr)}
	srv1.Config = &http.Server{Handler: wptHandler(srvAddr, srv1Addr, srv2Addr)}
	srv2.Config = &http.Server{Handler: wptHandler(srvAddr, srv1Addr, srv2Addr)}

	srv.Start()
	srv1.Start()
	srv2.Start()

	defer srv.Close()
	defer srv1.Close()
	defer srv2.Close()

	_, err := http.Get(srv.URL + "/resources/testharness.js")
	if err != nil {
		t.Fatal(err)
	}
}
