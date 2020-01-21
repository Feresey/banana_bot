package db

// func QueryRow(q string, args ...interface{}) (*sql.Row, error) {
// 	log.Info("query:", q)
// 	res, err := db.Exec(q, args...)
// 	if err != nil {
// 		return nil, err
// 	}
// 	res.
// 	return row, nil
// }

// func Query(q string, args ...interface{}) (rows *sql.Rows, err error) {
// 	log.Info("query:", q)
// 	conn, err := db.Conn(context.Background())
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer conn.Close()
// 	rows, err = conn.QueryContext(context.Background(), q, args...)
// 	return
// }
