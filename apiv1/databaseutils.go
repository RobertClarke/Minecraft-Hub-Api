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

func getDatabaseUtilFlags() *bool {
	sqlMapVerify := flag.Bool("sqlmapverify", false, "scan all maps in database and attempt to verify them all")
	return sqlMapVerify
}

func testAllMaps() {
	service := CreateGetMapService()
	maps, _, err := service.GetAllMaps(0, 1, "")
	if err != nil {
		panic(err)
	}
	var count = 0
	for i := range maps {
		count++
		themap := maps[i]
		fmt.Printf("Map %v ID:%v Title:%v Result:", count, themap.Id, themap.MapTitle)
		success, hash := downloadContent(themap.MapDownloadUri, "maps", "application/zip", ".zip")
		if success {
			fmt.Printf("valid with hash %v URI %v\n", hash, themap.MapDownloadUri)
			//id, _ := strconv.Atoi(themap.Id) //comment this out for now
			//service.EnsureDirectDL(id) //comment this out for now
		}
	}
}

func downloadContent(uri string, dir string, acceptMime string, ext string) (bool, string) {
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
		//fmt.Println(filepath)
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
