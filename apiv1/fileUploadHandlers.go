package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	jwtauth "github.com/clarkezone/jwtauth-go"
	"github.com/robertclarke/Minecraft-Hub-Api/mcpemapcore"
)

func registerFileUploadHandlers(mux *http.ServeMux, auth *jwtauth.ApiSecurity) {
	mux.HandleFunc("/upload", auth.CorsOptions(apiCounter(auth.RequireTokenAuthentication(handleUpload))))
}

func handleUpload(w http.ResponseWriter, req *http.Request) {
	var err error
	user := getUser(w, req)

	if user != nil {

		fmt.Printf("we have a user with name %v\n", user.Username)

		//TODO: switch to production file backend
		if !mcpemapcore.Exists("uploads") {
			err = os.Mkdir("uploads", 0777)
			if err != nil {
				log.Fatal(err)
			}
		}

		if !mcpemapcore.Exists("uploads/" + user.Username) {
			err = os.Mkdir("uploads/"+user.Username, 0777)
			if err != nil {
				log.Fatal(err)
			}
		}

		formFile, formHeader, err := req.FormFile("TheFile")
		ext := path.Ext(formHeader.Filename)
		filename, _ := genUUID()
		filename += ext
		if err != nil {
			log.Fatal(err)
		}

		defer formFile.Close()

		osFile, err := os.Create("uploads/" + user.Username + "/" + filename)
		if err != nil {
			log.Fatal(err)
		}
		defer osFile.Close()

		count, err := io.Copy(osFile, formFile)

		if err != nil {
			log.Fatal(err)
		}

		type FileName struct {
			Filename string
		}

		fn := FileName{Filename: filename}

		buffer, _ := json.Marshal(fn)

		w.Write(buffer) //ignore errors
		fmt.Printf("Upload complete: %v bytes\n", count)
		uploadCounter.Add(1)
	} else {
		fmt.Printf("shit no user")
		log.Fatal()
	}
}

func genUUID() (string, error) {

	uuid := make([]byte, 16)

	n, err := rand.Read(uuid)

	if n != len(uuid) || err != nil {
		return "", err
	}

	// TODO: verify the two lines implement RFC 4122 correctly

	uuid[8] = 0x80 // variant bits see page 5

	uuid[4] = 0x40 // version 4 Pseudo Random, see page 7

	return hex.EncodeToString(uuid), nil

}

func getUser(wr http.ResponseWriter, r *http.Request) *mcpemapcore.User {
	userid := r.Header.Get("userid")
	//TODO switch to production auth backend
	user, err := mcpemapcore.LoadUserInfo(userid)
	if hasFailed(wr, err) {
		return nil
	}
	return user
}
