package db

func Warn(id int, add bool) (total int, err error) {
	if !checkIDExists(id) {
		createID(id)
	}
	tx, _ := DB.Begin()
	err = tx.QueryRow(
		`SELECT total
	FROM warn
	WHERE
		id=$1`,
		id,
	).Scan(&total)
	if err != nil {
		return
	}
	err = tx.Commit()
	if err != nil {
		return
	}

	if add {
		total++
	} else if total > 0 {
		total--
	}

	_, err = DB.Exec(
		"UPDATE warn SET total=$1 WHERE id=$2",
		total, id,
	)

	return
}

func checkIDExists(id int) bool {
	_, err := DB.Exec(
		`SELECT total
	FROM warn
	WHERE
		id=$id`, id)
	return err == nil
}

func createID(id int) bool {
	_, err := DB.Exec(
		`INSERT INTO warn
	VALUES
		($1, $2)`,
		id, 0)
	return err == nil
}
