package useractivitymanager

import (
	metadatamanager "eboox/MetaDataManager"
	utils "eboox/Utils"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type BookActivity struct {
	// XXX - DEV
	BookUuid string

	LastOpened   string       `json:"last_opened"`
	LastPage     string       `json:"last_page_opened"`
	BookProgress string       `json:"book_progress"`
	Bookmarks    []string     `json:"bookmarks"`
	Highlights   []*Highlight `json:"highlights"`
}

type Highlight struct {
	Date    string `json:"date"`
	Content string `json:"content"`
	Bounds  string `json:"bounds"`
	Note    string `json:"note"`
}

type UserActivityManager struct {
	LibraryMetaDataPTR *map[metadatamanager.BookUuid]metadatamanager.BookMetaData
	BooksActivity      *map[metadatamanager.BookUuid]*BookActivity

	UserActivityDir string
}

func (ba *UserActivityManager) UpdateProgress(bookUuid, newProg string) {
	// ba.Activities[Bo]
}

func UserActivityManagerInit(libraryMetaDataPTR *map[metadatamanager.BookUuid]metadatamanager.BookMetaData,
	userActivityDir string,
) *UserActivityManager {

	return &UserActivityManager{
		LibraryMetaDataPTR: libraryMetaDataPTR,
		UserActivityDir:    userActivityDir,
		BooksActivity:      LoadBooksActivities(userActivityDir, libraryMetaDataPTR),
	}
}

func LoadBooksActivities(
	actPath string,
	libraryMetaDataPTR *map[metadatamanager.BookUuid]metadatamanager.BookMetaData,
) *map[metadatamanager.BookUuid]*BookActivity {
	booksActivities := map[metadatamanager.BookUuid]*BookActivity{}
	if utils.ActivitiesExists(actPath) {
		// load
		actData, err := os.ReadFile(actPath)
		utils.Assert(err == nil, "Assersion failure : Cannot load data from confFile.")
		json.Unmarshal(actData, &booksActivities)
		fmt.Printf("Loaded Books Records successfully from [%s]\n", actPath)

		// TODO check
		var initialCount int = len(booksActivities)
		for bookUuid := range *libraryMetaDataPTR {
			if _, ok := booksActivities[bookUuid]; !ok {
				// XXX - might break studff, idk
				log.Printf("Loaded New Book <%s>", bookUuid)
				booksActivities[bookUuid] = &BookActivity{
					BookUuid: string(bookUuid),
				}
			}
		}

		if initialCount != len(booksActivities) {
			// save new
			actData, _ := json.MarshalIndent(booksActivities, "", "  ")
			err := os.WriteFile(actPath, actData, 0644)
			utils.Assert(err == nil, "Assersion failure : Cannot write conf to file.")
		}

	} else {
		// create
		utils.MakeActivity(actPath)
		for bookUuid := range *libraryMetaDataPTR {
			booksActivities[bookUuid] = &BookActivity{
				BookUuid: string(bookUuid),
			}
		}
		actData, _ := json.MarshalIndent(booksActivities, "", "  ")
		err := os.WriteFile(actPath, actData, 0644)
		utils.Assert(err == nil, "Assersion failure : Cannot write conf to Conffile.")
		fmt.Printf("Initialized Books records at [%s]\n", actPath)
	}
	return &booksActivities
}

func (acm *UserActivityManager) Save() {
	actData, _ := json.MarshalIndent(acm.BooksActivity, "", "  ")
	err := os.WriteFile(acm.UserActivityDir, actData, 0644)
	utils.Assert(err == nil, "Assersion failure : Cannot write conf to file.")
	fmt.Printf("Saved Books records at [%s]\n", acm.UserActivityDir)
}
