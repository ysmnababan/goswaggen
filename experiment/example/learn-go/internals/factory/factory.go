package factory

import (
	"gorm.io/gorm"

	"learn-go/internals/pkg/database"

	"learn-go/config"
	"learn-go/internals/pkg/redisutil"
)

type Factory struct {
	Db *gorm.DB

	Redis *redisutil.Redis
}

func NewFactory() *Factory {

	f := &Factory{}

	f.SetupDb()

	f.SetupRedis()
	return f
}

func (f *Factory) SetupDb() {
	db := database.Connection()
	f.Db = db
}

func (f *Factory) SetupRedis() {
	cfg := config.Get().Redis
	f.Redis = redisutil.NewRedis(cfg)
}

func (f *Factory) SetupRepository() {
	if f.Db == nil {
		panic("Failed setup repository, db is undefined")
	}
}
