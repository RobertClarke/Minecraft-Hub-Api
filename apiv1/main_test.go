package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	jwtauth "github.com/clarkezone/jwtauth-go"
)

var (
	server   *httptest.Server
	reader   io.Reader //Ignore this for now
	usersURL string
)

func init() {
	mux := http.NewServeMux()
	registerAPIHandler(mux)
	server = httptest.NewServer(mux) //Creating new server with the user handlers
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestHello(t *testing.T) {
	res, err := http.Get(server.URL + "/hello")
	if err != nil {
		log.Fatal(err)
	}

	greeting, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Greeting:%s\n", greeting)

	if string(greeting) != "hello, world!\n" {
		log.Fatal("Wrong response")
	}
}

func TestGetMapsFilter(t *testing.T) {
	payload := MapListRequest{}
	payload.QueryName = "c:7"
	payload.Skip = 0
	payload.Take = 20
	byt, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err)
	}

	res, err := http.Post(server.URL+"/getmapsquery", "application/json", bytes.NewReader(byt))
	if err != nil {
		log.Fatal(err)
	}

	greeting, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	response := MapListResponse{}
	err = json.Unmarshal(greeting, &response)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("JSON Response contains %v items", len(response.Maps))
	if len(response.Maps) == 0 {
		log.Fatal("test failed; no maps")
	}
}

func TestFileUpload(t *testing.T) {
	//to run this test do sudo -E bash
	//go test -run FileUpload -v

	//TODO: test / create dir on "server"

	apiUploadRoot := "/var/www/minecrafthub.com/public/uploads/apiUploadRoot"
	i, err := ioutil.ReadDir(apiUploadRoot)

	if err != nil {
		log.Fatal("can't access directory: use sudo -E bash to run this test")
	}

	if len(i) > 0 {
		log.Fatal("Directory not empty")
	}

	mux := http.NewServeMux()

	registerAPIHandler(mux)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	url := ts.URL + "/upload"
	err = upload(url, "testdata/m4sBABVgAAA=.zip", "1")
	if err != nil {
		log.Fatal("upload failed")
	}

	i, err = ioutil.ReadDir(apiUploadRoot)

	if err != nil {
		log.Fatal("can't access upload dir")
	}

	if len(i) != 1 {
		log.Fatal("file count incorrect post upload")
	}

	log.Printf("Size %v\n", i[0].Size())

	if i[0].Size() != 1550890 {
		log.Fatal("uploaded file incorrect size")
	}
}

func TestExample(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "something failed", http.StatusInternalServerError)
	}

	req, err := http.NewRequest("GET", "http://example.com/foo", nil)
	if err != nil {
		log.Fatal(err)
	}

	w := httptest.NewRecorder()
	handler(w, req)

	fmt.Printf("%d - %s", w.Code, w.Body.String())
}

//func testGet(ser *httptest.Server, param interface{}, userid string, t *testing.T) interface{} {
//	var currentAuth jwtauth.JwtAuthProvider
//
//	token, _ := currentAuth.GenerateToken(userid)
//	client := &http.Client{}
//
//	var req *http.Request
//
//	if param != nil {
//		jsonBytes, err := json.Marshal(param)
//
//		if err != nil {
//			log.Fatal(err)
//		}
//		req, _ = http.NewRequest("GET", ser.URL, bytes.NewBuffer(jsonBytes))
//	} else {
//		req, _ = http.NewRequest("GET", ser.URL, nil)
//	}
//
//	req.Header.Set("Authorization", "Bearer "+token)
//	req.Header.Set("Content-Type", "application/json")
//
//	result, _ := client.Do(req)
//	res, _ := ioutil.ReadAll(result.Body)
//	responsePayload = CreateMapResponse{}
//	json.Unmarshal(res, &responsePayload)
//	return responsePayload
//}

func upload(url, file string, userID string) (err error) {
	var currentAuth jwtauth.JwtAuthProvider
	token, _ := currentAuth.GenerateToken(userID)
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	// Add your image file
	f, err := os.Open(file)
	if err != nil {
		return
	}
	defer f.Close()
	fw, err := w.CreateFormFile("TheFile", file)
	if err != nil {
		return
	}
	if _, err = io.Copy(fw, f); err != nil {
		return
	}
	//// Add the other fields
	//if fw, err = w.CreateFormField("key"); err != nil {
	//	return
	//}
	//if _, err = fw.Write([]byte("KEY")); err != nil {
	//	return
	//}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	// Submit the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return
	}

	// Check the response
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", res.Status)
	}
	return
}
