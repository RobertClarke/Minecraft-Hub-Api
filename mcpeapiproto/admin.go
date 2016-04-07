package main

import (
	"clarkezone-vs-com/mcpemapcore"
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
