package metadatamanager

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"strings"

	utils "eboox/Utils"
)

type MetaDataManager struct {
	BooksMetaData map[BookUuid]BookMetaData
	LibraryPath   string

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

	Cover     []byte                `json:"-"`
	BookSpine map[string]*SpineMeta `json:"-"`
	Size      int
}

type SpineMeta struct {
	ChunkSize        int
	RuningPercentage float32
	ChapterIndex     int
}

func MetaDataManagerInit(path string, inChan chan string, outChan chan BookFiles) *MetaDataManager {

	return &MetaDataManager{
		BooksMetaData: LoadBooksMetaData(path),
		LibraryPath:   path,
		inChan:        inChan,
		OutChan:       outChan,
	}
}

func (mdm *MetaDataManager) HandleChans() {
	for {
		bookUuid := <-mdm.inChan
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

	fmt.Printf("Loaded Library Books successfully from [%s]\n", booksPath)

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

	manifest := struct {
		Items []struct {
			Href      string `xml:"href,attr"`
			ID        string `xml:"id,attr"`
			MediaType string `xml:"media-type,attr"`
		} `xml:"manifest>item"`
	}{}
	spine := struct {
		ItemRefs []struct {
			IDRef string `xml:"idref,attr"`
		} `xml:"spine>itemref"`
	}{}

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
			err = xml.Unmarshal(buffer.Bytes(), &manifest)
			if err != nil {
				panic(err)
			}
			err = xml.Unmarshal(buffer.Bytes(), &spine)
			if err != nil {
				panic(err)
			}

		}
	}

	// For tracking progress (SOMEHOW THE DEVs OF EPUB.JS COULD NOT TRACK THE PROGRESS OF READING? Thanks for the library anyway <3)
	itemPositionMap := make(map[string]*SpineMeta)
	utils.Pp(itemPositionMap)
	booksSizeByIndex := make([]int, len(spine.ItemRefs))
	for i, itemRef := range spine.ItemRefs {
		for _, item := range manifest.Items {
			if item.ID == itemRef.IDRef {
				itemPositionMap[item.Href] = &SpineMeta{ChapterIndex: i}
				break
			}
		}
	}

	total := 0
	for _, f := range archive.File {
		rc, err := f.Open()
		defer rc.Close()

		var buffer bytes.Buffer
		if _, ok := itemPositionMap[f.Name]; ok {
			_, err = io.Copy(&buffer, rc)
			if err != nil {
				panic(err)
			}

			fileLength := len(buffer.Bytes())
			total += fileLength
			booksSizeByIndex[itemPositionMap[f.Name].ChapterIndex] = len(buffer.Bytes())
			itemPositionMap[f.Name].ChunkSize = fileLength
		}
	}

	for file, spineMeta := range itemPositionMap {
		temp := 0
		for i := 0; i < spineMeta.ChapterIndex; i++ {
			temp += booksSizeByIndex[i]
		}

		itemPositionMap[file].RuningPercentage = (float32(temp) / float32(total)) * 100
	}

	println(book.Title)
	println(total)

	book.BookSpine = itemPositionMap
	book.Size = total
	book.Date = utils.FormatDate(book.Date)
	book.FilePath = fileInfo[0]
	book.FileName = fileInfo[1]
	book.Cover = CoverBytes
}
