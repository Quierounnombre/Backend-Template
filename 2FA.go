package main

import (
	"errors"
	"log"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

func create_2FA_table(db *Db_data) {
	var sql		string
	var err		error

	sql = `
	CREATE TABLE IF NOT EXISTS pending_2fa (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		name TEXT NOT NULL,
		password_hash TEXT,
		picture TEXT
	);`
	ctx, cancel := db.ctx()
	defer cancel()
	_, err = db.pool.Exec(ctx, sql)
	if err != nil {
		log.Fatalf("error creating table: %s", err.Error())
	}
	go cleanup_2FA_table(db)
}

func cleanup_2FA_table(db *Db_data) {
	var sql				string

	sql = `
	DELETE FROM pending_2fa WHERE created_at < NOW() - $1::interval
	`
	ticker := time.NewTicker(D_Reset_check_time)
	defer ticker.Stop()
	for range ticker.C {
		ctx, cancel := db.ctx()
		_, err := db.pool.Exec(ctx, sql, D_2FA_time.String())
		if err != nil {
			slog.Error("cleanup failed", "err", err)
		}
		slog.Info("Cleaned 2fa_pending")
		cancel()
	}
}

func Move_2FA_to_users(db *Db_data, id string) error {
	var sql		string
	var err		error
	var user	User
	var hash	string

	sql = `
	SELECT email, name, password_hash, picture FROM pending_2fa WHERE id=$1
	`
	ctx, cancel := db.ctx()
	defer cancel()
	row := db.pool.QueryRow(ctx, sql, id)
	err = row.Scan(&user.Email, &user.Name, &hash, &user.Picture)
	if err != nil {
		return err
	}
	err = AddUser(db, &user)
	if err != nil {
		return err
	}
	err = StorePassSimple(db, hash, user.Email, "users")
	if err != nil {
		return nil
	}
	return err
}

func create_a_2FA(db *Db_data, user *User, password string) (string, error) {
	var sql		string
	var err		error

	if user.Email == "" {
		return "", errors.New("email is empty")
	}
	sql = `
	INSERT INTO pending_2fa (id, email, name, picture) VALUES ($1, $2, $3, $4)
	`
	id := uuid.New().String()
	ctx, cancel := db.ctx()
	defer cancel()
	_, err = db.pool.Exec(ctx, sql, id, user.Email, user.Name, user.Picture)
	if err != nil {
		return "", err
	}
	err = StorePass(db, password, user.Email, "pending_2fa")
	return id, err
}

func delete_a_2FA(db *Db_data, id string) error {
	var sql		string
	var err		error

	sql = `
	DELETE FROM pending_2fa WHERE id=$1
	`
	ctx, cancel := db.ctx()
	defer cancel()
	_, err = db.pool.Exec(ctx, sql, id)
	return err
}

func Get2FA(db *Db_data, id string) (string, error) {
	var err		error
	var sql		string
	var email	string

	sql = `
	SELECT email FROM pending_2fa WHERE id=$1
	`
	ctx, cancel := db.ctx()
	defer cancel()
	row := db.pool.QueryRow(ctx, sql, email)
	err = row.Scan(&email)
	return email, err
}
