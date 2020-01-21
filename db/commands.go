package db

func Warn(id int, add bool) (total int, err error) {
	if !checkIDExists(id) {
		createID(id)
	}
	_, err = QueryRow(
		`SELECT total
	FROM warn
	WHERE
		id=$1`,
		id,
	)
	if err != nil {
		return
	}

	if add {
		total++
	} else if total > 0 {
		total--
	}

	_, err = Query(
		"UPDATE warn SET total=$1 WHERE id=$2",
		total, id,
	)
	return
}

func checkIDExists(id int) bool {
	_, err := QueryRow(
		`SELECT total
	FROM warn
	WHERE
		id=$id`, id)
	return err == nil
}

func createID(id int) bool {
	_, err := QueryRow(
		`INSERT INTO warn
	VALUES
		($1, $2)`,
		id, 0)
	return err == nil
}
