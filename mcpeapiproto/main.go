package main

import (
	"clarkezonegit/Minecraft-Hub-Api/mcpemapcore"
	"clarkezonegit/Minecraft-Hub-Api/mysqlauthprovider"
	"clarkezonegit/Minecraft-Hub-Api/redisauthprovider"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/clarkezone/jwtauth"
)

type MapListResponse struct {
	Maps []*mcpemapcore.Map
}

func UpdateFavoriteMap(wr http.ResponseWriter, r *http.Request) {
	fmt.Printf("UpdateFaveMap\n")
	userid := r.Header.Get("userid")
	user, err := mcpemapcore.LoadUserInfo(userid)
	if hasFailed(wr, err) {
		return
	}

	type Body struct {
		MapId string
		Add   bool
	}

	var b Body

	bodyBytes, err := ioutil.ReadAll(r.Body)

	if hasFailed(wr, err) {
		return
	}
	err = json.Unmarshal(bodyBytes, &b)

	if hasFailed(wr, err) {
		fmt.Printf("Failed to unmarshall" + err.Error())
		return
	}
	err = mcpemapcore.UpdateFavoriteMap(user, b.MapId, b.Add)

	if hasFailed(wr, err) {
		return
	}
}

func GetMaps(wr http.ResponseWriter, r *http.Request) {
	var mapResponse MapListResponse
	var err error
	//fmt.Printf("Request:%+v", r)
	switch r.RequestURI {
	case "/getmaplist":
		fmt.Println("Request: All maps")
		mapResponse.Maps, _, err = mcpemapcore.GetAllMaps(0, 12, r.Host)
		break
	case "/getfeaturedmaplist":
		fmt.Println("Request: featured maps")
		mapResponse.Maps, _, err = mcpemapcore.GetFeaturedMaps(0, 8, r.Host)
		break
	case "/getmostdownloaded":
		fmt.Println("Request: most downloaded")
		mapResponse.Maps, _, err = mcpemapcore.GetMostDownloadedMaps(0, 8, r.Host)
		break
	case "/getmostfavorited":
		fmt.Println("Request: most favorited")
		mapResponse.Maps, _, err = mcpemapcore.GetMostFavoritedMaps(0, 8, r.Host)
		break
	case "/getuserfavorites":
		user := GetUser(wr, r)
		fmt.Printf("Request: user favorites for user %v\n", user.Id)
		mapResponse.Maps, err = mcpemapcore.GetFavoriteMaps(user, r.Host)
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

func GetUser(wr http.ResponseWriter, r *http.Request) *mcpemapcore.User {
	userid := r.Header.Get("userid")
	user, err := mcpemapcore.LoadUserInfo(userid)
	if hasFailed(wr, err) {
		return nil
	}
	return user
}

type LogHandler struct {
	wrapper http.Handler
}

func CreateLogHandler(h http.Handler) http.Handler {
	logger := new(LogHandler)
	logger.wrapper = h
	return logger
}

func (h LogHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	//	_, fn := path.Split(r.RequestURI)
	//	wp := path.Ext(fn)
	//	fnnp := fn[:len(fn)-len(wp)]
	//	fmt.Println("file:" + wp)
	//	mcpemapcore.UpdateMapDownloadCount(fnnp)
	h.wrapper.ServeHTTP(rw, r)
}

func hasFailed(wr http.ResponseWriter, err error) bool {
	if err != nil {
		wr.WriteHeader(http.StatusInternalServerError)

		o := errorObj{err.Error()}

		buffer, _ := json.Marshal(o)

		wr.Write(buffer) //ignore errors

		return true
	}
	return false
}

type errorObj struct {
	Message string
}

func HelloServer(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("hello world\n")
	fmt.Fprintf(w, "hello, world!\n")
}

func Upload(w http.ResponseWriter, req *http.Request) {
	var err error
	user := GetUser(w, req)

	if user != nil {

		fmt.Printf("we have a user with name %v\n", user.Username)
		filename, _ := GenUUID()
		filename += ".zip"

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

		formFile, _, err := req.FormFile("TheFile")
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
	} else {
		fmt.Printf("shit no user")
		log.Fatal()
	}
}

func GenUUID() (string, error) {

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

func main() {
	useSsl := flag.Bool("ssl", false, "enable SSL")
	flag.Parse()
	//var provider = redisauth.RedisUserProvider{}
	var provider = mysqlauth.MysqlAuthProvider{}
	auth := jwtauth.CreateApiSecurity(provider)
	auth.RegisterLoginHandlers()
	redisauth.RegisterUserRegistrationHandler()

	http.HandleFunc("/hello", auth.CorsOptions(HelloServer))
	http.HandleFunc("/getmaplist", auth.CorsOptions(GetMaps))
	http.HandleFunc("/getfeaturedmaplist", auth.CorsOptions(GetMaps))
	http.HandleFunc("/getmostdownloaded", auth.CorsOptions(GetMaps))
	http.HandleFunc("/getmostfavorited", auth.CorsOptions(GetMaps))
	http.HandleFunc("/setfavoritemap", auth.CorsOptions(auth.RequireTokenAuthentication(UpdateFavoriteMap)))
	http.HandleFunc("/getuserfavorites", auth.CorsOptions(auth.RequireTokenAuthentication(GetMaps)))
	http.HandleFunc("/admin/getbadmaplist", auth.CorsOptions(auth.RequireTokenAuthentication(AdminGetBadMapList)))
	http.HandleFunc("/admin/geteditedmaplist", auth.CorsOptions(auth.RequireTokenAuthentication(AdminGetEditedMapList)))
	http.HandleFunc("/admin/updatemapfromupload", auth.CorsOptions(auth.RequireTokenAuthentication(AdminUpdateMapFromUpload)))
	http.HandleFunc("/upload", auth.CorsOptions(auth.RequireTokenAuthentication(Upload)))
	// use http.stripprefix to redirect
	//http.Handle("/maps/", http.FileServer(http.Dir(".")))
	http.Handle("/maps/", CreateLogHandler(http.FileServer(http.Dir("."))))
	http.Handle("/mapimages/", http.FileServer(http.Dir(".")))
	if *useSsl {
		fmt.Printf("Listening for TLS on 80\n")
		//panic(http.ListenAndServeTLS(":80", "dev.objectivepixel.com.crt", "dev.objectivepixel.com.key", nil))
		panic(http.ListenAndServeTLS(":443", "dev2.minecrafthub.com.crt", "dev2.minecrafthub.com.key", nil))
	} else {
		fmt.Printf("Listening for HTTP on 80\n")
		panic(http.ListenAndServe(":80", nil))
	}
}
