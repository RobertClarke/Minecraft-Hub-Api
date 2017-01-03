package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path"
)

type FileService interface {
	CreateUploadFileForUser(u *User, ext string) (f *os.File, fn string, e error)
}

type productionFileService struct {
	uploadRoot    string
	apiUploadRoot string
}

// CreateFileService creates a new instance of a FileService
func CreateFileService() FileService {
	prod := productionFileService{
		uploadRoot:    "/var/www/minecrafthub.com/public/uploads/",
		apiUploadRoot: "/var/www/minecrafthub.com/public/uploads/apiUploadRoot",
	}
	prod.ensureUploadRootExists()
	return prod
}

func (f productionFileService) CreateUploadFileForUser(u *User, ext string) (fi *os.File, fn string, e error) {
	filename, _ := genUUID()
	filename += ext

	filePath := path.Join(f.apiUploadRoot, filename)

	file, err := os.Create(filePath)
	return file, filename, err
}

func (f productionFileService) ensureUploadRootExists() error {
	if !exists(f.uploadRoot) {
		return fmt.Errorf("upload root doesn't exist:%v", f.uploadRoot)
	}
	if !exists(f.apiUploadRoot) {
		err := os.Mkdir(f.apiUploadRoot, 0777)
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func genUUID() (string, error) {

	uuid := make([]byte, 16)

	n, err := rand.Read(uuid)

	if n != len(uuid) || err != nil {
		return "", err
	}

	// TODO: verify the two lines implement RFC 4122 correctly

	uuid[8] = 0x80 // variant bits see page 5

	uuid[4] = 0x40 // version 4 Pseudo Random, see page 7

	return hex.EncodeToString(uuid), nil
}

func exists(path string) bool {
	log.Printf("exists:%v\n", path)
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	} else if err != nil {
		log.Printf(err.Error())
		log.Fatal(err)
	}
	return true
}
