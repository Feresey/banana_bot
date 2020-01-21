package service

import (
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
	reply := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("@%s /warn [%d/5]", msg.ReplyToMessage.From.UserName, total))
	sendMsg(reply)
}

func addWarn(id int) (total int, err error) {
	if !checkIDExists(id) {
		createID(id)
	}
	err = db.QueryRow(
		`SELECT total
	FROM warn
	WHERE
		id=$1`,
		id,
	).Scan(&total)
	if err != nil {
		return
	}

	total++
	_, err = db.Query(
		"UPDATE warn SET total=$1 WHERE id=$2",
		total, id,
	)
	if err != nil {
		return
	}

	return
}

func checkIDExists(id int) bool {
	var b string
	err := db.QueryRow(
		`SELECT total
	FROM warn
	WHERE
		id=$id`, id).Scan(&b)
	log.Info(b)
	return err == nil
}

func createID(id int) bool {
	var b string
	err := db.QueryRow(
		`INSERT INTO warn
	VALUES
		($1, $2)`,
		id, 0).Scan(&b)
	log.Info(b)
	return err == nil
}
