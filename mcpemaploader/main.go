package main

import (
	"clarkezone-vs-com/mcpemapcore"
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	showstats := flag.Bool("stats", false, "displays database stats")
	doimport := flag.Bool("import", false, "Import records from mcpehub")
	flag.Parse()
	if *showstats {
		dumpStats()
	}
	if *doimport {
		GetWriteMapsFromService()
	}
	//	m, _ := mcpemapcore.GetMapFromRedis("2")
	//	maps, remaining, _ := mcpemapcore.GetMapsFromRedis(0, 4)
	//	fmt.Printf("Returned Maps:%v remaining maps:%v\n", len(maps), remaining)
	//	for i := range maps {
	//		fmt.Println(maps[i].MapTitle)
	//	}
	//	maps, remaining, _ = mcpemapcore.GetMapsFromRedis(4, 4)
	//	fmt.Printf("Returned Maps:%v remaining maps:%v\n", len(maps), remaining)
	//	for i := range maps {
	//		fmt.Println(maps[i].MapTitle)
	//	}
}

func dumpStats() {
	s := mcpemapcore.GetStats()
	fmt.Printf("Total maps:%v\n", s.Total_maps)
	fmt.Printf("Total good maps:%v\n", s.Total_good)
	fmt.Printf("Total bad maps:%v\n", s.Total_bad)
	fmt.Printf("Total good tested maps:%v\n", s.Total_good_tested)
	fmt.Printf("Total bad tested maps:%v\n", s.Total_bad_tested)
	fmt.Printf("Total good featured maps:%v\n", s.Total_good_featured)
	fmt.Printf("Total bad featuered maps:%v\n", s.Total_bad_featured)
	//fmt.Printf("Total featured maps:%v", s.Total_maps)
}

func GetWriteMapsFromService() {
	fmt.Println("Getting maps")
	thebytes := Get("http://mcpehub.com/api/v1/list-maps?tested")

	var s interface{}
	err := json.Unmarshal(thebytes, &s)
	if err != nil {
		log.Fatal(err)
	}

	m := s.(map[string]interface{})

	food := m["MapList"]
	list := food.([]interface{})
	//first := list[5]
	for i := range list {
		cool := list[i].(map[string]interface{})
		name := cool["MapDownloadUri"].(string)
		success, hash := downloadMap(name)
		if success {
			fmt.Println("Valid:" + name)
			mcpemapcore.WriteNextMap(cool, true, hash)
		} else {
			mcpemapcore.WriteNextMap(cool, false, hash)
			fmt.Println("Invalid:"+name, false, hash)
		}
	}
	//fmt.Println(first)
	//	firstAsMap := first.(map[string]interface{})
	//	//fmt.Println(firstAsMap["Description"])
	//	fmt.Println(firstAsMap["MapDownloadUri"])
	//	//writeMapNext(firstAsMap)
	//
	//	//writeMap(2, firstAsMap)
	//	downloadMap(firstAsMap["MapDownloadUri"].(string))
}

func downloadMap(uri string) (bool, string) {
	resp, err := http.Get(uri)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	headerType := resp.Header.Get("Content-Type")
	fmt.Printf("Header MimeType:%v\n", headerType)
	if headerType == "application/zip" {
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		fn := md5.Sum([]byte(uri))
		filename := fmt.Sprintf("%x.zip", fn)
		hash := fmt.Sprintf("%x", fn)
		err = ioutil.WriteFile(filename, bytes, os.FileMode(0777))

		if err != nil {
			log.Fatal(err)
		}
		return true, hash
	}
	return false, ""
}

func iterate() {
	//	for k, v := range m {
	//		switch vv := v.(type) {
	//		case string:
	//			fmt.Println(k, "is string", vv)
	//		case int:
	//			fmt.Println(k, "is int", vv)
	//		case []interface{}:
	//			fmt.Println(k, "is an array:")
	//			for i, u := range vv {
	//				fmt.Println(i, u)
	//				break
	//			}
	//		default:
	//			fmt.Println(k, "is of a type I don't know how to handle")
	//		}
	//	}
}

func Get(s string) []byte {
	resp, err := http.Get(s)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return body
}
