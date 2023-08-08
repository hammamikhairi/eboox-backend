package main

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	// "io"
	// "io/ioutil"
	// "bytes"
	// "path/filepath"
	// "strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run get_cover_photo.go path/to/epub-file.epub")
		return
	}

	filePath := os.Args[1]

	// Open the EPUB file
	r, err := zip.OpenReader(filePath)
	if err != nil {
		log.Fatalf("Error opening EPUB file: %s", err)
	}
	defer r.Close()

	for _, f := range r.File {
		// fmt.Printf("%+v\n", f)
		if strings.Contains(f.Name, "stylesheet.css") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			defer rc.Close()

			buffer, _ := ioutil.ReadAll(rc)
			fmt.Println(string(buffer))
		}
	}
	// return nil, fmt.Errorf("cover image not found in EPUB file")
}
