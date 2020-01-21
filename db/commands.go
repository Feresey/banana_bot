package db

func Warn(id int, add bool) (total int, err error) {
	if !checkIDExists(id) {
		createID(id)
	}
	err = QueryRow(
		`SELECT total
	FROM warn
	WHERE
		id=$1`,
		id,
	).Scan(&total)
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
	if err != nil {
		return
	}

	return
}

func checkIDExists(id int) bool {
	var b string
	err := QueryRow(
		`SELECT total
	FROM warn
	WHERE
		id=$id`, id).Scan(&b)
	log.Info(b)
	return err == nil
}

func createID(id int) bool {
	var b string
	err := QueryRow(
		`INSERT INTO warn
	VALUES
		($1, $2)`,
		id, 0).Scan(&b)
	log.Info(b)
	return err == nil
}
