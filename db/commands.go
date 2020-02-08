package db

import (
	"errors"

	"github.com/Feresey/banana_bot/model"
)

// Warn : warns a person in a chat
func Warn(person *model.Person, add bool) (total int, err error) {
	exist, err := checkPersonExists(person, warn)
	if err != nil {
		return
	}

	if !exist {
		err = createPerson(person, warn)
		if err != nil {
			return
		}
	}

	err = db.QueryRow(
		`SELECT total
		FROM `+warn+`
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
		`UPDATE `+warn+`
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
		FROM `+subscriptions+`
		WHERE chatid=$1`, chatID)
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

// Subscribe : get user ids, who subscribed on reports
func Subscribe(p *model.Person) (err error) {
	var ok bool
	ok, err = checkPersonExists(p, subscriptions)
	if err != nil {
		return err
	}

	if ok {
		err = errors.New("Person already subscribed")
	} else {
		createPerson(p, subscriptions)
	}

	return
}

// UnSubscribe : get user ids, who subscribed on reports
func UnSubscribe(p *model.Person) (err error) {
	var ok bool
	ok, err = checkPersonExists(p, subscriptions)
	if err != nil {
		return err
	}

	if !ok {
		err = errors.New("Person not exists")
	} else {
		deletePerson(p, subscriptions)
	}

	return
}
