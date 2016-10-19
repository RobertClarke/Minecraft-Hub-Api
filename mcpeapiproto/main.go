package main

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/robertclarke/Minecraft-Hub-Api/mcpemapcore"
	"github.com/robertclarke/Minecraft-Hub-Api/mysqlauthprovider"

	"net"

	"github.com/clarkezone/jwtauth-go"
	"github.com/dkumor/acmewrapper"
)

type MapListResponse struct {
	Maps []*mcpemapcore.Map
}

func SecureHello(wr http.ResponseWriter, r *http.Request) {
	fmt.Printf("UpdateFaveMap\n")

	userid := r.Header.Get("userid")
	_, err := mcpemapcore.LoadUserInfo(userid)
	if hasFailed(wr, err) {
		return
	}
	fmt.Fprintf(wr, "secure hello, world!\n")
}

func UpdateFavoriteMap(wr http.ResponseWriter, r *http.Request) {
	fmt.Printf("UpdateFaveMap\n")
	userid := r.Header.Get("userid")
	user, err := mcpemapcore.LoadUserInfo(userid)
	if hasFailed(wr, err) {
		return
	}

	type Body struct {
		MapId string
		Add   bool
	}

	var b Body

	bodyBytes, err := ioutil.ReadAll(r.Body)

	if hasFailed(wr, err) {
		return
	}
	err = json.Unmarshal(bodyBytes, &b)

	if hasFailed(wr, err) {
		fmt.Printf("Failed to unmarshall" + err.Error())
		return
	}
	err = mcpemapcore.UpdateFavoriteMap(user, b.MapId, b.Add)

	if hasFailed(wr, err) {
		return
	}
}

func GetMaps(wr http.ResponseWriter, r *http.Request) {
	var mapResponse MapListResponse
	var err error
	//fmt.Printf("Request:%+v", r)
	switch r.RequestURI {
	case "/getmaplist":
		fmt.Println("Request: All maps")
		mapResponse.Maps, _, err = mcpemapcore.GetAllMaps(0, 12, r.Host)
		break
	case "/getfeaturedmaplist":
		fmt.Println("Request: featured maps")
		mapResponse.Maps, _, err = mcpemapcore.GetFeaturedMaps(0, 8, r.Host)
		break
	case "/getmostdownloaded":
		fmt.Println("Request: most downloaded")
		mapResponse.Maps, _, err = mcpemapcore.GetMostDownloadedMaps(0, 8, r.Host)
		break
	case "/getmostfavorited":
		fmt.Println("Request: most favorited")
		mapResponse.Maps, _, err = mcpemapcore.GetMostFavoritedMaps(0, 8, r.Host)
		break
	case "/getuserfavorites":
		user := GetUser(wr, r)
		fmt.Printf("Request: user favorites for user %v\n", user.Id)
		mapResponse.Maps, err = mcpemapcore.GetFavoriteMaps(user, 0, 8, r.Host)
		break
	}
	if hasFailed(wr, err) {
		return
	}
	bytes, err := json.Marshal(mapResponse)
	if err == nil {
		wr.Header().Set("Content-Type", "application/json")
		wr.Write(bytes)
	} else {
		log.Fatal(err)
	}
}

func GetUser(wr http.ResponseWriter, r *http.Request) *mcpemapcore.User {
	userid := r.Header.Get("userid")
	user, err := mcpemapcore.LoadUserInfo(userid)
	if hasFailed(wr, err) {
		return nil
	}
	return user
}

type LogHandler struct {
	wrapper http.Handler
}

func CreateLogHandler(h http.Handler) http.Handler {
	logger := new(LogHandler)
	logger.wrapper = h
	return logger
}

func (h LogHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	_, fn := path.Split(r.RequestURI)
	wp := path.Ext(fn)
	fnnp := fn[:len(fn)-len(wp)]
	fmt.Println("file:" + wp)
	mcpemapcore.UpdateMapDownloadCount(fnnp)
	h.wrapper.ServeHTTP(rw, r)
}

func hasFailed(wr http.ResponseWriter, err error) bool {
	if err != nil {
		wr.WriteHeader(http.StatusInternalServerError)

		o := errorObj{err.Error()}

		buffer, _ := json.Marshal(o)

		wr.Write(buffer) //ignore errors

		return true
	}
	return false
}

type errorObj struct {
	Message string
}

func HelloServer(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("hello world\n")
	fmt.Fprintf(w, "hello, world!\n")
}

func Upload(w http.ResponseWriter, req *http.Request) {
	var err error
	user := GetUser(w, req)

	if user != nil {

		fmt.Printf("we have a user with name %v\n", user.Username)
		filename, _ := GenUUID()
		filename += ".zip"

		if !mcpemapcore.Exists("uploads") {
			err = os.Mkdir("uploads", 0777)
			if err != nil {
				log.Fatal(err)
			}
		}

		if !mcpemapcore.Exists("uploads/" + user.Username) {
			err = os.Mkdir("uploads/"+user.Username, 0777)
			if err != nil {
				log.Fatal(err)
			}
		}

		formFile, _, err := req.FormFile("TheFile")
		if err != nil {
			log.Fatal(err)
		}

		defer formFile.Close()

		osFile, err := os.Create("uploads/" + user.Username + "/" + filename)
		if err != nil {
			log.Fatal(err)
		}
		defer osFile.Close()

		count, err := io.Copy(osFile, formFile)

		if err != nil {
			log.Fatal(err)
		}

		type FileName struct {
			Filename string
		}

		fn := FileName{Filename: filename}

		buffer, _ := json.Marshal(fn)

		w.Write(buffer) //ignore errors
		fmt.Printf("Upload complete: %v bytes\n", count)
	} else {
		fmt.Printf("shit no user")
		log.Fatal()
	}
}

func CreateFromUpload(wr http.ResponseWriter, r *http.Request) {
	user := GetUser(wr, r)
	if user == nil {
		fmt.Printf("shit user is nil after get user\n")
	}
	type UpdateMapParams struct {
		MapId      int    `json:"mapId"`
		UploadHash string `json:"uploadHash"`
	}

	var parsedParams UpdateMapParams

	bytes, err := ioutil.ReadAll(r.Body)
	if hasFailed(wr, err) {
		fmt.Printf("Failed to read bytes" + err.Error())
		return
	}

	err = json.Unmarshal(bytes, &parsedParams)

	if hasFailed(wr, err) {
		fmt.Printf("Failed to unmarshall" + err.Error())
		return
	}

	fmt.Printf("Unmarshalled %+v\n", parsedParams)

	fmt.Printf("update map %v id with filehash %v\n", parsedParams.MapId, parsedParams.UploadHash)

	//err = mcpemapcore.AdminUpdateMap(user, parsedParams.MapId, parsedParams.UploadHash)

	if hasFailed(wr, err) {
		fmt.Printf("Failed to unmarshall%v\n", err.Error())
		return
	}
}

func GenUUID() (string, error) {

	uuid := make([]byte, 16)

	n, err := rand.Read(uuid)

	if n != len(uuid) || err != nil {
		return "", err
	}

	// TODO: verify the two lines implement RFC 4122 correctly

	uuid[8] = 0x80 // variant bits see page 5

	uuid[4] = 0x40 // version 4 Pseudo Random, see page 7

	return hex.EncodeToString(uuid), nil

}

func configureTLS(hostname string) (net.Listener, *tls.Config) {
	w, err := acmewrapper.New(acmewrapper.Config{
		Domains: []string{hostname},
		Address: ":443",

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

	listener, err := tls.Listen("tcp", ":443", tlsconfig)
	if err != nil {
		log.Fatal("Listener: ", err)
	}
	return listener, tlsconfig
}

func main() {
	var err error
	useSsl := flag.Bool("ssl", false, "enable SSL")
	flag.Parse()
	//var provider = redisauth.RedisUserProvider{}
	var provider = mysqlauth.MysqlAuthProvider{}
	auth := jwtauth.CreateApiSecurity(provider)

	mux := http.NewServeMux()
	auth.RegisterLoginHandlerMux(mux)
	//mysqlauth.RegisterUserRegistrationHandler(mux) <-- user registration needs hooking up to the real db
	mux.HandleFunc("/hello", auth.CorsOptions(HelloServer))
	mux.HandleFunc("/getmaplist", auth.CorsOptions(GetMaps))
	mux.HandleFunc("/getfeaturedmaplist", auth.CorsOptions(GetMaps))
	mux.HandleFunc("/getmostdownloaded", auth.CorsOptions(GetMaps))
	mux.HandleFunc("/getmostfavorited", auth.CorsOptions(GetMaps))
	mux.HandleFunc("/securehello", auth.CorsOptions(auth.RequireTokenAuthentication(SecureHello)))
	mux.HandleFunc("/setfavoritemap", auth.CorsOptions(auth.RequireTokenAuthentication(UpdateFavoriteMap)))
	mux.HandleFunc("/getuserfavorites", auth.CorsOptions(auth.RequireTokenAuthentication(GetMaps)))
	mux.HandleFunc("/admin/getbadmaplist", auth.CorsOptions(auth.RequireTokenAuthentication(AdminGetBadMapList)))
	mux.HandleFunc("/admin/geteditedmaplist", auth.CorsOptions(auth.RequireTokenAuthentication(AdminGetEditedMapList)))
	mux.HandleFunc("/admin/updatemapfromupload", auth.CorsOptions(auth.RequireTokenAuthentication(AdminUpdateMapFromUpload)))
	mux.HandleFunc("/upload", auth.CorsOptions(auth.RequireTokenAuthentication(Upload)))
	mux.HandleFunc("/createmapfromupload", auth.CorsOptions(auth.RequireTokenAuthentication(CreateFromUpload)))
	// use http.stripprefix to redirect
	//http.Handle("/maps/", http.FileServer(http.Dir(".")))
	mux.Handle("/maps/", CreateLogHandler(http.FileServer(http.Dir("."))))
	mux.Handle("/mapimages/", http.FileServer(http.Dir(".")))
	var server *http.Server
	if *useSsl {
		listener, tlsconfig := configureTLS("dev2.minecrafthub.com")
		//fmt.Printf("Listening for TLS with cert for hostname %v port %v\n", config.Hostname, config.Port)
		server = &http.Server{
			Addr:      ":" + strconv.Itoa(443),
			Handler:   mux,
			TLSConfig: tlsconfig,
		}
		err = server.Serve(listener)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Printf("Listening for HTTP on 8080\n")
		err = http.ListenAndServe(":8080", mux)
		if err != nil {
			log.Fatal(err)
		}
	}
}
