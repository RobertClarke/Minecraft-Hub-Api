package main

import (
	"clarkezone-vs-com/mcpemapcore"
	"clarkezone-vs-com/redisauthprovider"
	"encoding/json"
	"fmt"
	"github.com/clarkezone/jwtauth"
	"io/ioutil"
	"log"
	"net/http"
	"path"
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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Printf("hello world\n")
	fmt.Fprintf(w, "hello, world!\n")
}

func main() {
	redisauth.RegisterAuthHandlers()
	http.HandleFunc("/hello", HelloServer)
	http.HandleFunc("/getmaplist", GetMaps)
	http.HandleFunc("/getfeaturedmaplist", GetMaps)
	http.HandleFunc("/getmostdownloaded", GetMaps)
	http.HandleFunc("/getmostfavorited", GetMaps)
	http.HandleFunc("/setfavoritemap", jwtauth.RequireTokenAuthentication(UpdateFavoriteMap))
	http.HandleFunc("/getuserfavorites", jwtauth.RequireTokenAuthentication(GetMaps))
	http.HandleFunc("/admin/updatemapfromupload", jwtauth.RequireTokenAuthentication(UpdateMapFromUpload))
	// use http.stripprefix to redirect
	//http.Handle("/maps/", http.FileServer(http.Dir(".")))
	http.Handle("/maps/", CreateLogHandler(http.FileServer(http.Dir("."))))
	http.Handle("/mapimages/", http.FileServer(http.Dir(".")))
	panic(http.ListenAndServeTLS(":8080", "dev.objectivepixel.com.crt", "dev.objectivepixel.com.key", nil))
	//panic(http.ListenAndServe(":8080", nil))
}
