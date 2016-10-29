package mcpemapcore

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
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
	MapImageFileNames []string
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
	//tracer := log.New(os.Stdout, "TRACE:", log.Ldate|log.Ltime|log.Lshortfile)
	if user == nil {
		return "", errors.New("not authenticated, no user")
	}

	dir, _ := os.Getwd()
	downloads := path.Join(dir, "downloads")
	//mapDir := path.Join(dir, "maps")
	mapImages := path.Join(dir, "mapimages")

	s.tracer.Println(dir)

	// get md5 of mapname
	// move map and rename
	for i := range newMap.MapImageFileNames {
		name := newMap.MapImageFileNames[i]
		s.tracer.Print(name)
		md5Name := md5.Sum([]byte(name))
		hash := fmt.Sprintf("%x", md5Name)

		err := os.Rename(path.Join(downloads, name), path.Join(mapImages, hash))
		if err != nil {
			panic(err)
		}
	}

	// move map zip file

	// move all images

	// add to database
	s.myBackend.CreateMap(user, newMap)

	// return id
	return "", nil
}
