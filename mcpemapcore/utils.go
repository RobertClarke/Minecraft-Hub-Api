package mcpemapcore

import (
	"io"
	"os"
)

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
