package main

import (
	"clarkezone-vs-com/mcpemapcore"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/clarkezone/jwtauth"
)

type AdminMapListResponse struct {
	Maps []*mcpemapcore.AdminMap
}

func AdminUpdateMapFromUpload(wr http.ResponseWriter, r *http.Request) {
	role := mcpemapcore.GetRole("Administrator")

	if jwtauth.IsInRole(role.Id, r) {
		user := GetUser(wr, r)
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

		mcpemapcore.AdminUpdateMap(user, parsedParams.MapId, parsedParams.UploadHash)

	} else {
		wr.WriteHeader(http.StatusUnauthorized)
	}
}

func AdminGetBadMapList(wr http.ResponseWriter, r *http.Request) {
	var role mcpemapcore.Role
	role = mcpemapcore.GetRole("Administrator")

	if jwtauth.IsInRole(role.Id, r) {
		var mapResponse AdminMapListResponse
		var err error
		mapResponse.Maps, _, err = mcpemapcore.AdminGetBadMaps(0, 8, r.Host)
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

func AdminGetEditedMapList(wr http.ResponseWriter, r *http.Request) {
	var role mcpemapcore.Role
	role = mcpemapcore.GetRole("Administrator")

	if jwtauth.IsInRole(role.Id, r) {
		var mapResponse AdminMapListResponse
		var err error
		mapResponse.Maps, _, err = mcpemapcore.AdminGetEditedMaps(0, 8, r.Host)
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
