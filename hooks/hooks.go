package hooks

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	bolt "go.etcd.io/bbolt"
)

type Repository interface {
	CreateHook(h Hook) error
	ListAllHooks() ([]*Hook, error)
	ListHooksForOwner(owner string) ([]*Hook, error)
	DeleteHooksForOwner(owner string) error
	DeleteHookByOwnerAndName(owner, name string) error
	Events() chan HookAPIEvent
	Cleanup() error
}

type Hook struct {
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	Destination string `json:"destination"`
}

func (h Hook) Fire(bdy io.Reader) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, h.Destination, bdy)
	if err != nil {
		return nil, err
	}

	c := http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("received bad status code from webhook destination hook: %d", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}

type EventType int

const (
	EventTypeCreate = iota
	EventTypePut    = iota
	EventTypeDeleteForOwner
	EventTypeDelete
)

type HookAPIEvent struct {
	EventType EventType
	Hooks     []*Hook
}

type repo struct {
	db     *bolt.DB
	events chan HookAPIEvent
}

func NewHooksRepository(dbFilename string) (*repo, error) {
	db, err := bolt.Open(dbFilename, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &repo{
		db:     db,
		events: make(chan HookAPIEvent),
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

			r.events <- HookAPIEvent{
				EventType: EventTypeCreate,
				Hooks:     []*Hook{&h},
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

func (r *repo) ListAllHooks() ([]*Hook, error) {
	hooks := make([]*Hook, 0)

	err := r.db.View(func(tx *bolt.Tx) error {
		tx.ForEach(func(name []byte, buck *bolt.Bucket) error {
			bucketHooks, err := r.ListHooksForOwner(string(name))
			if err != nil {
				return err
			}

			for _, v := range bucketHooks {
				hooks = append(hooks, v)
			}

			return nil
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return hooks, nil
}

func (r *repo) DeleteHooksForOwner(owner string) (err error) {
	tx, err := r.db.Begin(true)
	if err != nil {
		return err
	}

	var hooks []*Hook

	defer func() {
		if err == nil {
			e := tx.Commit()
			if e != nil {
				log.Printf("error encountered committing tx: %v\n", e)
			}

			r.events <- HookAPIEvent{
				EventType: EventTypeDeleteForOwner,
				Hooks:     hooks,
			}
		} else {
			e := tx.Rollback()
			if e != nil {
				log.Printf("error encountered rolling back: %v\n", e)
			}
		}
	}()

	hooks, err = r.ListHooksForOwner(owner)
	if err != nil {
		return err
	}

	err = tx.DeleteBucket([]byte(owner))
	if err != nil {
		return err
	}

	log.Printf("Hooks deleted for owner: %s!\n", owner)

	return nil
}

func (r *repo) DeleteHookByOwnerAndName(owner, name string) (err error) {
	tx, err := r.db.Begin(true)
	if err != nil {
		return err
	}

	var h Hook

	defer func() {
		if err == nil {
			e := tx.Commit()
			if e != nil {
				log.Printf("error encountered committing tx: %v\n", e)
			}

			r.events <- HookAPIEvent{
				EventType: EventTypeDelete,
				Hooks:     []*Hook{&h},
			}
		} else {
			e := tx.Rollback()
			if e != nil {
				log.Printf("error encountered rolling back: %v\n", e)
			}
		}
	}()

	buck := tx.Bucket([]byte(owner))
	if buck == nil {
		return fmt.Errorf("could not find owner: %s", owner)
	}

	k := []byte(name)

	v := buck.Get(k)
	if v == nil {
		return fmt.Errorf("could not find hook: %s to delete for owner: %s", name, owner)
	}

	h = Hook{
		Owner:       owner,
		Name:        name,
		Destination: string(v),
	}

	err = buck.Delete(k)
	if err != nil {
		return err
	}

	log.Printf("Hook: %+v deleted!\n", h)

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

func (r *repo) Events() chan HookAPIEvent {
	return r.events
}

func (r *repo) Cleanup() error {
	return r.db.Close()
}
