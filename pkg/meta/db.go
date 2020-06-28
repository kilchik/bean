package meta

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

type Storage interface {
	GetCard(topic, key string) (*Card, error)
	DropCard(topic, key string) error
	GetCardsInTopic(topic string) ([]*Card, error)
	SetCard(topic string, card *Card) error
	GetNextCard(topic string) (*Card, error)
	GetTopicList() ([]string, error)
	Close()
}

type StorageImpl struct {
	db *bolt.DB
}

func Connect(dbPath string) (*StorageImpl, error) {
	dbBean, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, errors.Wrap(err, "open meta db")
	}
	return &StorageImpl{dbBean}, nil
}

func (s *StorageImpl) Close() {
	s.db.Close()
}

func (s *StorageImpl) GetNextCard(topic string) (*Card, error) {
	var res *Card
	var min *time.Time
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(topic))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
			cur := &Card{}
			json.Unmarshal(v, cur)
			if res.NextRehearsal.Before(time.Now()) {
				if min == nil || min.After(cur.NextRehearsal) {
					min = &cur.NextRehearsal
					res = cur
					res.Key = string(k)
				}
			}
		}
		return nil
	})
	return res, err
}

func (s *StorageImpl) GetCard(topic, key string) (*Card, error) {
	res := &Card{}
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(topic))
		encoded := b.Get([]byte(key))
		if len(encoded) == 0 {
			return errors.New("not found")
		}
		json.Unmarshal(encoded, res)
		return nil
	})
	return res, err
}

func (s *StorageImpl) SetCard(topic string, card *Card) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(topic))
		encoded, _ := json.Marshal(card)
		return b.Put([]byte(card.Key), encoded)
	})
}

func (s *StorageImpl) GetTopicList() ([]string, error) {
	var topics []string
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			topics = append(topics, string(name))
			return nil
		})
	})
	return topics, err
}

func (s *StorageImpl) GetCardsInTopic(topic string) ([]*Card, error) {
	var cards []*Card
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(topic))
		return b.ForEach(func(k, v []byte) error {
			c := &Card{Key: string(k)}
			json.Unmarshal(v, c)
			cards = append(cards, c)
			return nil
		})
	})
	return cards, err
}

func (s *StorageImpl) DropCard(topic, key string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(topic))
		return b.Delete([]byte(key))
	})
}
