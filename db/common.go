package db

import "github.com/Feresey/banana_bot/model"

func checkPersonExists(person *model.Person, table string) (bool, error) {
	res, err := db.Exec(
		`SELECT total
		FROM `+table+`
		WHERE
			chatid=$1 AND userid=$2`, person.ChatID, person.UserID)
	if err != nil {
		return false, err
	}

	rows, err := res.RowsAffected()
	return rows != 0, err
}

func createID(person *model.Person, table string) error {
	_, err := db.Exec(
		`INSERT INTO `+table+`
		VALUES
			($1, $2, $3)`,
		person.ChatID, person.UserID, 0)
	return err
}
