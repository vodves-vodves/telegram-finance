package sql

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type UserDatas struct {
	Data     int
	Category string
	Comment  string
	Date     time.Time
}

type Users struct {
	UserId   int64
	UserName string
	RegDate  int
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
	qData := `CREATE TABLE IF NOT EXISTS users (userID INTEGER PRIMARY KEY, userName TEXT, regDate INTEGER);
              CREATE TABLE IF NOT EXISTS data (dataID INTEGER PRIMARY KEY AUTOINCREMENT, 
												data INTEGER, 
												category TEXT, 
												comment TEXT NOT NULL , 
												timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
												user_id INTEGER NOT NULL, 
												FOREIGN KEY (user_id)  REFERENCES users (userID));
			  CREATE TABLE IF NOT EXISTS credits (creditID INTEGER PRIMARY KEY AUTOINCREMENT, 
												data INTEGER,
												from_id INTEGER NOT NULL ,
												comment TEXT NOT NULL , 
												active INTEGER NOT NULL ,
												timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
												user_id INTEGER NOT NULL, 
												FOREIGN KEY (user_id)  REFERENCES users (userID));`
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
	return nil
}

func (d *Db) SaveData(data int, category string, userId int64) (int64, error) {
	q := `PRAGMA foreign_keys = ON;
		  INSERT INTO data (data, category,comment, user_id) VALUES (?, ?, ?, ?);`
	res, err := d.db.Exec(q, data, category, "", userId)
	if err != nil {
		return 0, fmt.Errorf("error save data: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error get last id data: %w", err)
	}
	return id, nil
}

func (d *Db) SetComment(comment string, dataID int64) error {
	q := `UPDATE data SET comment = ? WHERE dataID = ?`
	_, err := d.db.Exec(q, comment, dataID)
	if err != nil {
		return fmt.Errorf("error edit comment data: %w", err)
	}
	return nil
}

func (d *Db) GetSum(userId int64, year int, month time.Month) ([]UserDatas, error) {
	var all []UserDatas
	qSum := `SELECT data, category, comment, timestamp FROM data WHERE user_id = ? AND timestamp BETWEEN ? AND ?`

	monthStart := time.Date(year, month, 0, 0, 0, 0, 0, time.Local)
	monthEnd := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local)

	rows, err := d.db.Query(qSum, userId, monthStart, monthEnd)
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

func (d *Db) AllSum(userId int64) (int, error) {
	var allSum int
	qSum := `SELECT sum(data) FROM data WHERE user_id = ?`

	rows, err := d.db.Query(qSum, userId)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&allSum); err != nil {
			return 0, err
		}
	}

	return allSum, nil
}

func (d *Db) GetUsers() ([]Users, error) {
	var all []Users
	q := `SELECT * FROM users`

	rows, err := d.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var userData Users
		if err := rows.Scan(&userData.UserId, &userData.UserName, &userData.RegDate); err != nil {
			return nil, err
		}
		all = append(all, userData)
	}

	return all, nil
}

func (d *Db) DeleteAllUserData(userId int64) error {
	q := `DELETE FROM data WHERE user_id=?`
	_, err := d.db.Exec(q, userId)
	if err != nil {
		return fmt.Errorf("error delete all user data: %w", err)
	}
	return nil
}

func (d *Db) DeleteUserData(recordId int64) error {
	q := `DELETE FROM data WHERE dataID=?`
	_, err := d.db.Exec(q, recordId)
	if err != nil {
		return fmt.Errorf("error delete user data: %w", err)
	}
	return nil
}

func (d *Db) DeleteUser(userId int64) error {
	q := `DELETE FROM users WHERE userID=?`
	_, err := d.db.Exec(q, userId)
	if err != nil {
		return fmt.Errorf("error delete user: %w", err)
	}
	return nil
}
