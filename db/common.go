package db

import "github.com/Feresey/banana_bot/model"

func checkPersonExists(person *model.Person, table string) bool {
	_, err := db.Exec(
		`SELECT total
		FROM `+table+`
		WHERE
			charid=$1 userid=$2`, person.ChatID, person.UserID)
	return err == nil
}

func createID(person *model.Person, table string) error {
	_, err := db.Exec(
		`INSERT INTO `+table+`
		VALUES
			($1, $2, $3)`,
		person.ChatID, person.UserID, 0)
	return err
}
