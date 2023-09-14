package useractivitymanager

import (
	metadatamanager "eboox/MetaDataManager"
)

type BookActivity struct {
	Progress   string
	BookMarks  []string
	highlights []Highlight
}

type Highlight struct {
}

type UserActivityManager struct {
	LibraryMetaDataPTR *map[metadatamanager.BookUuid]metadatamanager.BookMetaData
	UserActivity       map[metadatamanager.BookUuid]BookActivity

	UserActivityDir string
}

func UserActivityManagerInit(metadataptr *map[metadatamanager.BookUuid]metadatamanager.BookMetaData, userActivityDir string) *UserActivityManager {
	return &UserActivityManager{
		LibraryMetaDataPTR: metadataptr,
		UserActivityDir:    userActivityDir,
	}
}
