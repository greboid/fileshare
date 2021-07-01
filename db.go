package fileshare

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tidwall/buntdb"
)

type DB struct {
	db      *buntdb.DB
	workdir string
}

func NewDB(workdir string, filename string) (*DB, error) {
	db, err := buntdb.Open(filename)
	if err != nil {
		return nil, err
	}
	return &DB{
		db:      db,
		workdir: workdir,
	}, nil
}

func (db *DB) AddEntry(ud UploadDescription) error {
	jsonData, err := json.Marshal(ud)
	if err != nil {
		return err
	}
	err = db.db.Update(func(tx *buntdb.Tx) error {
		_, _, err = tx.Set(ud.GetFullName(), string(jsonData), nil)
		return err
	})
	return err
}

func (db *DB) DeleteFile(filename string, folder string) error {
	ud, err := db.getFile(strings.TrimPrefix(filename, folder))
	if err != nil {
		return err
	}
	db.deleteUD(*ud, folder)
	return nil
}

func (db *DB) StartBackgroundPrune(folder string) {
	ticker := time.NewTicker(1 * time.Minute)
	db.pruneFiles(folder)
	go func() {
		for {
			select {
			case <-ticker.C:
				db.pruneFiles(folder)
			}
		}
	}()
}

func (db *DB) CheckFileName(filename string, folder string) {
	ud, err := db.getFile(strings.TrimPrefix(filename, folder))
	if err == nil {
		db.checkFile(*ud, folder)
	}
}

func (db *DB) checkFile(ud UploadDescription, folder string) {
	if !ud.Expiry.IsZero() && time.Now().After(ud.Expiry) {
		db.deleteUD(ud, folder)
	}
}

func (db *DB) deleteUD(ud UploadDescription, folder string) {
	log.Printf("Removing: %s", ud.GetFullName())
	err := os.Remove(filepath.Join(db.workdir, folder, ud.GetFullName()))
	if err != nil {
		log.Printf("Error removing file %s: %s", ud.GetFullName(), err.Error())
	}
	_ = db.db.Update(func(tx *buntdb.Tx) error {
		_, err = tx.Delete(ud.GetFullName())
		return err
	})
}

func (db *DB) GetFiles() []UploadDescription {
	var uploads []UploadDescription
	_ = db.db.View(func(tx *buntdb.Tx) error {
		err := tx.Ascend("", func(key, value string) bool {
			ud := UploadDescription{}
			_ = json.Unmarshal([]byte(value), &ud)
			uploads = append(uploads, ud)
			return true
		})
		return err
	})
	return uploads
}

func (db *DB) pruneFiles(folder string) {
	uploads := db.GetFiles()
	for index := range uploads {
		db.checkFile(uploads[index], folder)
	}
}

func (db *DB) Close() {
	_ = db.db.Close()
}

func (db *DB) getFile(fullname string) (*UploadDescription, error) {
	var dbValue string
	ud := &UploadDescription{}
	var err error
	err = db.db.View(func(tx *buntdb.Tx) error {
		dbValue, err = tx.Get(fullname)
		return err
	})
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(dbValue), ud)
	if err != nil {
		return nil, err
	}
	return ud, nil
}
