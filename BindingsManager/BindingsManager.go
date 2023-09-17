package bindingsmanager

import (
	metaDataM "eboox/MetaDataManager"
	userActivityM "eboox/UserActivityManager"
	utils "eboox/Utils"
	"log"

	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Endpoint string

type BindingsManager struct {
	LibraryMetaDataPTR *map[metaDataM.BookUuid]metaDataM.BookMetaData
	BooksActivitiesPTR *map[metaDataM.BookUuid]*userActivityM.BookActivity
	CurrentBookFiles   metaDataM.BookFiles
	CurrentBookUuid    string
	PreviousBookUuid   string

	OutChan       chan string
	BookFilesChan chan metaDataM.BookFiles
}

func BindingsManagerInit(
	metaDataPTR *map[metaDataM.BookUuid]metaDataM.BookMetaData,
	userActivityDir string,
	outChan chan string,
	bookFileChan chan metaDataM.BookFiles,
	bookActivitiesPTR *map[metaDataM.BookUuid]*userActivityM.BookActivity,
) *BindingsManager {
	return &BindingsManager{
		LibraryMetaDataPTR: metaDataPTR,
		OutChan:            outChan,
		BookFilesChan:      bookFileChan,
		BooksActivitiesPTR: bookActivitiesPTR,
	}
}

// FIXME : check for existance before handling

func (sm *BindingsManager) HandleLibraryMetaData() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(utils.MapToSlice(sm.LibraryMetaDataPTR, sm.BooksActivitiesPTR))
	}
}

func (sm *BindingsManager) HandleBookMetaData() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		bookUuid := req.URL.Query().Get("book_uuid")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if book, ok := (*sm.LibraryMetaDataPTR)[metaDataM.BookUuid(bookUuid)]; ok {

			bookAct := (*sm.BooksActivitiesPTR)[metaDataM.BookUuid(bookUuid)]
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"metadata": book, "activity": bookAct})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}
}

func (sm *BindingsManager) HandleBook() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		bookUuid := req.URL.Query().Get("book_uuid")
		sm.PreviousBookUuid = sm.CurrentBookUuid
		sm.CurrentBookUuid = bookUuid

		(*sm.BooksActivitiesPTR)[metaDataM.BookUuid(bookUuid)].LastOpened = utils.GetToday()

		var (
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

func (sm *BindingsManager) HandleCover() func(http.ResponseWriter, *http.Request) {
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

func (sm *BindingsManager) HandleBookFiles(w http.ResponseWriter, req *http.Request) {
	if sm.CurrentBookUuid != sm.PreviousBookUuid {
		sm.PreviousBookUuid = sm.CurrentBookUuid
		sm.OutChan <- sm.CurrentBookUuid
		sm.CurrentBookFiles = <-sm.BookFilesChan
	}
	log.Printf("reqested [%s] for <%s>", req.URL.Path, sm.CurrentBookUuid)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(sm.CurrentBookFiles[req.URL.Path[1:]])
}

func (sm *BindingsManager) HandlersInit() {
	http.HandleFunc("/librarymetadata", sm.HandleLibraryMetaData())
	http.HandleFunc("/bookmetadata", sm.HandleBookMetaData())
	http.HandleFunc("/book", sm.HandleBook())
	http.HandleFunc("/cover", sm.HandleCover())

	// Activities Handlers
	http.HandleFunc("/progress", sm.HandleProgress)
	http.HandleFunc("/bookmark", sm.HandleBookmarks)
	http.HandleFunc("/highlight", sm.HandleHighlights)
	http.HandleFunc("/note", sm.HandleNotes)
	http.Handle("/", http.HandlerFunc(sm.HandleBookFiles))
}

// Activities handlers
func (sm *BindingsManager) HandleProgress(w http.ResponseWriter, req *http.Request) {
	var requestData map[string]string
	body, _ := ioutil.ReadAll(req.Body)
	if err := json.Unmarshal(body, &requestData); err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	bookUuid := requestData["book_uuid"]
	(*sm.BooksActivitiesPTR)[metaDataM.BookUuid(bookUuid)].BookProgress = requestData["progress"]
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Updated Book Progress!"))
}

func (sm *BindingsManager) HandleHighlights(w http.ResponseWriter, req *http.Request) {

	var requestData map[string]string
	body, _ := ioutil.ReadAll(req.Body)
	if err := json.Unmarshal(body, &requestData); err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	var (
		bookUuid = requestData["book_uuid"]
		action   = requestData["action"]
		bounds   = requestData["bounds"]
		content  = requestData["content"]
	)

	currentBook := (*sm.BooksActivitiesPTR)[metaDataM.BookUuid(bookUuid)]

	switch action {
	case "add":
		for _, other := range currentBook.Highlights {
			if other.Bounds == bounds {
				w.WriteHeader(http.StatusMethodNotAllowed)
				w.Write([]byte("Highlight already Exists."))
				return
			}
		}
		currentBook.Highlights = append(currentBook.Highlights, &userActivityM.Highlight{Content: content, Bounds: bounds, Note: "", Date: utils.GetToday()})
		break
	case "remove":
		for i, other := range currentBook.Highlights {
			if other.Bounds == bounds {
				currentBook.Highlights = append(currentBook.Highlights[:i], currentBook.Highlights[i+1:]...)
				break
			}
		}
		break
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Action Not Allowed on highlights."))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Updated Book Highlights!"))

}

func (sm *BindingsManager) HandleNotes(w http.ResponseWriter, req *http.Request) {
	var requestData map[string]string
	body, _ := ioutil.ReadAll(req.Body)
	if err := json.Unmarshal(body, &requestData); err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	var (
		bookUuid = requestData["book_uuid"]
		action   = requestData["action"]
		bounds   = requestData["highlight"]
		content  = requestData["content"]
	)

	currentBook := (*sm.BooksActivitiesPTR)[metaDataM.BookUuid(bookUuid)]

	switch action {
	case "add", "update":
		for _, highlight := range currentBook.Highlights {
			if highlight.Bounds == bounds {
				highlight.Note = content
				break
			}
		}
		break
	case "remove":
		for _, highlight := range currentBook.Highlights {
			if highlight.Bounds == bounds {
				highlight.Note = ""
				break
			}
		}
		break
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Action Not Allowed on notes."))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Updated Book Notes!"))

}

func (sm *BindingsManager) HandleBookmarks(w http.ResponseWriter, req *http.Request) {

	var requestData map[string]string
	body, _ := ioutil.ReadAll(req.Body)
	if err := json.Unmarshal(body, &requestData); err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	var (
		bookUuid = requestData["book_uuid"]
		action   = requestData["action"]
		bookmark = requestData["bookmark"]
	)

	currentBook := (*sm.BooksActivitiesPTR)[metaDataM.BookUuid(bookUuid)]

	switch action {
	case "add":
		currentBook.Bookmarks = append(currentBook.Bookmarks, bookmark)
		break
	case "remove":
		for i, other := range currentBook.Bookmarks {
			if other == bookmark {
				currentBook.Bookmarks = append(currentBook.Bookmarks[:i], currentBook.Bookmarks[i+1:]...)
				break
			}
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Action Not Allowed on bookmarks."))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Updated Book Bookmarks!"))
}

// TODO : DONT GORGET
func (sm *BindingsManager) HandleBookDate() {}
