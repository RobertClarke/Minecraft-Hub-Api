package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	jwtauth "github.com/clarkezone/jwtauth-go"
)

type MapListResponse struct {
	Maps []*Map
}

type errorObj struct {
	Message string
}

func registerGetMapsHandlers(mux *http.ServeMux, auth *jwtauth.ApiSecurity) {
	mux.HandleFunc("/getmaplist", apiCounter(auth.CorsOptions(getMaps)))
	mux.HandleFunc("/getfeaturedmaplist", apiCounter(auth.CorsOptions(getMaps)))
	mux.HandleFunc("/getmostdownloaded", apiCounter(auth.CorsOptions(getMaps)))
	mux.HandleFunc("/getmostfavorited", apiCounter(auth.CorsOptions(getMaps)))
}

func getMaps(wr http.ResponseWriter, r *http.Request) {

	var mapService = CreateGetMapService()

	var mapResponse MapListResponse
	var err error
	//fmt.Printf("Request:%+v", r)
	switch r.RequestURI {
	case "/getmaplist":
		fmt.Println("Request: All maps")
		mapResponse.Maps, _, err = mapService.GetAllMaps(0, 20, r.Host)
		break
	case "/getfeaturedmaplist":
		fmt.Println("Request: featured maps")
		//mapResponse.Maps, _, err = mcpemapcore.GetFeaturedMaps(0, 8, r.Host)
		break
	case "/getmostdownloaded":
		fmt.Println("Request: most downloaded")
		//mapResponse.Maps, _, err = mcpemapcore.GetMostDownloadedMaps(0, 8, r.Host)
		break
	case "/getmostfavorited":
		fmt.Println("Request: most favorited")
		//mapResponse.Maps, _, err = mcpemapcore.GetMostFavoritedMaps(0, 8, r.Host)
		break
	case "/getuserfavorites":
		//user := GetUser(wr, r)
		//fmt.Printf("Request: user favorites for user %v\n", user.Id)
		//mapResponse.Maps, err = mcpemapcore.GetFavoriteMaps(user, 0, 8, r.Host)
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

func hasFailed(wr http.ResponseWriter, err error) bool {
	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		return true
	}
	return false
}
