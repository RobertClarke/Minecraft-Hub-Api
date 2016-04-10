package main

import (
	"clarkezone-vs-com/mcpemapcore"
	"clarkezone-vs-com/redisauthprovider"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"

	"github.com/clarkezone/jwtauth"
)

type MapListResponse struct {
	Maps []*mcpemapcore.Map
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
		mapResponse.Maps, _, err = mcpemapcore.GetAllMapsFromRedis(0, 8, r.Host)
		break
	case "/getfeaturedmaplist":
		fmt.Println("Request: featured maps")
		mapResponse.Maps, _, err = mcpemapcore.GetFeaturedMapsFromRedis(0, 8, r.Host)
		break
	case "/getmostdownloaded":
		fmt.Println("Request: most downloaded")
		mapResponse.Maps, _, err = mcpemapcore.GetMostDownloadedMapsFromRedis(0, 8, r.Host)
		break
	case "/getmostfavorited":
		fmt.Println("Request: most favorited")
		mapResponse.Maps, _, err = mcpemapcore.GetMostFavoritedMapsFromRedis(0, 8, r.Host)
		break
	case "/getuserfavorites":
		userid := r.Header.Get("userid")
		user, err := mcpemapcore.LoadUserInfo(userid)
		if hasFailed(wr, err) {
			return
		}

		fmt.Printf("Request: user favorites for user %v\n", user.Id)
		mapResponse.Maps, err = mcpemapcore.GetFavoriteMaps(user, r.Host)
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

func main() {
	useSsl := flag.Bool("ssl", false, "enable SSL")
	flag.Parse()
	var provider = redisauth.RedisUserProvider{}
	auth := jwtauth.CreateApiSecurity(provider)
	auth.RegisterLoginHandlers()
	redisauth.RegisterUserRegistrationHandler()

	http.HandleFunc("/hello", auth.CorsOptions(HelloServer))
	http.HandleFunc("/getmaplist", auth.CorsOptions(GetMaps))
	http.HandleFunc("/getfeaturedmaplist", auth.CorsOptions(GetMaps))
	http.HandleFunc("/getmostdownloaded", auth.CorsOptions(GetMaps))
	http.HandleFunc("/getmostfavorited", auth.CorsOptions(GetMaps))
	http.HandleFunc("/setfavoritemap", auth.CorsOptions(auth.RequireTokenAuthentication(UpdateFavoriteMap)))
	http.HandleFunc("/getuserfavorites", auth.CorsOptions(auth.RequireTokenAuthentication(GetMaps)))
	http.HandleFunc("/admin/getbadmaplist", auth.CorsOptions(auth.RequireTokenAuthentication(GetBadMapList)))
	http.HandleFunc("/admin/updatemapfromupload", auth.CorsOptions(auth.RequireTokenAuthentication(UpdateMapFromUpload)))
	// use http.stripprefix to redirect
	//http.Handle("/maps/", http.FileServer(http.Dir(".")))
	http.Handle("/maps/", CreateLogHandler(http.FileServer(http.Dir("."))))
	http.Handle("/mapimages/", http.FileServer(http.Dir(".")))
	if *useSsl {
		fmt.Printf("Listening for TLS on 8080\n")
		panic(http.ListenAndServeTLS(":8080", "dev.objectivepixel.com.crt", "dev.objectivepixel.com.key", nil))
	} else {
		fmt.Printf("Listening for HTTP on 8080\n")
		panic(http.ListenAndServe(":8080", nil))
	}
}
