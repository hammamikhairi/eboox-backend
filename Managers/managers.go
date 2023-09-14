package managers

import (
	metadatamanager "eboox/MetaDataManager"
	serverManager "eboox/ServerManager"
	userActivityManager "eboox/UserActivityManager"
	"eboox/UserConfManager"
	utils "eboox/Utils"
)

type BooksManagers struct {
	UserActivityM *userActivityManager.UserActivityManager
	MetaDataM     *metadatamanager.MetaDataManager
	ServerM       *serverManager.ServerManager
	UserConfM     *userconfmanager.UserConfManager
}

func ManagersInit(libraryPath string) BooksManagers {

	var (
		bookUuidChan  chan string                    = make(chan string, 1)
		bookFilesChan chan metadatamanager.BookFiles = make(chan metadatamanager.BookFiles, 1)
	)
	metaDataManager := metadatamanager.MetaDataManagerInit(libraryPath, bookUuidChan, bookFilesChan)
	userConfManager := userconfmanager.UserConfManagerInit()

	go metaDataManager.HandleChans()

	bm := BooksManagers{
		ServerM:       serverManager.ServerManagerInit(&metaDataManager.BooksMetaData, userConfManager.UserActivityDir, bookUuidChan, bookFilesChan),
		UserActivityM: userActivityManager.UserActivityManagerInit(&metaDataManager.BooksMetaData, userConfManager.UserActivityDir),
		UserConfM:     userConfManager,
		MetaDataM:     metaDataManager,
	}

	go bm.ServerM.HandlersInit()

	return bm
}

func (m *BooksManagers) Save() {
	utils.Assert(false, "[BooksManagers.Save] NOT IMPLEMENTED.")
}
