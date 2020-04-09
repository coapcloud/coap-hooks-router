package hooks

import (
	"bytes"
	"errors"
	"fmt"
	"log"

	bolt "go.etcd.io/bbolt"
)

type Repository interface {
	CreateHook(h Hook) error
	ListHooksForOwner(owner string) ([]*Hook, error)
	Cleanup() error
}

type Hook struct {
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	Destination string `json:"destination"`
}

type repo struct {
	db *bolt.DB
}

func NewHooksRepository(dbFilename string) (*repo, error) {
	db, err := bolt.Open(dbFilename, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &repo{
		db: db,
	}, nil
}

func (r *repo) CreateHook(h Hook) (err error) {
	tx, err := r.db.Begin(true)
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			e := tx.Commit()
			if e != nil {
				log.Printf("error encountered committing tx: %v\n", e)
			}
		} else {
			e := tx.Rollback()
			if e != nil {
				log.Printf("error encountered rolling back: %v\n", e)
			}
		}
	}()

	buck, err := tx.CreateBucketIfNotExists([]byte(h.Owner))
	if err != nil {
		return err
	}

	k := []byte(h.Name)
	v := []byte(h.Destination)

	err = buck.Put(k, v)
	if err != nil {
		log.Println(err)
		return err
	}

	found := buck.Get(k)
	if found == nil {
		return errors.New("value not created properly")
	}

	if bytes.Compare(found, v) != 0 {
		return errors.New("value corrupted")
	}

	log.Printf("Hook: %+v created!\n", h)

	return nil
}

func (r *repo) ListHooksForOwner(owner string) ([]*Hook, error) {
	hooks := make([]*Hook, 0)

	err := r.db.View(func(tx *bolt.Tx) error {
		buck := tx.Bucket([]byte(owner))
		if buck == nil {
			return fmt.Errorf("no hooks found for owner: %s", owner)
		}

		cur := buck.Cursor()
		for k, v := cur.First(); k != nil; k, v = cur.Next() {
			hooks = append(hooks, &Hook{
				Owner:       owner,
				Name:        string(k),
				Destination: string(v),
			})
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return hooks, nil
}

func (r *repo) Cleanup() error {
	return r.db.Close()
}
