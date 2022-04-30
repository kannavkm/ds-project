package main

import (
	"github.com/boltdb/bolt"
	"github.com/buraksezer/consistent"
	"github.com/cespare/xxhash/v2"
)

var (
	groups = []byte("Groups")
)

// In your code, you probably have a custom data type
// for your cluster members. Just add a String function to implement
// consistent.Member interface.
type group string

func (m group) String() string {
	return string(m)
}

// consistent package doesn't provide a default hashing function.
// You should provide a proper one to distribute keys/members uniformly.
type hasher struct{}

func (h hasher) Sum64(data []byte) uint64 {
	// you should use a proper hash function for uniformity.
	return xxhash.Sum64(data)
}

// maps the key value to the group id, should be persistent right?
type consistentHashHandler struct {
	c  *consistent.Consistent
	db *bolt.DB
}

func newConsistentHashHandler(db *bolt.DB) (*consistentHashHandler, error) {
	cfg := consistent.Config{
		// groups are the members of the ring
		// each key can map to
		PartitionCount:    271,
		ReplicationFactor: 20,
		Load:              1.25,
		Hasher:            hasher{},
	}

	c := consistent.New(nil, cfg)
	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Create all the buckets
	if _, err := tx.CreateBucketIfNotExists(groups); err != nil {
		return nil, err
	}
	tx.Commit()

	return &consistentHashHandler{
		c:  c,
		db: db,
	}, nil
}

func (ch *consistentHashHandler) addGroup(id string) error {
	ch.c.Add(group(id))
	txn, err := ch.db.Begin(true)
	if err != nil {
		return err
	}
	defer txn.Rollback()
	bucket := txn.Bucket(groups)
	if err := bucket.Put([]byte(id), []byte("")); err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func (ch *consistentHashHandler) removeGroup(id string) error {
	ch.c.Remove(id)
	txn, err := ch.db.Begin(true)
	if err != nil {
		return err
	}
	defer txn.Rollback()
	bucket := txn.Bucket(groups)
	if err := bucket.Delete([]byte(id)); err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func (ch *consistentHashHandler) getGroupForKey(key string) (string, error) {
	grp := ch.c.LocateKey([]byte(key))
	return grp.String(), nil
}
