package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	jwtauth "github.com/clarkezone/jwtauth-go"
)

type MapListRequest struct {
	QueryName string
	Skip      int64
	Take      int64
}

//MapListResponse represents the wireformat for map responses
type MapListResponse struct {
	Maps []*Map
}

type errorObj struct {
	Message string
}

func registerGetMapsHandlers(mux *http.ServeMux, auth *jwtauth.ApiSecurity) {
	mux.HandleFunc("/getmapsquery", apiCounter(auth.CorsOptions(getMapsArgs)))
}

func getMapsArgs(wr http.ResponseWriter, r *http.Request) {
	var b MapListRequest
	var mapService = CreateGetMapService()
	var mapResponse MapListResponse
	var err error

	bodyBytes, err := ioutil.ReadAll(r.Body)

	if hasFailed(wr, err) {
		return
	}
	err = json.Unmarshal(bodyBytes, &b)

	if hasFailed(wr, err) {
		log.Printf("Failed to unmarshall" + err.Error())
		return
	}

	mapResponse.Maps, _, err = mapService.GetAllMapsQuery(b.Skip, b.Take, r.Host, b.QueryName)
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
