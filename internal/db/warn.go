package db

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
)

const warn = schemaName + "warn"

var warnColumns = []string{"person_id", "total"}

// Warn доставляет плохишу +1 в карму (точнее в счётчик выговоров)
// параметр `add` регулирует добавлять ли в счётчик или убирать. Ну а вдруг человек хороший.
func (db *Database) Warn(ctx context.Context, person *Person, add bool) (int64, error) {
	var total int64
	err := db.tx(ctx, func(tx pgx.Tx) error {
		var err error
		total, err = db.warn(ctx, tx, person, add)
		return err
	})
	return total, err
}

// newWarn добавляет плохиша на карандаш.
func (db *Database) newWarn(ctx context.Context, tx pgx.Tx, personID int64, value int64) error {
	qb := psql.
		Insert(warn).
		Columns(warnColumns...).
		Values(personID, value)
	return zero(ctx, tx, qb)
}

func (db *Database) warn(
	ctx context.Context,
	tx pgx.Tx,
	person *Person,
	add bool,
) (int64, error) {
	id, err := db.checkPersonExists(ctx, tx, person)
	if err != nil {
		return 0, err
	}

	var total int64
	qb := psql.
		Select("total").
		From(warn).
		Where(squirrel.Eq{"person_id": id})
	if err := one(ctx, tx, qb, &total); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return 0, err
		}

		var setWarn int64 = 1
		// wtf???
		// даже если человек очень хороший, то делать ему одолжение не стоит.
		// А вот на карандаш можно и посадить для профилактики.
		if !add {
			setWarn--
		}
		return setWarn, db.newWarn(ctx, tx, id, setWarn)
	}

	// костылики
	total++
	if !add {
		total -= 2
	}
	if total < 0 {
		total = 0
	}

	upd := psql.
		Update(warn).
		Set("total", total)
	return total, zero(ctx, tx, upd)
}
