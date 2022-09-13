package db

import (
	"errors"
	"path"

	log "github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"

	"github.com/vishhvaan/lab-bot/logging"
)

type database struct {
	db     *bolt.DB
	logger *log.Entry
}

var botDB database

func Open() {
	botDB.logger = logging.CreateNewLogger("database", "database")
	exePath := logging.FindExeDir()
	dbPath := path.Join(exePath, dbFile)

	var err error
	botDB.db, err = bolt.Open(dbPath, 0600, nil)
	if err != nil {
		botDB.logger.WithError(err).Panic("database could not be opened")
	} else {
		botDB.logger.Info("opened database")
	}
	defer botDB.db.Close()
}

func bucketFinder(tx *bolt.Tx, path []string) (b *bolt.Bucket, err error) {
	notExist := errors.New("bucket does not exist at path")
	var bucket *bolt.Bucket
	if len(path) == 0 {
		return nil, notExist
	} else {
		bucket = tx.Bucket([]byte(path[0]))
		path = path[1:]
		for len(path) > 0 && bucket != nil {
			bucket = bucket.Bucket([]byte(path[0]))
			path = path[1:]
		}
	}

	if bucket == nil {
		return nil, notExist
	} else {
		return bucket, nil
	}
}

func bucketCreator(tx *bolt.Tx, path []string) (b *bolt.Bucket, err error) {
	if len(path) == 0 {
		return nil, errors.New("cannot create empty bucket")
	} else {
		bucket, err := tx.CreateBucketIfNotExists([]byte(path[0]))
		if err != nil {
			return nil, err
		}
		path = path[1:]
		for len(path) > 0 {
			bucket, err = bucket.CreateBucketIfNotExists([]byte(path[0]))
			if err != nil {
				return nil, err
			}
			path = path[1:]
		}

		return bucket, nil
	}
}

func CheckBucketExists(path []string) (exists bool) {
	err := botDB.db.View(func(tx *bolt.Tx) error {
		b, err := bucketFinder(tx, path)
		if b != nil {
			exists = true
			return nil
		}
		return err
	})

	if err != nil {
		botDB.logger.WithError(err).WithFields(log.Fields{
			"path": path,
		}).Error("cannot check if bucket exists")
	}

	return exists
}

func CreateBucket(path []string) error {
	err := botDB.db.Update(func(tx *bolt.Tx) error {
		_, err := bucketCreator(tx, path)
		return err
	})

	if err != nil {
		botDB.logger.WithError(err).WithField("path", path).Error("cannot create bucket at path")
	}

	return err
}

func AddValue(path []string, key string, value []byte) error {
	err := botDB.db.Update(func(tx *bolt.Tx) error {
		b, err := bucketFinder(tx, path)
		if err != nil {
			return err
		}
		err = b.Put([]byte(key), value)
		return err
	})

	if err != nil {
		botDB.logger.WithError(err).WithFields(log.Fields{
			"path":  path,
			"key":   key,
			"value": value,
		}).Error("cannot update database")
	}

	return err
}

func ReadValue(path []string, key string) (value []byte, err error) {
	// returns nil if key doesn't exist or is a nested bucket
	err = botDB.db.View(func(tx *bolt.Tx) error {
		b, err := bucketFinder(tx, path)
		if err != nil {
			return err
		}
		value = b.Get([]byte(key))
		return nil
	})

	if err != nil {
		botDB.logger.WithError(err).WithFields(log.Fields{
			"path": path,
			"key":  key,
		}).Error("cannot read database")
	}

	return value, err
}

func GetAllKeysValues(path []string) (keys [][]byte, values [][]byte, err error) {
	err = botDB.db.View(func(tx *bolt.Tx) error {
		b, err := bucketFinder(tx, path)
		if err != nil {
			return err
		}

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			keys = append(keys, k)
			values = append(values, v)
		}

		return nil
	})

	if err != nil {
		botDB.logger.WithError(err).WithFields(log.Fields{
			"path": path,
		}).Error("cannot get keys or values in this bucket")
	}

	return keys, values, err
}
