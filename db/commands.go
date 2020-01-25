package db

import "github.com/Feresey/banana_bot/model"

// Warn : warns a person in a chat
func Warn(person *model.Person, add bool) (total int, err error) {
	if !checkPersonExists(person, warn) {
		createID(person, warn)
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

	switch {
	case total >= 5:
		total = 6
	case total < 0:
		total = 0
	}

	_, err = db.Exec(
		`UPDATE warn
		SET total=$3 WHERE
			chatid=$1 AND userid=$2`,
		person.ChatID, person.UserID, total)

	return
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
