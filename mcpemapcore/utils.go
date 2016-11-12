package mcpemapcore

import (
	"io"
	"log"
	"os"
)

func copyFile(source, destination string) error {
	var err error
	file, err := os.Open(source)
	if err != nil {
		log.Printf("Couldn't open %v\n", source)
		return err
	}
	defer file.Close()
	output, err := os.Create(destination)
	defer output.Close()
	if err != nil {
		log.Printf("Couldn't create %v\n", destination)
		return err
	}
	_, err = io.Copy(output, file)
	return err
}
