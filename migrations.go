package gochat

import (
	"database/sql"
	"log"
	"os"
	"path"
	"time"
)

func RunMigrations(db *sql.DB) error {
	_, err := db.Exec(`
create table if not exists  migrations (
    filename text not null,
    timestamp int not null,
    primary key (filename)
);`)

	if err != nil {
		return err
	}

	contents, err := os.ReadDir("./migrations")
	if err != nil {
		return err
	}

	for _, content := range contents {
		if content.IsDir() {
			continue
		}

		ext := path.Ext(content.Name())
		if ext != ".sql" {
			continue
		}

		exists := false
		err := db.QueryRow("select exists (select 1 from migrations where filename = ?)", content.Name()).Scan(&exists)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		log.Println("running migration", content.Name())

		sql, err := os.ReadFile("./migrations/" + content.Name())
		if err != nil {
			continue
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		_, err = tx.Exec(string(sql))
		if err != nil {
			return err
		}

		_, err = tx.Exec("insert into migrations (filename, timestamp) values (?, ?)", content.Name(), time.Now().UnixMilli())
		if err != nil {
			return err
		}

		tx.Commit()
	}

	return nil
}
