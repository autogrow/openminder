package openminder

import (
	"encoding/json"

	"github.com/boltdb/bolt"
)

// BoltedJSON is a JSON database backed by bolt that handles common operations
// internally to provide a clean and easy to use interface
type BoltedJSON struct {
	db     *bolt.DB
	bucket []byte
}

// NewBoltedJSON returns a new database at the given filename in the given
// bucket name
func NewBoltedJSON(fn, bucket string) (*BoltedJSON, error) {
	db, err := bolt.Open(fn, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		return err
	})

	return &BoltedJSON{db, []byte(bucket)}, err
}

// Set will marshal the given object as JSON and store it at the given key
func (jdb *BoltedJSON) Set(key string, obj interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	return jdb.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(jdb.bucket)
		return b.Put([]byte(key), data)
	})
}

// Get will get the data from the given key and unmarshal it into the given object
func (jdb *BoltedJSON) Get(key string, obj interface{}) error {
	return jdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(jdb.bucket)
		data := b.Get([]byte(key))
		return json.Unmarshal(data, obj)
	})
}
