package database

import "database/sql"

func ConnectDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func CreateTable(db *sql.DB) error {
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS health (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		usuario TEXT NOT NULL,
		sentimento TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS sensorial (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		usuario TEXT NOT NULL,
		descricao TEXT NOT NULL,
		gatilho TEXT NOT NULL,
		nivel INTEGER NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.Exec(sqlStmt)
	return err
}

func InsertSensorialData(db *sql.DB, usuario string, descricao string, gatilho string, nivel string) error {
	stmt, err := db.Prepare("INSERT INTO sensorial(usuario, descricao, gatilho, nivel) VALUES(?, ?, ? ,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(usuario, descricao, gatilho, nivel)
	return err
}

func InsertHealthData(db *sql.DB, usuario string, sentimento string) error {
	stmt, err := db.Prepare("INSERT INTO health(usuario, sentimento) VALUES(?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(usuario, sentimento)
	return err
}
