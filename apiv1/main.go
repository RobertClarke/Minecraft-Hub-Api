package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	jwtauth "github.com/clarkezone/jwtauth-go"
	"github.com/dkumor/acmewrapper"
	redisauth "github.com/robertclarke/Minecraft-Hub-Api/redisauthprovider"
)

func main() {
	useSsl, help, port, hostName := parseFlags()

	if *help {
		flag.Usage()
		return
	}

	if *useSsl && *hostName == "" {
		fmt.Println("SSL requires a hostname; please specify with -hostname")
return
	}

	mux := http.NewServeMux()

	// Metrics
	registerMetrics(mux)

	// Authentication
	var provider = redisauth.RedisUserProvider{}
	auth := jwtauth.CreateApiSecurity(provider)
	auth.RegisterLoginHandlerMux(mux)
	redisauth.RegisterUserRegistrationHandler(mux)

	// Service Handlers
	registerHelloHandlers(mux, auth)

	// run server
	runServer(*hostName, useSsl, port, mux)
}

func parseFlags() (*bool, *bool, *int, *string) {
	useSsl := flag.Bool("ssl", false, "enable SSL")
	host := flag.String("hostname", "", "hostname for SSL")
	help := flag.Bool("?", false, "get help")
	port := flag.Int("port", -1, "port to listen on")
	flag.Parse()
	return useSsl, help, port, host
}

func runServer(hostName string, useSsl *bool, port *int, mux *http.ServeMux) {
	var actualPort int
	var err error
	var server *http.Server

	if *useSsl {
		if *port == -1 {
			actualPort = 443
		} else {
			actualPort = *port
		}

		listener, tlsconfig := configureTLS(hostName, actualPort)
		//log.Printf("Listening for TLS with cert for hostname %v port %v\n", tlsconfig.Hostname, tlsconfig.Port)
		server = &http.Server{
			Addr:      ":" + strconv.Itoa(actualPort),
			Handler:   mux,
			TLSConfig: tlsconfig,
		}
		err = server.Serve(listener)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		if *port == -1 {
			actualPort = 80
		} else {
			actualPort = *port
		}
		usePortStr := strconv.Itoa(actualPort)
		log.Printf("Listening for HTTP on %v\n", usePortStr)
		err = http.ListenAndServe(":"+usePortStr, mux)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func configureTLS(hostname string, port int) (net.Listener, *tls.Config) {
log.Printf("ConfigureTLS for port %v", port)
	w, err := acmewrapper.New(acmewrapper.Config{
		Domains: []string{hostname},
		Address: strconv.Itoa(port),

		TLSCertFile: hostname + ".crt",
		TLSKeyFile:  hostname + ".key",

		RegistrationFile: "user.reg",
		PrivateKeyFile:   "user.pem",

		TOSCallback: acmewrapper.TOSAgree,
	})

	if err != nil {
		log.Fatal("acmewrapper: ", err)
	}

	tlsconfig := w.TLSConfig()

	listener, err := tls.Listen("tcp", ":"+strconv.Itoa(port), tlsconfig)
	if err != nil {
		log.Fatal("Listener: ", err)
	}
	return listener, tlsconfig
}
