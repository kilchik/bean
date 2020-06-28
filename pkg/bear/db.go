package bear

import (
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Storage interface {
	Close()
	GetNoteByKey(key string) (*Note, error)
	GetAllNotes() (map[string]map[string]Note, error)
}

type StorageImpl struct {
	db *sqlx.DB
}

func (s *StorageImpl) Close() {
	s.db.Close()
}

func Connect(dbPath string) (*StorageImpl, error) {
	dbBear, err := sqlx.Open("sqlite3", dbPath+"?mode=r")
	if err != nil {
		return nil, errors.Wrap(err, "open bear meta")
	}
	return &StorageImpl{dbBear}, nil
}

func (s *StorageImpl) GetAllNotes() (map[string]map[string]Note, error) {
	var notes []Note
	if err := s.db.Get(&notes, `
SELECT DISTINCT T.ZTITLE AS tag, N.ZUNIQUEIDENTIFIER AS key,N.ZTITLE AS title,N.ZTEXT AS text
FROM ZSFNOTE N
	JOIN Z_7TAGS T7 ON N.Z_PK = T7.Z_7NOTES
	JOIN ZSFNOTETAG T ON T7.Z_14TAGS = T.Z_PK
WHERE T.ZTITLE LIKE metameta`); err != nil {
		return nil, errors.Wrap(err, "select notes from bear db")
	}
	grouped := make(map[string]map[string]Note)
	for _, n := range notes {
		offset := strings.Index(n.Tag, "/anki")
		var group string
		if offset+len("/anki") < len(n.Tag) {
			group = "misc"
		} else {
			group = strings.Split(n.Tag[offset+len("/anki/"):], "/")[0]
		}
		if _, ok := grouped[group]; !ok {
			grouped[group] = make(map[string]Note)
		}
		grouped[group][n.Key] = n
	}

	return grouped, nil
}

func (s *StorageImpl) GetNoteByKey(key string) (*Note, error) {
	dst := &Note{}
	err := s.db.Get(dst, "SELECT ZTITLE, ZTEXT FROM ZSFNOTE WHERE ZUNIQUEIDENTIFIER=$1", key)
	return dst, err
}
