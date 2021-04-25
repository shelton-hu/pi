package mysql

import (
	"context"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"

	"github.com/shelton-hu/logger"

	"github.com/shelton-hu/pi/config"
)

type Mysql struct {
	*gorm.DB
}

var dbPool = make(map[string]*Mysql)

func ConnectMysql(ctx context.Context, mysqlConfigs map[string]config.Mysql) {
	for name, mysqlConfig := range mysqlConfigs {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&loc=Local", mysqlConfig.User, mysqlConfig.Password, mysqlConfig.Host, mysqlConfig.Port, mysqlConfig.Database, mysqlConfig.Charset)
		db, err := gorm.Open(mysqlConfig.Dialect, dsn)
		if err != nil {
			panic(fmt.Errorf("fatal error: connect database: %s\n", err))
		}

		db.DB().SetMaxIdleConns(mysqlConfig.MaxIdleConnNum)
		db.DB().SetMaxOpenConns(mysqlConfig.MaxOpenConnNum)
		db.BlockGlobalUpdate(true)
		db.InstantSet("gorm:save_associations", false)
		db.InstantSet("gorm:association_save_reference", false)

		addGormCallbacks(db)

		dbPool[name] = &Mysql{db}
	}

	if _, ok := dbPool["default"]; !ok {
		panic("default db is not set")
	}
}

func CloseMysql(ctx context.Context) {
	for _, db := range dbPool {
		if err := db.Close(); err != nil {
			logger.Error(ctx, err.Error())
		}
	}
}

func GetConnect(ctx context.Context, name ...string) *Mysql {
	key := "default"
	if len(name) > 0 {
		key = name[0]
	}

	db, ok := dbPool[key]
	if !ok {
		logger.Error(ctx, "db name is wrong")
		db = dbPool["default"]
	}

	db.setSpanToGorm(ctx)

	return db
}
