package mcpemapcore

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"testing"
)

func TestGetZipThruProxy(t *testing.T) {
	name := "http://www.mediafire.com/?p3ijl1h0s4iy10s"
	success, hash := DownloadContentRedirect(name, "maps", "application/zip", "zip")
	fmt.Printf("%v %v\n", hash, success)

}
func TestMySqlGetAllMaps(t *testing.T) {
	maps, err := MySqlGetAllMaps(0, 0, "")
	if err != nil {
		t.Fail()
	}
	fmt.Printf("got %v\n", len(maps))
}

func TestCreateMap(t *testing.T) {
	logger := log.New(os.Stdout, "TRACE:", log.Ldate|log.Ltime|log.Lshortfile)
	//tb := testBackend{}
	//	mapservice := NewCreateMapServiceWithBackend(tb, logger)

	//	u := User{}
	newMap := NewMap{
		Title:             "The test map",
		Description:       "Describing the map",
		MapFilename:       "filename",
		MapImageFileNames: []string{"image1.png", "image2.png"},
	}

	dir, _ := os.Getwd()
	testDir := path.Join(dir, "testdata")
	downloadDir := path.Join(dir, "downloads")
	os.Mkdir(downloadDir, 0666)

	for i := range newMap.MapImageFileNames {
		name := newMap.MapImageFileNames[i]
		logger.Printf("Copying %v", name)
		source := path.Join(testDir, "")
		dest := path.Join(downloadDir, name)
		err := copyFile(source, dest)
		if err != nil {
			log.Panic(err)
		}
	}

	//mapservice.CreateMap(&u, &newMap)

	//Cleanup
}

func copyFile(source, destination string) error {
	var err error
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file.Close()
	output, err := os.Create(destination)
	defer output.Close()
	if err != nil {
		return err
	}
	_, err = io.Copy(output, file)
	return err
}

type testBackend struct {
}

func (t testBackend) CreateMap(user *User,
	newMap *NewMap) {
}
