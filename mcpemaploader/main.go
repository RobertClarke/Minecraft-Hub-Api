package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/robertclarke/Minecraft-Hub-Api/mcpemapcore"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	showstats := flag.Bool("stats", false, "displays database stats")
	doimport := flag.Bool("import", false, "Import records from mcpehub")
	doMySqlimport := flag.Bool("mysqlimport", false, "Import records from mcpehub direct from mysql")
	makeAdmin := flag.String("makeadmin", "", "If user with this username exists, make them admin")
	flag.Parse()
	if *showstats {
		dumpStats()
	} else if *doimport {
		GetWriteMapsFromService()
	} else if *doMySqlimport {
		GetWriteMapsFromMySqlService()
	} else if *makeAdmin != "" {
		makeUserAdmin(*makeAdmin)
	} else {
		flag.Usage()
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

func makeUserAdmin(userName string) {
	mcpemapcore.MakeUserAdmin(userName)
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

func GetWriteMapsFromMySqlService() {
	maps, _ := mcpemapcore.MySqlGetAllMapsInclUntested(7001, 10000, "")
	var count = 0
	for i := range maps {
		count++
		fmt.Printf("Map %v\n", count)
		themap := maps[i]
		fmt.Printf("downloaduri %v\n", themap.MapDownloadUri)
		success, hash := mcpemapcore.DownloadContent(themap.MapDownloadUri, "maps", "application/zip", "zip")
		if success {
			fmt.Println("Valid:"+themap.MapDownloadUri, false, hash)
			mcpemapcore.MySqlUpdateMapValid(themap, true, hash)
			//mcpemapcore.WriteNextMap(cool, true, hash)
		} else {
			mcpemapcore.MySqlUpdateMapValid(themap, false, hash)
			//mcpemapcore.WriteNextMap(cool, false, hash)
			fmt.Println("Invalid:"+themap.MapDownloadUri, false, hash)
		}
	}
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
		success, hash := mcpemapcore.DownloadContent(name, "maps", "application/zip", "zip")
		if success {
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
