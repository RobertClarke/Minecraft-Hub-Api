package main

import (
	"clarkezone-vs-com/mcpemapcore"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
)

type MapListResponse struct {
	Maps []*mcpemapcore.Map
}

func GetMaps(wr http.ResponseWriter, r *http.Request) {
	var mapResponse MapListResponse
	//fmt.Printf("Request:%+v", r)
	switch r.RequestURI {
	case "/getmaplist":
		fmt.Println("Request: All maps")
		mapResponse.Maps, _, _ = mcpemapcore.GetAllMapsFromRedis(0, 8, r.Host)
		break
	case "/getfeaturedmaplist":
		fmt.Println("Request: featured maps")
		mapResponse.Maps, _, _ = mcpemapcore.GetFeaturedMapsFromRedis(0, 8, r.Host)
		break
	case "/getmostdownloaded":
		fmt.Println("Request: most downloaded")
		mapResponse.Maps, _, _ = mcpemapcore.GetMostDownloadedMapsFromRedis(0, 8, r.Host)
		break
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

func main() {
	http.HandleFunc("/getmaplist", GetMaps)
	http.HandleFunc("/getfeaturedmaplist", GetMaps)
	http.HandleFunc("/getmostdownloaded", GetMaps)
	// use http.stripprefix to redirect
	//http.Handle("/maps/", http.FileServer(http.Dir(".")))
	http.Handle("/maps/", CreateLogHandler(http.FileServer(http.Dir("."))))
	http.Handle("/mapimages/", http.FileServer(http.Dir(".")))
	panic(http.ListenAndServe(":8080", nil))
}
