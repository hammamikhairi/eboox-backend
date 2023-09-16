package managers

import (
	metadatamanager "eboox/MetaDataManager"
	serverManager "eboox/ServerManager"
	"eboox/UserActivityManager"
	"eboox/UserConfManager"
	utils "eboox/Utils"
)

type BooksManagers struct {
	UserActivityM *useractivitymanager.UserActivityManager
	MetaDataM     *metadatamanager.MetaDataManager
	ServerM       *serverManager.ServerManager
	UserConfM     *userconfmanager.UserConfManager
}

const BROAD_CAST_COUNT int = 1

func ManagersInit(libraryPath string) BooksManagers {

	var (
		bookUuidBroadCast chan string                    = make(chan string, 1)
		bookUuidChan      chan string                    = make(chan string, 2)
		bookFilesChan     chan metadatamanager.BookFiles = make(chan metadatamanager.BookFiles, 1)
	)
	metaDataManager := metadatamanager.MetaDataManagerInit(libraryPath, bookUuidChan, bookFilesChan)
	userConfManager := userconfmanager.UserConfManagerInit()
	userActivityManager := useractivitymanager.UserActivityManagerInit(&metaDataManager.BooksMetaData, userConfManager.UserActivityDir)

	go func() {
		var request string
		for {
			select {
			case request = <-bookUuidBroadCast:
				for i := 0; i < BROAD_CAST_COUNT; i++ {
					bookUuidChan <- request
				}
			}

		}
	}()

	go metaDataManager.HandleChans()

	bm := BooksManagers{
		ServerM:       serverManager.ServerManagerInit(&metaDataManager.BooksMetaData, userConfManager.UserActivityDir, bookUuidBroadCast, bookFilesChan, userActivityManager.BooksActivity),
		UserConfM:     userConfManager,
		UserActivityM: userActivityManager,
		MetaDataM:     metaDataManager,
	}

	go bm.ServerM.HandlersInit()

	return bm
}

func (m *BooksManagers) Save() {
	utils.Assert(false, "[BooksManagers.Save] NOT IMPLEMENTED.")
}
