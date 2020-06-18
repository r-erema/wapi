package SessionWorks

import (
	"encoding/gob"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/r-erema/wapi/src/Session"
)

const sessionFileExt = ".gob"

type fileSystemSession struct {
	sessionStoragePath string
}

func NewFileSystemSession(sessionStoragePath string) (*fileSystemSession, error) {
	if _, err := os.Stat(sessionStoragePath); os.IsNotExist(err) {
		err := os.MkdirAll(sessionStoragePath, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}
	return &fileSystemSession{sessionStoragePath: sessionStoragePath}, nil
}

func (f fileSystemSession) ReadSession(sessionId string) (*Session.WapiSession, error) {
	session := &Session.WapiSession{}
	file, err := os.Open(f.resolveSessionFilePath(sessionId))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&session)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (f fileSystemSession) WriteSession(session *Session.WapiSession) error {
	file, err := os.Create(f.resolveSessionFilePath(session.SessionId))
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(session)
	if err != nil {
		return err
	}
	return nil
}

func (f fileSystemSession) GetAllSavedSessionIds() ([]string, error) {
	var ids []string
	files, err := ioutil.ReadDir(f.sessionStoragePath)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		ext := path.Ext(file.Name())
		if sessionFileExt == ext {
			id := strings.TrimSuffix(file.Name(), path.Ext(file.Name()))
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (f fileSystemSession) RemoveSession(sessionId string) error {
	if err := os.Remove(f.resolveSessionFilePath(sessionId)); err != nil {
		return err
	}
	return nil
}

func (f fileSystemSession) resolveSessionFilePath(sessionId string) string {
	return f.sessionStoragePath + "/" + sessionId + sessionFileExt
}
