package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func GetDatabaseUtilFlags() *bool {
	sqlMapVerify := flag.Bool("sqlmapverify", false, "scan all maps in database and attempt to verify them all")
	return sqlMapVerify
}

func TestAllMaps() {
	service := CreateGetMapService()
	maps, _, err := service.GetAllMaps(1, 10, "")
	if err != nil {
		panic(err)
	}
	var count = 0
	for i := range maps {
		count++
		fmt.Printf("Map %v\n", count)
		themap := maps[i]
		fmt.Printf("Title %v downloaduri %v\n", themap.MapTitle, themap.MapDownloadUri)
		success, hash := DownloadContent(themap.MapDownloadUri, "maps", "application/zip", ".zip")
		if success {
			fmt.Println("Valid:"+themap.MapDownloadUri, false, hash)
			//		mcpemapcore.MySqlUpdateMapValid(themap, true, hash)
			//		//mcpemapcore.WriteNextMap(cool, true, hash)
		} else {
			//		mcpemapcore.MySqlUpdateMapValid(themap, false, hash)
			//		//mcpemapcore.WriteNextMap(cool, false, hash)
			fmt.Println("Invalid:"+themap.MapDownloadUri, false, hash)
		}
	}
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
		filename := fmt.Sprintf("%x%v", fn, ext)
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
