package db

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
)

const personsTableName = schemaName + "persons"

var personColumns = []string{"person_id", "chat_id", "user_id"}

// Person : структура лежащая в базе
type Person struct {
	ID     int64
	ChatID int64
	UserID int64
}

// GetPersonID возвращает некоторый id человечка. Если готового id нет, он создаётся.
func (db *Database) GetPersonID(
	ctx context.Context,
	tx executor,
	person *Person,
) (id int64, err error) {
	id, err = db.checkPersonExists(ctx, tx, person)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return 0, err
		}
	}
	if id != 0 {
		return id, nil
	}
	return db.createPerson(ctx, tx, person)
}

// где мой pgx.ErrNotExists
func (db *Database) checkPersonExists(
	ctx context.Context,
	tx executor,
	person *Person,
) (id int64, err error) {
	qb := psql.
		Select(personColumns[0]).
		From(personsTableName).
		Where(squirrel.Eq{
			"chat_id": person.ChatID,
			"user_id": person.UserID,
		})
	err = one(ctx, tx, qb, &id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return id, err
}

func (db *Database) createPerson(
	ctx context.Context,
	tx executor,
	person *Person,
) (id int64, err error) {
	qb := psql.
		Insert(personsTableName).
		Columns(personColumns[1:]...).
		Values(person.ChatID, person.UserID).
		Suffix("RETURNING person_id")
	err = one(ctx, tx, qb, &id)
	return id, err
}

func (db *Database) deletePerson(ctx context.Context, tx executor, id int64) error {
	qb := psql.
		Delete(personsTableName).
		Where(squirrel.Eq{personColumns[0]: id})
	return zero(ctx, tx, qb)
}
