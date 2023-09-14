package servermanager

import (
	metaDataM "eboox/MetaDataManager"
	userActivityM "eboox/UserActivityManager"
	utils "eboox/Utils"

	"encoding/json"
	"net/http"
	"os"
)

type Endpoint string

type ServerManager struct {
	LibraryMetaDataPTR *map[metaDataM.BookUuid]metaDataM.BookMetaData
	CurrentBookFiles   metaDataM.BookFiles
	CurrentBookUuid    string
	PreviousBook       string

	OutChan          chan string
	BookFileChan     chan metaDataM.BookFiles
	BookActivityChan chan userActivityM.BookActivity
}

func ServerManagerInit(metaDataPTR *map[metaDataM.BookUuid]metaDataM.BookMetaData, userActivityDir string, outChan chan string, bookFileChan chan metaDataM.BookFiles, bookActivityChan chan userActivityM.BookActivity) *ServerManager {
	return &ServerManager{
		LibraryMetaDataPTR: metaDataPTR,
		OutChan:            outChan,
		BookFileChan:       bookFileChan,
		BookActivityChan:   bookActivityChan,
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
		sm.PreviousBook = sm.CurrentBookUuid
		sm.CurrentBookUuid = bookUuid

		var (
			book metaDataM.BookMetaData
			ok   bool
		)
		if book, ok = (*sm.LibraryMetaDataPTR)[metaDataM.BookUuid(bookUuid)]; !ok {
			w.WriteHeader(http.StatusNotFound)
		}
		bookPath := book.FilePath + book.FileName

		file, err := os.Open(bookPath)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		defer file.Close()
		fileInfo, err := file.Stat()
		if err != nil {
			http.Error(w, "Unable to get file info", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Disposition", "attachment; filename="+fileInfo.Name())
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.ServeContent(w, req, fileInfo.Name(), fileInfo.ModTime(), file)
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
	http.HandleFunc("/book", sm.HandleBook())
	http.HandleFunc("/cover", sm.HandleCover())
	// http.HandleFunc("/META-INF/container.xml", sm.TEMPContainerHandler())
	http.Handle("/", http.HandlerFunc(sm.ServeAllHandler))
}

func (sm *ServerManager) ServeAllHandler(w http.ResponseWriter, req *http.Request) {
	if sm.CurrentBookUuid != sm.PreviousBook {
		sm.PreviousBook = sm.CurrentBookUuid
		// broadcast
		sm.OutChan <- sm.CurrentBookUuid

		// responses
		sm.CurrentBookFiles = <-sm.BookFileChan
		temp := <-sm.BookActivityChan
		utils.Pp(temp)
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(sm.CurrentBookFiles[req.URL.Path[1:]])
}
