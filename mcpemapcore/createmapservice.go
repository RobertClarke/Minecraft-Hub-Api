package mcpemapcore

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

// CreateMapService Service for creating maps from uploads
type CreateMapService interface {
	CreateMap(user *User, request *NewMap) (string, error)
}

// NewMap CreateMap.CreateMap request payload
type NewMap struct {
	Title             string
	Description       string
	MapFilename       string
	MapChecksum       string
	MapImageFileNames []string
	MapImageChecksums []string
}

// NewCreateMapService creates new instance of service
func NewCreateMapService() CreateMapService {
	return createMapService{
		myBackend: currentBackend,
		tracer:    log.New(ioutil.Discard, "TRACE:", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// NewCreateMapServiceWithBackend creates new instance of service
func NewCreateMapServiceWithBackend(backend Backend, logger *log.Logger) CreateMapService {
	service := createMapService{myBackend: backend}
	service.tracer = logger
	return service
}

// implementation
type createMapService struct {
	myBackend Backend
	tracer    *log.Logger
}

func (s createMapService) CreateMap(user *User, newMap *NewMap) (mapid string, err error) {
	if user == nil {
		return "", errors.New("not authenticated, no user")
	}

	if user.Username == "" {
		return "", errors.New("not authenticated, user has no username")
	}

	if len(newMap.MapImageFileNames) == 0 {
		return "", errors.New("must have at least 1 map image")
	}

	if len(newMap.Title) == 0 {
		return "", errors.New("map must have a title")
	}

	if len(newMap.Description) == 0 {
		return "", errors.New("map must have a description")
	}

	if len(newMap.MapFilename) == 0 {
		return "", errors.New("map must have a filename")
	}

	dir, _ := os.Getwd()
	downloads := path.Join(path.Join(dir, "uploads"), user.Username)
	mapDir := path.Join(dir, "maps")
	mapImages := path.Join(dir, "mapimages")

	_, err = os.Stat(mapImages)
	if os.IsNotExist(err) {
		s.tracer.Println("Creating mapimages dir")
		os.Mkdir(mapImages, 0777)
	}

	_, err = os.Stat(mapDir)
	if os.IsNotExist(err) {
		s.tracer.Println("Creating mapdir")
		os.Mkdir(mapDir, 0777)
	}

	filePath := path.Join(downloads, newMap.MapFilename)
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		return "", errors.New(fmt.Sprintf("map file doesn't exist filename: %v server path %v", newMap.MapFilename, filePath))
	}

	s.tracer.Println(dir)

	var mapBytes []byte
	mapBytes, err = ioutil.ReadFile(filePath)
	if err != nil {
		return "", errors.New("error getting checksum for file")
	}
	chkSum := md5.Sum(mapBytes)
	sh := fmt.Sprintf("%x", chkSum)
	if sh != strings.ToLower(newMap.MapChecksum) {
		return "", errors.New(fmt.Sprintf("checksum doesn't match: Calculated: %v Provided %v", sh, newMap.MapChecksum))
	} else {
		s.tracer.Println("Hashes match")
	}

	// get md5 of mapname
	// move map and rename
	for i := range newMap.MapImageFileNames {
		name := newMap.MapImageFileNames[i]

		iFn := path.Join(downloads, name)
		_, err = os.Stat(iFn)
		if os.IsNotExist(err) {
			return "", errors.New("map image doesn't exist " + iFn)
		} else {
			s.tracer.Printf("map image %v exists\n", iFn)
		}

		s.tracer.Println("Looking at file " + name)
		md5Name := md5.Sum([]byte(name))
		hash := fmt.Sprintf("%x", md5Name)
		if len(newMap.MapImageChecksums) == 0 {
			newMap.MapImageChecksums = make([]string, len(newMap.MapImageFileNames))
		}
		newMap.MapImageChecksums[i] = hash
		filename := fmt.Sprintf("%v%v", hash, path.Ext(name))
		err := os.Rename(path.Join(downloads, name), path.Join(mapImages, filename))
		if err != nil {
			return "", err
		}
		newMap.MapImageFileNames[i] = filename
	}

	s.tracer.Println("Done with images")

	err = os.Rename(path.Join(downloads, newMap.MapFilename), path.Join(mapDir, strings.ToLower(newMap.MapChecksum))+".zip")
	if err != nil {
		return "", err
	}
	s.tracer.Println("Done with map rename")

	newMap.MapFilename = sh
	// add to database
	return s.myBackend.CreateMap(user, newMap)

	// return id
	return "", nil
}
