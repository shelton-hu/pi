package mysql

import (
	"errors"

	"github.com/jinzhu/gorm"
)

type TransFunc func(tx *gorm.DB) error

func (m *Mysql) MysqlTransaction(closures ...TransFunc) (err error) {
	tx := m.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			switch r := r.(type) {
			case error:
				err = r
			case string:
				err = errors.New(r)
			default:
				err = errors.New("system internal error")
			}
		}
	}()

	if tx.Error != nil {
		return tx.Error
	}

	for _, closure := range closures {
		if err := closure(tx); err != nil {
			tx.Rollback()
			return err
		}
		if tx.Error != nil {
			tx.Rollback()
			return tx.Error
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
