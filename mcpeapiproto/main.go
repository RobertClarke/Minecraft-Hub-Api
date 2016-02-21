package main

import (
	"clarkezone-vs-com/mcpemapcore"
	"encoding/json"
	"log"
	"net/http"
)

type MapListResponse struct {
	Maps []*mcpemapcore.Map
}

func GetMaps(wr http.ResponseWriter, r *http.Request) {
	var mapResponse MapListResponse
	mapResponse.Maps, _, _ = mcpemapcore.GetMapsFromRedis(0, 8)
	bytes, err := json.Marshal(mapResponse)
	if err == nil {
		wr.Header().Set("Content-Type", "application/json")
		wr.Write(bytes)
	} else {
		log.Fatal(err)
	}
}

func main() {
	http.HandleFunc("/getmaplist", GetMaps)
	// use http.stripprefix to redirect
	http.Handle("/maps/", http.FileServer(http.Dir(".")))
	http.Handle("/mapimages/", http.FileServer(http.Dir(".")))
	panic(http.ListenAndServe(":8080", nil))
}
