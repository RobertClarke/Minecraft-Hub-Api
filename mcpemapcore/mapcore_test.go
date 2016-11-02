package mcpemapcore

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
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

func prepare() (*CreateMapService, *log.Logger) {
	conn.Do("flushdb")
	logger := log.New(os.Stdout, "trace:", log.Ldate|log.Ltime|log.Lshortfile)
	tb := CreateRedisBackendWithDatabase(0)
	mapservice := NewCreateMapServiceWithBackend(tb, logger)
	return &mapservice, logger
}

func getHash(path string, logger *log.Logger) string {
	logger.Printf("getting hash for map")
	mapBytes, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Fatal("Test couldn't check checksum")
	}
	chkSum := md5.Sum(mapBytes)
	sh := fmt.Sprintf("%x", chkSum)
	return sh
}

func TestCreateMapBadUser(t *testing.T) {
	mapservice, logger := prepare()

	newMap := NewMap{
		Title:             "",
		Description:       "",
		MapFilename:       "",
		MapImageFileNames: nil}

	_, err := (*mapservice).CreateMap(nil, &newMap)
	if err == nil || err.Error() != "not authenticated, no user" {
		logger.Fatal("createmap not checking for nil user")
		t.Fail()
	}
}

func TestCreateMapEmptyWithNoImages(t *testing.T) {
	mapservice, logger := prepare()

	u := User{}
	newMap := NewMap{
		Title:             "",
		Description:       "",
		MapFilename:       "",
		MapImageFileNames: nil}

	_, err := (*mapservice).CreateMap(&u, &newMap)
	if err == nil || err.Error() != "must have at least 1 map image" {
		logger.Fatal("createmap not checking for images")
		t.Fail()
	}
}

func TestCreateMapEmptyWithNoTitle(t *testing.T) {
	mapservice, logger := prepare()

	u := User{}
	newMap := NewMap{
		Title:             "",
		Description:       "",
		MapFilename:       "",
		MapImageFileNames: nil}

	newMap.MapImageFileNames = []string{"image1.png", "image2.png"}

	_, err := (*mapservice).CreateMap(&u, &newMap)
	if err == nil || err.Error() != "map must have a title" {
		logger.Fatal("createmap not checking for empty title")
		t.Fail()
	}
}

func TestCreateMapEmptyWithNoDescription(t *testing.T) {
	mapservice, logger := prepare()

	u := User{}
	newMap := NewMap{
		Title:             "",
		Description:       "",
		MapFilename:       "",
		MapImageFileNames: nil}

	newMap.MapImageFileNames = []string{"image1.png", "image2.png"}

	newMap.Title = "Test Title"

	_, err := (*mapservice).CreateMap(&u, &newMap)
	if err == nil || err.Error() != "map must have a description" {
		logger.Fatal("createmap not checking for empty description " + err.Error())
		t.Fail()
	}
}

func TestCreateMapEmptyWithNoFilename(t *testing.T) {
	mapservice, logger := prepare()

	u := User{}
	newMap := NewMap{
		Title:             "Title",
		Description:       "Description",
		MapFilename:       "",
		MapImageFileNames: []string{"image1.png", "image2.png"},
	}

	_, err := (*mapservice).CreateMap(&u, &newMap)
	if err == nil || err.Error() != "map must have a filename" {
		logger.Fatal("createmap not checking for empty filename " + err.Error())
		t.Fail()
	}
}

func TestCreateMapEmptyWithBadFilename(t *testing.T) {
	mapservice, logger := prepare()

	u := User{}
	newMap := NewMap{
		Title:             "Title",
		Description:       "Description",
		MapFilename:       "xxxxx",
		MapImageFileNames: []string{"image1.png", "image2.png"},
	}

	_, err := (*mapservice).CreateMap(&u, &newMap)
	if err == nil || err.Error() != "map file doesn't exist" {
		logger.Fatal("createmap not checking for empty filename " + err.Error())
		t.Fail()
	}
}

func TestCreateMapEmptyWithMissingChecksum(t *testing.T) {
	mapservice, logger := prepare()

	u := User{}
	newMap := NewMap{
		Title:             "Title",
		Description:       "Description",
		MapFilename:       "m4sBABVgAAA=.zip",
		MapImageFileNames: []string{"image1.png", "image2.png"},
	}

	_, err := setupTestFiles(&newMap, logger)

	_, err = (*mapservice).CreateMap(&u, &newMap)
	if err == nil || err.Error() != "checksum doesn't match" {
		logger.Fatal("checksum doesn't match " + err.Error())
		t.Fail()
	}
	cleanupTestDir()
}

func TestSetupTestFile(t *testing.T) {
	_, logger := prepare()

	newMap := NewMap{
		Title:             "Title",
		Description:       "Description",
		MapFilename:       "m4sBABVgAAA=.zip",
		MapImageFileNames: []string{"image1.png", "image2.png"},
	}

	dir, _ := os.Getwd()
	testDir := path.Join(dir, "testdata")
	downloadDir := path.Join(dir, "downloads")
	_, err := os.Stat(path.Join(testDir, newMap.MapImageFileNames[0]))
	if os.IsNotExist(err) {
		t.Fatal(err)
	}

	_, err = setupTestFiles(&newMap, logger)

	_, err = os.Stat(path.Join(downloadDir, newMap.MapImageFileNames[0]))
	if os.IsNotExist(err) {
		t.Fatal(err)
	}
}

func TestCreateMapEmptyWithMissingImage(t *testing.T) {
	mapservice, logger := prepare()

	u := User{}
	newMap := NewMap{
		Title:             "Title",
		Description:       "Description",
		MapFilename:       "m4sBABVgAAA=.zip",
		MapImageFileNames: []string{"image.png", "image2.png"},
	}

	testDir, _ := setupTestFiles(&newMap, logger)

	newMap.MapChecksum = getHash(path.Join(testDir, newMap.MapFilename), logger)

	_, err := (*mapservice).CreateMap(&u, &newMap)
	if err == nil || err.Error()[0:23] != "map image doesn't exist" {
		logger.Fatal("createmap not checking for images:")
		t.Fail()
	}
}

func TestCreateMapEmptyWithGoodChecksum(t *testing.T) {
	mapservice, logger := prepare()

	u := User{}
	newMap := NewMap{
		Title:             "Title",
		Description:       "Description",
		MapFilename:       "m4sBABVgAAA=.zip",
		MapImageFileNames: []string{"image1.png", "image2.png"},
	}

	testDir, _ := setupTestFiles(&newMap, logger)

	newMap.MapChecksum = getHash(path.Join(testDir, newMap.MapFilename), logger)

	_, err := (*mapservice).CreateMap(&u, &newMap)
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
	cleanupTestDir()
}

func setupTestFiles(newMap *NewMap, logger *log.Logger) (string, error) {
	logger.Printf("setuptestfiles\n")
	dir, _ := os.Getwd()
	testDir := path.Join(dir, "testdata")
	downloadDir := path.Join(dir, "downloads")
	//mapDir := path.Join(dir, "maps")
	//mapImagesDir := path.Join(dir, "mapimages")
	os.Mkdir(downloadDir, 0777)

	for i := range newMap.MapImageFileNames {
		name := newMap.MapImageFileNames[i]
		source := path.Join(testDir, name)
		ino, _ := os.Stat(source)
		if ino != nil {
			dest := path.Join(downloadDir, name)
			logger.Printf("Test is Copying %v from %v to %v", name, source, dest)
			err := copyFile(source, dest)
			if err != nil {
				logger.Panic(err)
			}
		} else {
			logger.Printf("Skipping %v\n", source)
		}
	}

	err := copyFile(path.Join(testDir, "m4sBABVgAAA=.zip"), path.Join(downloadDir, "m4sBABVgAAA=.zip"))
	if err != nil {
		return testDir, err
	} else {
		return testDir, nil
	}
}

func cleanupTestDir() {
	dir, _ := os.Getwd()
	downloadDir := path.Join(dir, "downloads")
	mapDir := path.Join(dir, "maps")
	mapImagesDir := path.Join(dir, "mapimages")

	os.RemoveAll(downloadDir)
	os.RemoveAll(mapDir)
	os.RemoveAll(mapImagesDir)
}
