package main

import (
	"clarkezone-vs-com/mcpemapcore"
	"encoding/json"
	"fmt"
	"log"
	"testing"
)

func TestGetMap(t *testing.T) {
	themap, err := mcpemapcore.GetMapFromRedis("1")
	if err != nil {
		log.Fatal(err)
	}
	bytes, err := json.Marshal(themap)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", bytes)
}
