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
	mysqlauth "github.com/robertclarke/Minecraft-Hub-Api/mysqlauthprovider"
)

func main() {
	verifyMaps := getDatabaseUtilFlags()
	useSsl, help, port, hostName, ipAddress := parseFlags()

	if *help {
		flag.Usage()
		return
	}

	if *verifyMaps {
		fmt.Println("Verify maps")
		testAllMaps()
		return
	}

	if *useSsl && *hostName == "" {
		fmt.Println("SSL requires a hostname; please specify with -hostname")
		return
	}

	mux := http.NewServeMux()

	// Metrics
	registerMetrics(mux)

	registerAPIHandler(mux)

	// run server
	runServer(*ipAddress, *hostName, useSsl, port, mux)
}

func init() {
	// Ensure upload directory created to ensure verification of permisions early
	_ = CreateFileService()
}

func registerAPIHandler(mux *http.ServeMux) *jwtauth.ApiSecurity {
	// Authentication
	//var provider = redisauth.RedisUserProvider{}
	var provider = mysqlauth.MysqlAuthProvider{}
	auth := jwtauth.CreateApiSecurity(provider)
	auth.RegisterLoginHandlerMux(mux)
	//mysqlauth.RegisterUserRegistrationHandler(mux) <-- user registration needs hooking up to the real db
	//redisauth.RegisterUserRegistrationHandler(mux)

	// Service Handlers
	registerHelloHandlers(mux, auth)
	registerFileUploadHandlers(mux, auth)
	registerGetMapsHandlers(mux, auth)
	return auth
}

func parseFlags() (*bool, *bool, *int, *string, *string) {
	useSsl := flag.Bool("ssl", false, "enable SSL")
	host := flag.String("hostname", "", "hostname for SSL")
	help := flag.Bool("?", false, "get help")
	port := flag.Int("port", -1, "port to listen on")
	ipaddress := flag.String("ip", "", "IP address to bind server to")
	flag.Parse()
	return useSsl, help, port, host, ipaddress
}

func runServer(ipAddress, hostName string, useSsl *bool, port *int, mux *http.ServeMux) {
	var actualPort int
	var err error
	var server *http.Server

	if *useSsl {
		if *port == -1 {
			actualPort = 443
		} else {
			actualPort = *port
		}

		listener, tlsconfig := configureTLS(hostName, ipAddress, actualPort)
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
		log.Printf("Listening for HTTP on %v %v\n", ipAddress, usePortStr)

		err = http.ListenAndServe(ipAddress+":"+usePortStr, mux)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func configureTLS(hostname string, ipAddress string, port int) (net.Listener, *tls.Config) {
	log.Printf("ConfigureTLS for port %v", port)
	w, err := acmewrapper.New(acmewrapper.Config{
		Domains: []string{hostname},
		Address: ipAddress + ":" + strconv.Itoa(port),

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
