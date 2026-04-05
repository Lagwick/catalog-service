package rcpostgres

import (
	"database/sql"

	"github.com/Lagwick/catalog-service/internal/app/entity"
	"github.com/Lagwick/catalog-service/internal/app/util"
)

func RowsAffected(res sql.Result) int64 {
	ra, _ := res.RowsAffected()
	return ra
}

func UpdateErr(res sql.Result, err error) error {
	if err == nil && RowsAffected(res) == 0 {
		err = entity.ErrNotFound
	}
	return util.ReplaceErr1(err, sql.ErrNoRows, entity.ErrNotFound)
}

func DeleteErr(err error) error {
	return util.ReplaceErr1(err, sql.ErrNoRows, nil)
}
