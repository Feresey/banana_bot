package db

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Feresey/banana_bot/model"
)

// Warn : warns a person in a chat
func Warn(person *model.Person, add bool) (total int, err error) {
	exist, err := checkPersonExists(person, warn)
	if err != nil {
		return
	}

	if !exist {
		err = createID(person, warn)
		if err != nil {
			return
		}
	}

	err = db.QueryRow(
		`SELECT total
		FROM warn
		WHERE
			chatid=$1 AND userid=$2`,
		person.ChatID, person.UserID).
		Scan(&total)
	if err != nil {
		return
	}

	if add {
		total++
	} else {
		total--
	}

	if total < 0 {
		total = 0
	}

	res, err := db.Exec(
		`UPDATE warn
		SET total=$3 WHERE
			chatid=$1 AND userid=$2`,
		person.ChatID, person.UserID, total)
	if err != nil {
		return
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return
	}

	if rows != 1 {
		return total, errors.New("Error modify data")
	}
	return
}

// Report : get user ids, who subscribed on reports
func Report(chatID int64) ([]int, error) {
	rows, err := db.Query(
		`SELECT userid
		FROM `+admins+`
		WHERE chatid=$1 AND subscribed=true`, chatID)
	if err != nil {
		return nil, err
	}

	res := []int{}
	for rows.Next() {
		var tmp int
		_ = rows.Scan(&tmp)
		res = append(res, tmp)
	}

	return res, rows.Err()
}

// GetChatList : get ids of all chats with bot
func GetChatList() ([]int64, error) {
	chatsRaw, err := db.Query(
		`SELECT DISTINCT chatid
		FROM ` + admins)
	if err != nil {
		return nil, err
	}

	chat := []int64{}

	for chatsRaw.Next() {
		var tmp int64
		_ = chatsRaw.Scan(&tmp)
		chat = append(chat, tmp)
	}

	return chat, chatsRaw.Err()
}

// GetAdmins : get all admins from chat
func GetAdmins(chatid int64) ([]int, error) {
	adminsRaw, err := db.Query(
		`SELECT userid
		FROM `+admins+`
	WHERE chatid=$1`, chatid)
	if err != nil {
		return nil, err
	}

	admins := []int{}

	for adminsRaw.Next() {
		var tmp int
		_ = adminsRaw.Scan(&tmp)
		admins = append(admins, tmp)
	}

	return admins, adminsRaw.Err()
}

// GetChatsForAdmin : get all chats for admin
func GetChatsForAdmin(userid int) ([]int64, error) {
	adminsRaw, err := db.Query(
		`SELECT chatid
		FROM `+admins+`
	WHERE userid=$1`, userid)
	if err != nil {
		return nil, err
	}

	admins := []int64{}

	for adminsRaw.Next() {
		var tmp int64
		_ = adminsRaw.Scan(&tmp)
		admins = append(admins, tmp)
	}

	return admins, adminsRaw.Err()
}

// SetAdmins : get all admins from chat
func SetAdmins(chatid int64, pipls []int) error {
	if len(pipls) == 0 {
		return fmt.Errorf("Fuckoff")
	}
	adminS := make([]string, 0, len(pipls))
	for idx := range pipls {
		adminS = append(adminS, strconv.Itoa(pipls[idx]))
	}

	// drop non-admins
	_, err := db.Exec(fmt.Sprintf(
		`DELETE FROM %s WHERE (chatid, userid) IN
			(SELECT chatid, userid
				FROM `+admins+` WHERE
					chatid=$1
					AND NOT userid IN (%s)
			)`, admins, strings.Join(adminS, ",")),
		chatid)
	if err != nil {
		return err
	}

	q := fmt.Sprintf(
		`INSERT INTO `+admins+`
		SELECT %d, t.id as userid, false
			FROM (VALUES (%s)) AS t(id)
				EXCEPT SELECT $1::BIGINT, admins.userid, false
				FROM admins
					WHERE admins.chatid=$1::BIGINT`, chatid, strings.Join(adminS, "),("))
	log.Info(q)
	// insert new admins
	res, err := db.Exec(q, chatid)
	if err != nil {
		return err
	}

	total, err := res.RowsAffected()
	if err != nil {
		return err
	}

	log.Infof("%d admins added into %d", total, chatid)
	return nil
}
