package main

import (
	"clarkezone-vs-com/mcpemapcore"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func UpdateMapFromUpload(wr http.ResponseWriter, r *http.Request) {
	userid := r.Header.Get("userid")
	user, _ := mcpemapcore.LoadUserInfo(userid)

	//TODO: get the role from an application global place
	role := mcpemapcore.Role{}
	role.Id = 1

	if user.IsInRole(role) {
		//TODO: get the uploadid from json

	} else {

	}
}

func GetBadMapList(wr http.ResponseWriter, r *http.Request) {
	var mapResponse MapListResponse
	var err error
	//fmt.Printf("Request:%+v", r)
	fmt.Println("Request: admin Get Bad Map List")
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
}
