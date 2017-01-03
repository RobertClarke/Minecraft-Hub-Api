package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"

	jwtauth "github.com/clarkezone/jwtauth-go"
)

func registerFileUploadHandlers(mux *http.ServeMux, auth *jwtauth.ApiSecurity) {
	mux.HandleFunc("/upload", auth.CorsOptions(apiCounter(auth.RequireTokenAuthentication(handleUpload))))
}

func handleUpload(w http.ResponseWriter, req *http.Request) {
	userService := CreateGetUserService()
	user := userService.GetUser(w, req)

	fileService := CreateFileService()

	if user != nil {

		fmt.Printf("we have a user with name %v\n", user.Username)

		formFile, formHeader, err := req.FormFile("TheFile")
		if err != nil {
			log.Fatal(err)
		}

		defer formFile.Close()

		ext := path.Ext(formHeader.Filename)
		osFile, filename, err := fileService.CreateUploadFileForUser(user, ext)
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
