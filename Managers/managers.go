package managers

import (
	bindingsManager "eboox/BindingsManager"
	metadatamanager "eboox/MetaDataManager"
	"eboox/UserActivityManager"
	"eboox/UserConfManager"
	utils "eboox/Utils"
)

type BooksManagers struct {
	UserActivityM *useractivitymanager.UserActivityManager
	MetaDataM     *metadatamanager.MetaDataManager
	BindingsM     *bindingsManager.BindingsManager
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
		BindingsM:     bindingsManager.BindingsManagerInit(&metaDataManager.BooksMetaData, userConfManager.UserActivityDir, bookUuidBroadCast, bookFilesChan, userActivityManager.BooksActivity),
		UserConfM:     userConfManager,
		UserActivityM: userActivityManager,
		MetaDataM:     metaDataManager,
	}

	go bm.BindingsM.HandlersInit()

	return bm
}

// XXX - DEV
func (m *BooksManagers) Save() {
	m.UserActivityM.Save()
	utils.Assert(false, "[BooksManagers.Save] NOT Fully IMPLEMENTED.")
}
