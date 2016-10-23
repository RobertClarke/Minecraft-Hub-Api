package mcpemapcore

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/garyburd/redigo/redis"
)

const (
	address = "127.0.0.1:6379"
)

var (
	conn        redis.Conn
	rolesByName map[string]*Role
	rolesById   map[int]*Role
	//currentBackend *MySqlBackend
	currentBackend Backend
)

func init() {

	//currentBackend = &MySqlBackend{}
	currentBackend = &RedisBackend{}

	fmt.Println("mcpemapcoreinit")

	rolesByName = make(map[string]*Role)
	rolesById = make(map[int]*Role)

	var role1 = Role{}
	role1.Id = 1
	role1.Name = "Administrator"
	rolesByName[role1.Name] = &role1
	rolesById[role1.Id] = &role1

	var role2 = Role{}
	role2.Id = 2
	role2.Name = "Contributor"
	rolesByName[role2.Name] = &role2
	rolesById[role2.Id] = &role2
}

type Map struct {
	Id             string
	MapTitle       string `redis:"map_title" db:"title"`
	Description    string `redis:"description"`
	MapDownloadUri string `redis:"mapdownloaduri"`
	MapOriginalUri string
	MapFileHash    string `redis:"mapfilehash"`
	Author         string `redis:"author"`
	AuthorUri      string `redis:"author_uri"`
	NumViews       int    `redis:"numviews"`
	Tested         bool   `redis:"tested"`
	Featured       bool   `redis:"featured"`
	DownloadCount  int64  `redis:"downloadcount" db:"downloads"`
	FavoriteCount  int64  `redis:"favoritecount"`

	MapImageUriList []*MapImage
}

type MapImage struct {
	MapImageUri  string `redis:"mapimageuri"`
	MapImageHash string `redis:"mapimagehash"`
}

type Stats struct {
	Total_tested        int64 `redis:"total_tested"`
	Total_bad_tested    int64 `redis:"total_bad_tested"`
	Total_bad           int64 `redis:"total_bad"`
	Total_maps          int64 `redis:"total_maps"`
	Total_bad_featured  int64 `redis:"total_bad_featured"`
	Total_good_tested   int64 `redis:"total_good_tested"`
	Total_good_featured int64 `redis:"total_good_featured"`
	Total_good          int64 `redis:"total_good"`
}

func init() {
	var err error
	conn, err = redis.Dial("tcp", address)
	if nil != err {
		log.Fatalln("Error: Connection to redis:", err)
	}
	fmt.Printf("redis is alive\n")
	if !Exists("maps") {
		err = os.Mkdir("maps", 0777)
		if err != nil {
			log.Fatal(err)
		}
	}
	if !Exists("mapimages") {
		err = os.Mkdir("mapimages", 0777)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func Exists(path string) bool {
	fmt.Printf("exists:%v\n", path)
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	} else if err != nil {
		fmt.Printf(err.Error())
		log.Fatal(err)
	}
	return true
}

// UpdateMapDownloadCount updates map download count
func UpdateMapDownloadCount(fileHash string) {
	currentBackend.UpdateMapDownloadCount(fileHash)
}

// GetAllMaps returns all maps
func GetAllMaps(start, count int64, siteRoot string) ([]*Map, int64, error) {
	maps, next, err := currentBackend.GetAllMaps(start, count, siteRoot)
	//return GetMapsFromRedis(start, count, siteRoot, "goodmapset", false)
	return maps, next, err
}

// GetFeaturedMaps returns featured maps
func GetFeaturedMaps(start, count int64, siteRoot string) ([]*Map, int64, error) {
	maps, next, err := currentBackend.GetFeaturedMaps(start, count, siteRoot)
	//return GetMapsFromRedis(start, count, siteRoot, "featuredmapset", false)
	return maps, next, err
}

// GetMostDownloadedMaps returns most downloaded maps
func GetMostDownloadedMaps(start, count int64, siteRoot string) ([]*Map, int64, error) {
	maps, next, err := currentBackend.GetMostDownloadedMaps(start, count, siteRoot)
	return maps, next, err
}

func GetMostFavoritedMaps(start, count int64, siteRoot string) ([]*Map, int64, error) {
	maps, next, err := currentBackend.GetMostFavoritedMaps(start, count, siteRoot)
	return maps, next, err
}

func GetFavoriteMaps(u *User, start, count int64, siteRoot string) ([]*Map, error) {
	maps, _, err := currentBackend.GetFavoriteMaps(u, start, count, siteRoot)
	return maps, err
}

func DownloadContent(uri string, dir string, acceptMime string, ext string) (bool, string) {
	resp, err := http.Get(uri)
	if err != nil {
		log.Printf("bad uri:%v error:%v\n", uri, err.Error())
		return false, ""
	}
	defer resp.Body.Close()
	headerType := resp.Header.Get("Content-Type")
	if headerType == acceptMime || acceptMime == "" {
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		fn := md5.Sum([]byte(uri))
		filename := fmt.Sprintf("%x.%v", fn, ext)
		hash := fmt.Sprintf("%x", fn)
		filepath := fmt.Sprintf("%v/%v", dir, filename)
		fmt.Println(filepath)
		err = ioutil.WriteFile(filepath, bytes, os.FileMode(0777))

		if err != nil {
			log.Fatal(err)
		}
		return true, hash
	} else {
		fmt.Printf("Bad MimeType:%v\n", headerType)
	}
	return false, ""
}

func DownloadContentRedirect(uri string, dir string, acceptMime string, ext string) (bool, string) {
	check := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			fmt.Printf("Redirect from %v to %v\n", r.URL.Opaque, r.URL.Path)
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	resp, err := check.Get(uri)

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	headerType := resp.Header.Get("Content-Type")
	if headerType == acceptMime || acceptMime == "" {
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		fn := md5.Sum([]byte(uri))
		filename := fmt.Sprintf("%x.%v", fn, ext)
		hash := fmt.Sprintf("%x", fn)
		filepath := fmt.Sprintf("%v/%v", dir, filename)
		fmt.Println(filepath)
		err = ioutil.WriteFile(filepath, bytes, os.FileMode(0777))

		if err != nil {
			log.Fatal(err)
		}
		return true, hash
	} else {
		fmt.Printf("Bad MimeType:%v\n", headerType)
	}
	return false, ""
}

func DownloadContentRedirectSearch(uri string, dir string, acceptMime string, ext string) (bool, string) {

	check := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			fmt.Printf("Redirect from %v to %v\n", r.URL.Opaque, r.URL.Path)
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	resp, err := check.Get(uri)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	headerType := resp.Header.Get("Content-Type")
	fmt.Printf("ContentType={0}\n", headerType)
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if headerType == acceptMime || acceptMime == "" {
		fn := md5.Sum([]byte(uri))
		filename := fmt.Sprintf("%x.%v", fn, ext)
		hash := fmt.Sprintf("%x", fn)
		filepath := fmt.Sprintf("%v/%v", dir, filename)
		fmt.Println(filepath)
		err = ioutil.WriteFile(filepath, bytes, os.FileMode(0777))

		if err != nil {
			log.Fatal(err)
		}
		return true, hash
	} else {
		fmt.Println("HTML detected.. searching")
		//fmt.Printf("%v", string(bytes))
		searchFor :=
			`(http://[a-z_\/0-9\-\#=&\.|,|;|\?|\!]*/.zip)`
		stepRegex, _ := regexp.Compile(searchFor)
		captures := stepRegex.FindStringSubmatch(string(bytes))
		if len(captures) > 0 {
			fmt.Printf("Found:%v\n", len(captures))
			for i := range captures {
				fmt.Printf("Capture: %v %v\n", i, captures[i])

			}
		} else {
			fmt.Printf("Not found")
		}

	}
	return false, ""
}

func UpdateFavoriteMap(u *User, mapId string, fav bool) error {
	return currentBackend.UpdateFavoriteMap(u, mapId, fav)
}

func LoadUserInfo(userId string) (*User, error) {
	return currentBackend.LoadUserInfo(userId)
}
