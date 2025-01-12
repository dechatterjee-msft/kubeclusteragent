package db

import (
	"fmt"
	"os"
	"sync"

	"github.com/boltdb/bolt"
)

const (
	FilePermission                 = 0644
	DBClusterTableName             = "cluster"
	DBName                         = "state.db"
	DBFileLocation                 = "/opt/agent/kubeclusteragent/store"
	DBClusterStatusTableName       = "cluster-status"
	DBCustomisationStatus          = "customisation-status"
	DBClusterAuditHistoryTableName = "cluster-audit-history"
)

var db *bolt.DB

type Store struct {
	TableName      string
	DBName         string
	FilePermission os.FileMode
}

var mu = sync.Mutex{}

func (d Store) Set(key string, value interface{}) error {
	if db == nil {
		err := d.Launch()
		if err != nil {
			return err
		}
	}
	err := db.Update(func(tx *bolt.Tx) error {
		b, err2 := tx.CreateBucketIfNotExists([]byte(d.TableName))
		if err2 != nil {
			return err2
		}
		err := b.Put([]byte(key), []byte(fmt.Sprintf("%v", value)))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (d Store) Get(key string) interface{} {
	var result interface{}
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(d.TableName))
		if b == nil {
			return nil
		}
		dbResponse := b.Get([]byte(key))
		if dbResponse == nil {
			return nil
		}
		result = string(dbResponse)
		if result == "" {
			return fmt.Errorf("%v", "Unable to get key")
		}
		return nil
	})
	if err != nil {
		return nil
	}
	return result
}

func (d Store) DeleteAll() error {
	err := db.Update(func(tx *bolt.Tx) error {
		delErr := tx.DeleteBucket([]byte(DBClusterTableName))
		if delErr != nil {
			return delErr
		}
		delErr = tx.DeleteBucket([]byte(DBClusterStatusTableName))
		if delErr != nil {
			return delErr
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *Store) Launch() error {
	mu.Lock()
	var err error
	if db == nil {
		db, err = bolt.Open(DBFileLocation+"/"+d.DBName, d.FilePermission, nil)
	}
	if err != nil {
		return err
	}
	mu.Unlock()
	return nil
}

func (d *Store) Close() {
	err := db.Close()
	if err != nil {
		return
	}
}

func (d *Store) Connect(tableName string) *Store {
	d.FilePermission = FilePermission
	d.TableName = tableName
	d.DBName = DBName
	err := d.Launch()
	if err != nil {
		return nil
	}
	return d
}
