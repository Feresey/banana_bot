package db

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
)

const (
	chatsTableName = schemaName + "chats"
	chatsColumnID  = "chat_id"
)

func (db *Database) GetMyChats(ctx context.Context) ([]int64, error) {
	qb := psql.Select(chatsColumnID).From(chatsTableName)
	query, _, err := qb.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := db.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []int64
	for rows.Next() {
		var one int64
		if err := rows.Scan(&one); err != nil {
			return nil, err
		}
		res = append(res, one)
	}
	return res, rows.Err()
}

func (db *Database) AddChatWithMe(ctx context.Context, chatID int64) error {
	return db.tx(ctx, func(tx pgx.Tx) error {
		has := psql.
			Select("COUNT(*)").
			From(chatsTableName).
			Where(squirrel.Eq{chatsColumnID: chatID})
		var count int
		if err := one(ctx, tx, has, &count); err != nil {
			return err
		}
		if count != 0 {
			return nil
		}
		qb := psql.Insert(chatsTableName).Columns(chatsColumnID).Values(chatID)
		return zero(ctx, db.pool, qb)
	})
}

func (db *Database) DelChatWithMe(ctx context.Context, chatID int64) error {
	qb := psql.Delete(chatsTableName).Where(squirrel.Eq{chatsColumnID: chatID})
	return zero(ctx, db.pool, qb)
}
