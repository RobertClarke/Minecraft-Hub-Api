package main

import (
	"fmt"
	"net/http"
	"time"

	jwtauth "github.com/clarkezone/jwtauth-go"
)

func registerHelloHandlers(mux *http.ServeMux, auth *jwtauth.ApiSecurity) {
	mux.HandleFunc("/hello", apiCounter(auth.CorsOptions(helloServer)))
	mux.HandleFunc("/helloslow", apiCounter(auth.CorsOptions(slowHelloServer)))
}

func helloServer(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("hello world\n")
	fmt.Fprintf(w, "hello, world!\n")
}

func slowHelloServer(w http.ResponseWriter, req *http.Request) {
	time.Sleep(time.Duration(200 * time.Millisecond))
	fmt.Printf("slow hello world\n")
	fmt.Fprintf(w, "slow hello, world!\n")
}
