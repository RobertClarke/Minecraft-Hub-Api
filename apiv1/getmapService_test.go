package main

import (
	"log"
	"testing"
)

func TestGetAllMaps(t *testing.T) {
	service := CreateGetMapService()
	maps, _, err := service.GetAllMaps(0, 20, "minecrafthub.com")

	if err != nil {
		t.Error(err)
	}

	if len(maps) == 0 {
		t.Fatal("no maps found")
	}
}

func TestMapFields(t *testing.T) {
	log.Printf("TestMapFields\n")
	service := CreateGetMapService()
	maps, _, err := service.GetAllMaps(0, 20, "minecrafthub.com")

	if err != nil {
		t.Error(err)
	}

	for _, r := range maps {
		log.Printf("id %v download uri %v imageuri %v", r.Id, r.MapDownloadUri, r.MapImageUriList[0].MapImageUri)
	}
}
