package session

import (
	"encoding/gob"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/r-erema/wapi/internal/model/session"
)

const sessionFileExt = ".gob"

// Stores sessions metadata in filesystem.
type FileSystemSession struct {
	sessionStoragePath string
}

// Creates File System Repository.
func NewFileSystem(sessionStoragePath string) (*FileSystemSession, error) {
	if _, err := os.Stat(sessionStoragePath); os.IsNotExist(err) {
		err := os.MkdirAll(sessionStoragePath, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}
	return &FileSystemSession{sessionStoragePath: sessionStoragePath}, nil
}

// Retrieves session from repository.
func (f FileSystemSession) ReadSession(sessionID string) (*session.WapiSession, error) {
	ws := &session.WapiSession{}
	file, err := os.Open(f.resolveSessionFilePath(sessionID))
	if err != nil {
		return nil, err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			log.Println(err)
		}
	}()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&ws)
	if err != nil {
		return nil, err
	}
	return ws, nil
}

// Retrieves session from repository.
func (f FileSystemSession) WriteSession(s *session.WapiSession) error {
	file, err := os.Create(f.resolveSessionFilePath(s.SessionID))
	if err != nil {
		return err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			log.Println(err)
		}
	}()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(s)
	if err != nil {
		return err
	}
	return nil
}

// Retrieves all sessions ids from repository.
func (f FileSystemSession) AllSavedSessionIds() ([]string, error) {
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

// Removes session from repository.
func (f FileSystemSession) RemoveSession(sessionID string) error {
	if err := os.Remove(f.resolveSessionFilePath(sessionID)); err != nil {
		return err
	}
	return nil
}

func (f FileSystemSession) resolveSessionFilePath(sessionID string) string {
	return f.sessionStoragePath + "/" + sessionID + sessionFileExt
}
