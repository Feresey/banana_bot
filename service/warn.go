package service

import (
	"context"
	"fmt"

	"github.com/Feresey/bot-tg/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func makeWarn(msg *tgbotapi.Message) {
	if msg.ReplyToMessage == nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "Не указано кому /warn кидать")
		sendMsg(reply)
		return
	}

	total, err := addWarn(msg.ReplyToMessage.From.ID)
	if err != nil {
		log.Warn(err)
	}
	reply := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("@%s /warn [%d/5]", msg.ReplyToMessage.From.FirstName, total))
	sendMsg(reply)
}

func addWarn(id int) (int, error) {
	var total int
	err := db.DB.QueryRow(context.Background(),
		`SELECT total 
	FROM $table 
	WHERE
		id=$id`,
		map[string]interface{}{
			"table": "warn",
			"id":    id,
		}).Scan(&total)
	if err != nil {
		return 0, err
	}

	_, err = db.DB.Query(context.Background(),
		"UPDATE $table SET total=$total WHERE id=$id",
		map[string]interface{}{
			"table": "warn",
			"total": 1,
			"id":    id,
		})
	if err != nil {
		return 0, err
	}

	return total, nil
}
