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
		Title:             "",
		Description:       "",
		MapFilename:       "",
		MapImageFileNames: nil}

	_, err := mapservice.CreateMap(nil, &newMap)
	if err == nil || err.Error() != "not authenticated, no user" {
		logger.Fatal("createmap not checking for nil user")
		t.Fail()
	}

	_, err = mapservice.CreateMap(&u, &newMap)
	if err == nil || err.Error() != "must have at least 1 map image" {
		logger.Fatal("createmap not checking for images")
		t.Fail()
	}

	newMap.MapImageFileNames = []string{"image1.png", "image2.png"}

	_, err = mapservice.CreateMap(&u, &newMap)
	if err == nil || err.Error() != "map must have a title" {
		logger.Fatal("createmap not checking for empty title")
		t.Fail()
	}

	newMap.Title = "Test Title"

	_, err = mapservice.CreateMap(&u, &newMap)
	if err == nil || err.Error() != "map must have a description" {
		logger.Fatal("createmap not checking for empty description " + err.Error())
		t.Fail()
	}

	newMap.Description = "Test description"

	_, err = mapservice.CreateMap(&u, &newMap)
	if err == nil || err.Error() != "map must have a filename" {
		logger.Fatal("createmap not checking for empty filename " + err.Error())
		t.Fail()
	}

	newMap.MapFilename = "ABVgAAA=.zip"

	_, err = mapservice.CreateMap(&u, &newMap)
	if err == nil || err.Error() != "map file doesn't exist" {
		logger.Fatal("createmap not checking for empty filename " + err.Error())
		t.Fail()
	}

	newMap.MapFilename = "m4sBABVgAAA=.zip"

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
		logger.Printf("Test is Copying %v from %v to %v", name, source, dest)
		err := copyFile(source, dest)
		if err != nil {
			logger.Panic(err)
		}
	}

	copyFile(path.Join(testDir, "m4sBABVgAAA=.zip"), path.Join(downloadDir, "m4sBABVgAAA=.zip"))

	//check bad map images
	newMap.MapImageFileNames = []string{"image.png", "image2.png"}

	_, err = mapservice.CreateMap(&u, &newMap)
	if err == nil || err.Error()[0:23] != "map image doesn't exist" {
		logger.Fatal("createmap not checking for images:" + err.Error()[0:23])
		t.Fail()
	}

	newMap.MapImageFileNames = []string{"image1.png", "image2.png"}

	_, err = mapservice.CreateMap(&u, &newMap)
	if err != nil {
		logger.Fatal("createmap failed: " + err.Error())
		t.Fail()
	}

	var savedMap *Map
	savedMap, err = GetMapFromRedis("1", "")
	if savedMap == nil {

		logger.Fatal("no map")
		t.Fail()
	}
	if err != nil {
		logger.Fatal("get map failed")
		t.Fail()
	}

	if newMap.Title != savedMap.MapTitle {
		logger.Fatal("title doesn't match")
		t.Fail()
	}

	if newMap.Description != savedMap.Description {
		logger.Fatal("description doesn't match")
		t.Fail()
	}

	if newMap.MapFilename != savedMap.MapFileHash {
		logger.Fatal("filename doesn't match")
		t.Fail()
	}

	//TODO: verify the data in redis

	conn.Do("flushdb")

	os.RemoveAll(downloadDir)
	os.RemoveAll(mapDir)
	os.RemoveAll(mapImagesDir)
}
