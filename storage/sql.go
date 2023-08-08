package storage

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
												comment TEXT, 
												timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
												user_id INTEGER, 
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

func (d *Db) SaveData(data int, category, comment string, userId int64) error {
	q := `PRAGMA foreign_keys = ON;
		  INSERT INTO data (data, category, comment, user_id) VALUES (?, ?, ?, ?);`
	_, err := d.db.Exec(q, data, category, comment, userId)
	if err != nil {
		return fmt.Errorf("error save data: %w", err)
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

func (d *Db) GetUserInfo(userId int64) (Users, int, error) {
	var (
		user  Users
		count int
	)
	qUser := `SELECT * FROM users WHERE userID = ?`
	qCount := `SELECT count(data) FROM data WHERE user_id = ?`

	rows, err := d.db.Query(qUser, userId)
	if err != nil {
		return Users{}, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&user.UserId, &user.UserName, &user.RegDate); err != nil {
			return Users{}, 0, err
		}
	}

	rows, err = d.db.Query(qCount, userId)
	if err != nil {
		return Users{}, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return Users{}, 0, err
		}
	}
	return user, count, nil
}

func (d *Db) DeleteUserData(userId int64) error {
	q := `DELETE FROM data WHERE user_id=?`
	_, err := d.db.Exec(q, userId)
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
