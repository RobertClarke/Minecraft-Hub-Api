package main

import (
	"clarkezone-vs-com/mcpemapcore"
	"encoding/json"
	"log"
	"net/http"

	"github.com/clarkezone/jwtauth"
)

func UpdateMapFromUpload(wr http.ResponseWriter, r *http.Request) {
	role := mcpemapcore.GetRole("Administrator")

	if jwtauth.IsInRole(role.Id, r) {
		//TODO: get the uploadid from json

	} else {

	}
}

func GetBadMapList(wr http.ResponseWriter, r *http.Request) {
	var role mcpemapcore.Role
	role = mcpemapcore.GetRole("Administrator")
	//fmt.Printf("GetBadMapList:role:%v %v\n", role, role.Id)

	if jwtauth.IsInRole(role.Id, r) {
		var mapResponse MapListResponse
		var err error
		//fmt.Printf("Request:%+v", r)
		//fmt.Println("Request: admin Get Bad Map List")
		mapResponse.Maps, _, err = mcpemapcore.GetBadMapsFromRedis(0, 8, r.Host)
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
	} else {
		wr.WriteHeader(http.StatusUnauthorized)
	}
}
