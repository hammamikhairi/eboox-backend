package servermanager

import (
	metaDataM "eboox/MetaDataManager"
	userActivityM "eboox/UserActivityManager"
	utils "eboox/Utils"

	"encoding/json"
	"net/http"
)

type Endpoint string

type ServerManager struct {
	LibraryMetaDataPTR *map[metaDataM.BookUuid]metaDataM.BookMetaData
	BooksActivitiesPTR *userActivityM.BooksActivity

	CurrentBookFiles metaDataM.BookFiles
	CurrentBookUuid  string
	PreviousBookUuid string

	OutChan       chan string
	BookFilesChan chan metaDataM.BookFiles
}

func ServerManagerInit(
	metaDataPTR *map[metaDataM.BookUuid]metaDataM.BookMetaData,
	userActivityDir string,
	outChan chan string,
	bookFileChan chan metaDataM.BookFiles,
	bookActivitiesPTR *userActivityM.BooksActivity,
) *ServerManager {
	return &ServerManager{
		LibraryMetaDataPTR: metaDataPTR,
		OutChan:            outChan,
		BookFilesChan:      bookFileChan,
		BooksActivitiesPTR: bookActivitiesPTR,
	}
}

func (sm *ServerManager) HandleLibraryMetaData() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(utils.MapToSlice(sm.LibraryMetaDataPTR))
	}
}

func (sm *ServerManager) HandleBookMetaData() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		bookUuid := req.URL.Query().Get("book_uuid")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if book, ok := (*sm.LibraryMetaDataPTR)[metaDataM.BookUuid(bookUuid)]; ok {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(book)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}
}

func (sm *ServerManager) HandleBook() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		bookUuid := req.URL.Query().Get("book_uuid")
		sm.PreviousBookUuid = sm.CurrentBookUuid
		sm.CurrentBookUuid = bookUuid

		var (
			// book metaDataM.BookMetaData
			ok bool
		)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if _, ok = (*sm.LibraryMetaDataPTR)[metaDataM.BookUuid(bookUuid)]; !ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.Write([]byte("Book Found."))
		}
	}
}

func (sm *ServerManager) HandleCover() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		bookUuid := req.URL.Query().Get("book_uuid")
		var (
			book metaDataM.BookMetaData
			ok   bool
		)
		if book, ok = (*sm.LibraryMetaDataPTR)[metaDataM.BookUuid(bookUuid)]; !ok {
			w.WriteHeader(http.StatusNotFound)
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(book.Cover)
	}
}

func (sm *ServerManager) HandlersInit() {
	http.HandleFunc("/librarymetadata", sm.HandleLibraryMetaData())
	http.HandleFunc("/bookmetadata", sm.HandleBookMetaData())
	http.HandleFunc("/booki", sm.HandleBook())
	http.HandleFunc("/cover", sm.HandleCover())
	http.HandleFunc("/activity", sm.HandleBookActivity)
	// http.HandleFunc("/META-INF/container.xml", sm.TEMPContainerHandler())
	http.Handle("/", http.HandlerFunc(sm.HandleBookFiles))
}

func (sm *ServerManager) HandleBookActivity(w http.ResponseWriter, req *http.Request) {

	bookUuid := req.URL.Query().Get("book_uuid")
	book := (*sm.BooksActivitiesPTR)[metaDataM.BookUuid(bookUuid)]
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(book)
}
func (sm *ServerManager) HandleBookFiles(w http.ResponseWriter, req *http.Request) {
	if sm.CurrentBookUuid != sm.PreviousBookUuid {
		sm.PreviousBookUuid = sm.CurrentBookUuid
		sm.OutChan <- sm.CurrentBookUuid
		sm.CurrentBookFiles = <-sm.BookFilesChan
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(sm.CurrentBookFiles[req.URL.Path[1:]])
}
