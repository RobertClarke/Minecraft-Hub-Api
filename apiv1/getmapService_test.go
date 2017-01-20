package main

import (
	"log"
	"testing"
)

func TestGetAllMaps(t *testing.T) {
	service := CreateGetMapService()
	maps, _, err := service.GetAllMapsQuery(0, 20, "minecrafthub.com", "all")

	if err != nil {
		t.Error(err)
	}

	if len(maps) == 0 {
		t.Fatal("no maps found")
	}
}

func TestGetFeatuerdMaps(t *testing.T) {
	service := CreateGetMapService()
	maps, _, err := service.GetAllMapsQuery(0, 20, "minecrafthub.com", "featured")

	if err != nil {
		t.Error(err)
	}

	if len(maps) == 0 {
		t.Fatal("no maps found")
	}
}

func TestGetSurvivalMaps(t *testing.T) {
	service := CreateGetMapService()
	maps, _, err := service.GetAllMapsQuery(0, 20, "minecrafthub.com", "c:6")

	if err != nil {
		t.Error(err)
	}

	if len(maps) == 0 {
		t.Fatal("no maps found")
	}
}

func TestGetCreativeMaps(t *testing.T) {
	service := CreateGetMapService()
	maps, _, err := service.GetAllMapsQuery(0, 20, "minecrafthub.com", "c:7")

	if err != nil {
		t.Error(err)
	}

	if len(maps) == 0 {
		t.Fatal("no maps found")
	}
}

func TestGetMostDownloadedMaps(t *testing.T) {
	service := CreateGetMapService()
	maps, _, err := service.GetAllMapsQuery(0, 20, "minecrafthub.com", "mostdownloaded")

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
	maps, _, err := service.GetAllMapsQuery(0, 1, "minecrafthub.com", "all")

	if err != nil {
		t.Error(err)
	}

	for _, r := range maps {
		log.Printf("id %v download uri %v imageuri %v", r.ID, r.MapDownloadURI, r.MapImageURIList[0].MapImageURI)
	}
}
