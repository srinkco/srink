package shortener

import (
	"time"

	"github.com/AnimeKaizoku/cacher"
	"github.com/glebarez/sqlite"
	"github.com/srinkco/srink/utils/randomiser"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const DEF_SQLITE_FILE_NAME = "srink.sql"

type InSQLEngine struct {
	session *gorm.DB
	cache   *cacher.Cacher[string, string]
}

type Surl struct {
	Hash string `gorm:"primary_key"`
	Dest string
}

func newInSQLEngine(dbName string) *InSQLEngine {
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&Surl{})

	return &InSQLEngine{
		session: db,
		cache: cacher.NewCacher[string, string](&cacher.NewCacherOpts{
			TimeToLive:    time.Hour * 24 * 30,
			CleanInterval: time.Hour * 24 * 31,
		}),
	}
}

func (e *InSQLEngine) Shorten(url, hash string) string {
	if hash == "" {
		if hash, ok := e.getHash(url); ok {
			return hash
		}
		hash = randomiser.GetString(HASH_NUM)
	}
	e.saveUrl(hash, url)
	return hash
}

func (e *InSQLEngine) saveUrl(hash, dest string) {
	tx := e.session.Begin()
	v := &Surl{Hash: hash}
	tx.FirstOrCreate(v)
	v.Dest = dest
	tx.Save(v)
	tx.Commit()
	e.cache.Set(hash, dest)
}

func (e *InSQLEngine) getHash(url string) (string, bool) {
	var s Surl
	e.session.Where("dest = ?", url).Find(&s)
	return s.Hash, s.Hash != ""
}

func (e *InSQLEngine) GetUrl(hash string) string {
	url, ok := e.cache.Get(hash)
	if ok {
		return url
	}
	var s Surl
	e.session.Where("hash = ?", hash).Find(&s)
	if s.Dest != "" {
		e.cache.Set(hash, s.Dest)
	}
	return s.Dest
}
