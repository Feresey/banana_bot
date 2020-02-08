package db

import (
	"fmt"

	"github.com/Feresey/banana_bot/model"
)

func checkPersonExists(person *model.Person, table string) (bool, error) {
	res, err := db.Exec(
		`SELECT	FROM `+table+`
		WHERE
			chatid=$1 AND userid=$2`, person.ChatID, person.UserID)
	if err != nil {
		return false, err
	}

	rows, err := res.RowsAffected()
	return rows != 0, err
}

func createPerson(person *model.Person, table string) error {
	num, err := db.Exec(
		`INSERT INTO `+table+`
		(chatid, userid)
		VALUES
			($1, $2)`,
		person.ChatID, person.UserID)
	if err != nil {
		return err
	}

	if total, err := num.RowsAffected(); err == nil {
		if total != 1 {
			return fmt.Errorf("Modified %d rows. Expected 1", total)
		}
	} else {
		return err
	}

	return nil
}

func deletePerson(person *model.Person, table string) error {
	num, err := db.Exec(
		`DELETE FROM `+table+`
		WHERE
		chatid=$1 AND userid=$2`,
		person.ChatID, person.UserID)
	if err != nil {
		return err
	}

	if total, err := num.RowsAffected(); err == nil {
		if total != 1 {
			return fmt.Errorf("Modified %d rows. Expected 1", total)
		}
	} else {
		return err
	}

	return nil
}
