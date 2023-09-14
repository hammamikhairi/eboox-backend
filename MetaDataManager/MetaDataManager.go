package metadatamanager

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"log"
	"strings"

	utils "eboox/Utils"
)

type MetaDataManager struct {
	BooksMetaData map[BookUuid]BookMetaData

	inChan  chan string
	OutChan chan BookFiles
}

type BookUuid string
type BookFiles map[string][]byte
type BookMetaData struct {
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`

	Title    string `json:"title" xml:"metadata>title"`
	Author   string `json:"author" xml:"metadata>creator"`
	Language string `json:"language" xml:"metadata>language"`
	Date     string `json:"creation_date" xml:"metadata>date"`
	Uuid     string `json:"book_uuid" xml:"metadata>identifier"`

	Cover []byte `json:"-"`
}

func MetaDataManagerInit(path string, inChan chan string, outChan chan BookFiles) *MetaDataManager {

	return &MetaDataManager{
		BooksMetaData: LoadBooksMetaData(path),
		inChan:        inChan,
		OutChan:       outChan,
	}
}

func (mdm *MetaDataManager) HandleChans() {
	for {
		bookUuid := <-mdm.inChan
		println("meta")
		book := mdm.BooksMetaData[BookUuid(bookUuid)]
		fullPath := book.FilePath + book.FileName

		bookFiles := BookFiles{}

		r, err := zip.OpenReader(fullPath)
		if err != nil {
			log.Fatalf("Error opening EPUB file <%s> : %s", fullPath, err)
		}

		for _, f := range r.File {
			rc, _ := f.Open()
			var buffer bytes.Buffer
			_, err = io.Copy(&buffer, rc)
			if err != nil {
				panic(err)
			}
			bookFiles[f.Name] = buffer.Bytes()
			rc.Close()
		}
		mdm.OutChan <- bookFiles
		r.Close()
	}
}

func LoadBooksMetaData(booksPath string) map[BookUuid]BookMetaData {
	booksMap := make(map[BookUuid]BookMetaData)

	books := utils.GetFiles(booksPath)

	// TODO : Concurrent loading
	for _, book := range books {
		bookMetaData, bookUuid := LoadBookMetaData(booksPath, book)
		booksMap[bookUuid] = bookMetaData
	}

	return booksMap
}

func LoadBookMetaData(path, name string) (book BookMetaData, bookUuid BookUuid) {
	fullPath := path + name
	r, err := zip.OpenReader(fullPath)
	if err != nil {
		log.Fatalf("Error opening EPUB file <%s> : %s", fullPath, err)
	}
	defer r.Close()

	LoadFromArchive(&book, r, FileInfo{path, name})
	return book, BookUuid(book.Uuid)
}

type FileInfo [2]string

func LoadFromArchive(book *BookMetaData, archive *zip.ReadCloser, fileInfo FileInfo) {
	var CoverBytes []byte
	for _, f := range archive.File {
		// Optimization
		if len(CoverBytes) != 0 && book.Title != "" {
			break
		}

		rc, err := f.Open()
		if err != nil {
			log.Fatalf("Malformatted EPUB File <%s> : %s", fileInfo[1], err)
		}
		defer rc.Close()

		var buffer bytes.Buffer
		if strings.Contains(f.Name, "cover") {
			_, err = io.Copy(&buffer, rc)
			if err != nil {
				panic(err)
			}
			CoverBytes = buffer.Bytes()
		}
		if strings.Contains(f.Name, "content.opf") {
			_, err = io.Copy(&buffer, rc)
			if err != nil {
				panic(err)
			}
			err = xml.Unmarshal(buffer.Bytes(), &book)
			if err != nil {
				panic(err)
			}
		}
	}

	book.Date = utils.FormatDate(book.Date)
	book.FilePath = fileInfo[0]
	book.FileName = fileInfo[1]
	book.Cover = CoverBytes
}
