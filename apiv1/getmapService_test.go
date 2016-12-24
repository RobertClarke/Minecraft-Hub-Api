package main

import "testing"

func TestGetAllMaps(t *testing.T) {
	service := NewGetMapService()
	maps, _, err := service.GetAllMaps(0, 20, "minecrafthub.com")

	if err != nil {
		t.Error(err)
	}

	if len(maps) == 0 {
		t.Fatal("no maps found")
	}
}
