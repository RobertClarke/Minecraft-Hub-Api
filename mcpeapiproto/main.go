package main

import (
	"clarkezone-vs-com/mcpemapcore"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type MapListResponse struct {
	Maps []*mcpemapcore.Map
}

func GetMaps(wr http.ResponseWriter, r *http.Request) {
	var mapResponse MapListResponse
	fmt.Println("Host:" + r.Host)
	mapResponse.Maps, _, _ = mcpemapcore.GetMapsFromRedis(0, 8, r.Host)
	bytes, err := json.Marshal(mapResponse)
	if err == nil {
		wr.Header().Set("Content-Type", "application/json")
		wr.Write(bytes)
	} else {
		log.Fatal(err)
	}
}

type LogHandler struct {
}

func (LogHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	fmt.Println("file")
	http.FileServer(http.Dir("."))
}

func main() {
	http.HandleFunc("/getmaplist", GetMaps)
	// use http.stripprefix to redirect
	//http.Handle("/maps/", http.FileServer(http.Dir(".")))
	var logger LogHandler
	http.Handle("/maps/", logger)
	http.Handle("/mapimages/", http.FileServer(http.Dir(".")))
	panic(http.ListenAndServe(":8080", nil))
}
