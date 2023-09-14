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

	BookProgress string
	Bookmarks    []string    `json:"bookmarks"`
	Highlights   []Highlight `json:"highlights"`
}

type Highlight struct {
	Content string
	Bounds  []string
	Note    string
}

type BooksActivity map[metadatamanager.BookUuid]BookActivity

type UserActivityManager struct {
	LibraryMetaDataPTR *map[metadatamanager.BookUuid]metadatamanager.BookMetaData
	BooksActivity      *BooksActivity

	UserActivityDir string

	inChan  chan string
	outChan chan BookActivity
}

func UserActivityManagerInit(libraryMetaDataPTR *map[metadatamanager.BookUuid]metadatamanager.BookMetaData,
	userActivityDir string,
	inChan chan string,
	outChan chan BookActivity,
) *UserActivityManager {

	return &UserActivityManager{
		LibraryMetaDataPTR: libraryMetaDataPTR,
		UserActivityDir:    userActivityDir,
		inChan:             inChan,
		outChan:            outChan,
		BooksActivity:      LoadBooksActivities(userActivityDir, libraryMetaDataPTR),
	}
}

func LoadBooksActivities(
	actPath string,
	libraryMetaDataPTR *map[metadatamanager.BookUuid]metadatamanager.BookMetaData,
) *BooksActivity {
	booksActivities := BooksActivity{}
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
				booksActivities[bookUuid] = BookActivity{
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
			booksActivities[bookUuid] = BookActivity{
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
