package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type UserDatas struct {
	Data     int
	Category string
	Comment  string
	Date     int
}

type Db struct {
	db *sql.DB
}

func NewStorage(path string) (*Db, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("error open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connect db: %w", err)
	}
	return &Db{db: db}, nil
}

func (d *Db) Init() error {
	//qUsers := `CREATE TABLE IF NOT EXISTS users (userID INTEGER PRIMARY KEY, userName TEXT, regDate INTEGER)`
	//_, err := d.db.Exec(qUsers)
	//if err != nil {
	//	return fmt.Errorf("error create table users: %w", err)
	//}
	qData := `CREATE TABLE IF NOT EXISTS users (userID INTEGER PRIMARY KEY, userName TEXT, regDate INTEGER);
CREATE TABLE IF NOT EXISTS data (dataID INTEGER PRIMARY KEY AUTOINCREMENT, data INTEGER, category TEXT, comment TEXT, date INTEGER, user_id INTEGER, FOREIGN KEY (user_id)  REFERENCES users (userID));`
	_, err := d.db.Exec(qData)
	if err != nil {
		return fmt.Errorf("error create table data: %w", err)
	}
	return nil
}

func (d *Db) SaveUser(userId int64, regDate int, userName string) error {
	q := `INSERT OR IGNORE INTO users (userID, userName, regDate) VALUES (?,?,?)`
	_, err := d.db.Exec(q, userId, userName, regDate)
	if err != nil {
		return fmt.Errorf("error save user: %w", err)
	}
	//i, _ := res.LastInsertId()
	//log.Printf("User %s is saved %v", userName, i)
	return nil
}

func (d *Db) SaveData(data int, category, comment string, date int, userId int64) error {
	q := `PRAGMA foreign_keys = ON;
		  INSERT INTO data (data, category, comment, date, user_id) VALUES (?, ?, ?, ?, ?);`
	_, err := d.db.Exec(q, data, category, comment, date, userId)
	if err != nil {
		return fmt.Errorf("error save data: %w", err)
	}
	return nil
}

func (d *Db) GetSum(userId int64) ([]UserDatas, error) {
	var all []UserDatas
	qSum := `SELECT data,category,comment,date FROM data WHERE user_id = ?`

	rows, err := d.db.Query(qSum, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var userData UserDatas
		if err := rows.Scan(&userData.Data, &userData.Category, &userData.Comment, &userData.Date); err != nil {
			return nil, err
		}
		all = append(all, userData)
	}
	return all, nil
}

func (d *Db) CloseDB() {
	d.db.Close()
}
