package core

import (
	"crypto/md5"
	"math"
	"time"

	"bean/pkg/bear"
	"bean/pkg/meta"

	"github.com/pkg/errors"
)

type Core interface {
	Load() error
	ListTopics() ([]string, error)
	Get(topic string) (key, title string, content []byte, err error)
	Reflect(topic, key string, quality int) error
}

type CoreImpl struct {
	dbMeta meta.Storage
	dbBear bear.Storage
}

func New(dbMeta meta.Storage, dbBear bear.Storage) *CoreImpl {
	return &CoreImpl{dbMeta, dbBear}
}

func (c *CoreImpl) ListTopics() ([]string, error) {
	return c.dbMeta.GetTopicList()
}

func (c *CoreImpl) Get(topic string) (key, title string, content []byte, err error) {
	var card *meta.Card
	card, err = c.dbMeta.GetNextCard(topic)
	if err != nil {
		err = errors.Wrap(err, "get card from meta storage")
		return
	}

	note, err := c.dbBear.GetNoteByKey(card.Key)
	if err != nil {
		err = errors.Wrap(err, "get note from bear storage")
	}

	return card.Key, note.Title, note.Text, nil
}

func (c *CoreImpl) Reflect(topic, key string, quality int) error {
	card, err := c.dbMeta.GetCard(topic, key)
	if err != nil {
		return errors.Wrapf(err, "get card with key %q from topic %q", topic, key)
	}

	card.Quality = quality
	if quality < 3 {
		card.Attempt = 0
	} else {
		card.Attempt += 1
		card.Efactor = card.Efactor + (0.1 - (5-float32(quality))*(0.08+(5-float32(quality))*0.02))
		if card.Efactor < 1.3 {
			card.Efactor = 1.3
		}
	}
	if card.Quality < 4 {
		card.DaysInterval = 0
	} else {
		if card.Attempt == 1 {
			card.DaysInterval = 1
		} else if card.Attempt == 2 {
			card.DaysInterval = 6
		} else {
			card.DaysInterval = int(math.Round(float64(float32(card.DaysInterval) * card.Efactor)))
		}
	}
	card.NextRehearsal = time.Now().Add(time.Duration(card.DaysInterval) * 24 * time.Hour)

	if err := c.dbMeta.SetCard(topic, card); err != nil {
		return errors.Wrapf(err, "update card with key %q in topic %q in meta storage", key, topic)
	}

	return nil
}

func (c *CoreImpl) Load() error {
	// Select all notes
	notes, err := c.dbBear.GetAllNotes()
	if err != nil {
		return errors.Wrap(err, "get all notes from bear db")
	}

	// Foreach notes topic:
	for topic, noteGroup := range notes {
		//   Select all cards
		cards, err := c.dbMeta.GetCardsInTopic(topic)
		if err != nil {
			return errors.Wrapf(err, "get cards in topic %q", topic)
		}

		//   Foreach card:
		for _, card := range cards {
			//     if not in notes then drop card
			note, exists := noteGroup[card.Key]
			if !exists {
				if err := c.dbMeta.DropCard(topic, card.Key); err != nil {
					return errors.Wrapf(err, "drop card with key %q from topic %q", card.Key, topic)
				}
			}

			//     if equal checksum then drop note
			hash := md5.Sum(note.Text)
			if card.Hash != string(hash[:]) {
				delete(noteGroup, card.Key)
			}
		}

		//   Foreach note:
		for _, note := range noteGroup {
			//     insert new card
			hash := md5.Sum(note.Text)
			if err := c.dbMeta.SetCard(topic, &meta.Card{
				Key:           note.Key,
				Efactor:       2.5,
				NextRehearsal: time.Now(),
				Hash:          string(hash[:]),
			}); err != nil {
				return errors.Wrapf(err, "set card with key %q in topic %q: %v", note.Key, topic, err)
			}
		}
	}

	return nil
}
