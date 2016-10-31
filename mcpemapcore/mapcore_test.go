package mcpemapcore

import (
	"fmt"
	"log"
	"os"
	"path"
	"testing"
)

func TestGetZipThruProxy(t *testing.T) {
	return
	name := "http://www.mediafire.com/?p3ijl1h0s4iy10s"
	success, hash := DownloadContentRedirect(name, "maps", "application/zip", "zip")
	fmt.Printf("%v %v\n", hash, success)
}

func TestMySqlGetAllMaps(t *testing.T) {
	return
	maps, err := MySqlGetAllMaps(0, 0, "")
	if err != nil {
		t.Fail()
	}
	fmt.Printf("got %v\n", len(maps))
}

func TestCreateMap(t *testing.T) {
	conn.Do("flushdb")
	logger := log.New(os.Stdout, "TRACE:", log.Ldate|log.Ltime|log.Lshortfile)
	tb := CreateRedisBackendWithDatabase(1)
	mapservice := NewCreateMapServiceWithBackend(tb, logger)
	u := User{}
	newMap := NewMap{
		Title:             "The test map",
		Description:       "Describing the map",
		MapFilename:       "m4sBABVgAAA=.zip",
		MapImageFileNames: []string{"image1.png", "image2.png"},
	}

	dir, _ := os.Getwd()
	testDir := path.Join(dir, "testdata")
	downloadDir := path.Join(dir, "downloads")
	mapDir := path.Join(dir, "maps")
	mapImagesDir := path.Join(dir, "mapimages")
	os.Mkdir(downloadDir, 0777)

	for i := range newMap.MapImageFileNames {
		name := newMap.MapImageFileNames[i]
		source := path.Join(testDir, name)
		dest := path.Join(downloadDir, name)
		logger.Printf("Copying %v from %v to %v", name, source, dest)
		err := copyFile(source, dest)
		if err != nil {
			log.Panic(err)
		}
	}

	copyFile(path.Join(testDir, "m4sBABVgAAA=.zip"), path.Join(downloadDir, "m4sBABVgAAA=.zip"))

	_, err := mapservice.CreateMap(&u, &newMap)
	if err != nil {
		log.Fatal("createmap failed")
		t.Fail()
	}

	var savedMap *Map
	savedMap, err = GetMapFromRedis("1", "")
	if savedMap == nil {

		log.Fatal("no map")
		t.Fail()
	}
	if err != nil {
		log.Fatal("get map failed")
		t.Fail()
	}

	if newMap.Title != savedMap.MapTitle {
		log.Fatal("title doesn't match")
		t.Fail()
	}

	if newMap.Description != savedMap.Description {
		log.Fatal("description doesn't match")
		t.Fail()
	}

	if newMap.MapFilename != savedMap.MapFileHash {
		log.Fatal("filename doesn't match")
		t.Fail()
	}

	//TODO: verify the data in redis

	conn.Do("flushdb")

	os.RemoveAll(downloadDir)
	os.RemoveAll(mapDir)
	os.RemoveAll(mapImagesDir)
}
