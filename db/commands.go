package db

import (
	"errors"

	"github.com/Feresey/banana_bot/model"
)

// Warn : warns a person in a chat
func Warn(person *model.Person, add bool) (total int, err error) {
	exist, err := checkPersonExists(person, warn)
	if err != nil {
		return 0, err
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

	res, err := db.Exec(
		`UPDATE warn
		SET total=$3 WHERE
			chatid=$1 AND userid=$2`,
		person.ChatID, person.UserID, total)
	if err != nil {
		return 0, err
	}

	rows, err := res.RowsAffected()
	if rows != 1 {
		return 0, errors.New("Error modify data")
	}

	return total, err
}

// Report : get user ids, who subscribed on reports
func Report(chatID int64) ([]int64, error) {
	rows, err := db.Query(
		`SELECT userid
		FROM `+report+`
		WHERE chatid=$1`, chatID)
	if err != nil {
		return nil, err
	}

	res := []int64{}
	err = rows.Scan(&res)
	return res, err
}
