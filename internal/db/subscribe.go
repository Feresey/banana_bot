package db

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
)

const subscriptionsTableName = schemaName + "subscriptions"

var subscriptionsColumns = []string{"sub_id"}

// Report
// Если в чате появился плохой человечек, то более хорошие человечки настучат об этом
// добрым человечкам (админам), которые посадят плохого человечка на карандаш или сразу на молоток (banhammer).
// Не все админы хотят видеть такой спам, поэтому на него сначала надо подписаться.
// И уже потом, по подписке, будут прилетать сообщения о плохишах в чате.
func (db *Database) Report(ctx context.Context, chatID int64) ([]int64, error) {
	var res []int64
	err := db.tx(ctx, func(tx pgx.Tx) error {
		ids, err := db.report(ctx, tx, chatID)
		res = ids
		return err
	})
	return res, err
}

func (db *Database) report(ctx context.Context, tx pgx.Tx, chatID int64) (res []int64, err error) {
	qb := psql.
		Select("user_id").
		From(personsTableName).
		Join(subscriptionsTableName + " ON (sub_id = person_id)").
		Where(squirrel.Eq{"chat_id": chatID})

	query, params, err := qb.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := db.pool.Query(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tmp int64
		err := rows.Scan(&tmp)
		if err != nil {
			return nil, err
		}
		res = append(res, tmp)
	}
	rows.Close()

	return res, rows.Err()
}

// Subscribe
// Подписка на спам (добровольная).
func (db *Database) Subscribe(ctx context.Context, p *Person) error {
	return db.tx(ctx, func(tx pgx.Tx) error {
		id, err := db.GetPersonID(ctx, tx, p)
		if err != nil {
			return err
		}
		qb := psql.
			Insert(subscriptionsTableName).
			Columns(subscriptionsColumns...).
			Values(id).
			Suffix("ON CONFLICT (" + subscriptionsColumns[0] + ") DO NOTHING")
		return zero(ctx, tx, qb)
	})
}

// Unsubscribe
// Отписка от спама (пока бесплатная).
func (db *Database) Unsubscribe(ctx context.Context, p *Person) error {
	return db.tx(ctx, func(tx pgx.Tx) error {
		id, err := db.GetPersonID(ctx, tx, p)
		if err != nil {
			return err
		}
		qb := psql.
			Delete(subscriptionsTableName).
			Where(squirrel.Eq{subscriptionsColumns[0]: id})
		return zero(ctx, tx, qb)
	})
}
