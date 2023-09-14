package useractivitymanager

import (
	metadatamanager "eboox/MetaDataManager"
)

type BookActivity struct {
	BookProgress string
	Bookmarks    []string    `json:"bookmarks"`
	Highlights   []Highlight `json:"highlights"`
}

type Highlight struct {
	Content string
	Bounds  []string
	Note    string
}

type UserActivityManager struct {
	LibraryMetaDataPTR *map[metadatamanager.BookUuid]metadatamanager.BookMetaData
	BooksActivity      map[metadatamanager.BookUuid]BookActivity

	UserActivityDir string

	inChan  chan string
	outChan chan BookActivity
}

func UserActivityManagerInit(libraryMetaDataPTR *map[metadatamanager.BookUuid]metadatamanager.BookMetaData, userActivityDir string, inChan chan string, outChan chan BookActivity) *UserActivityManager {
	// check
	return &UserActivityManager{
		LibraryMetaDataPTR: libraryMetaDataPTR,
		UserActivityDir:    userActivityDir,
		inChan:             inChan,
		outChan:            outChan,
	}
}

func (uam *UserActivityManager) HandleChans() {
	for {
		bookUuid := <-uam.inChan
		println("act")
		uam.outChan <- BookActivity{
			BookProgress: bookUuid,
		}
	}
}
