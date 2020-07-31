package db

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
)

const subscriptions = schemaName + "subscriptions"

var subscriptionsColumns = []string{"id", "person_id", "chat_id"}

type subscription struct {
	ID       int64
	PersonID int64
	ChatID   int64
}

// Report
// Если в чате появился плохой человечек, то более хорошие человечки настучат об этом
// добрым человечкам (админам), которые посадят плохого человечка на карандаш или сразу на молоток (banhammer).
// Не все админы хотят видеть такой спам, поэтому на него сначала надо подписаться.
// И уже потом, по подписке, будут прилетать сообщения о плохишах в чате.
func (db *Database) Report(ctx context.Context, chatID int64) (res []int64, err error) {
	qb := psql.
		Select(subscriptions).
		Columns("person_id").
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
			Insert(subscriptions).
			Columns(subscriptionsColumns[1:]...).
			Values(id, p.ChatID)
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
		return db.deletePerson(ctx, db.pool, id)
	})
}
