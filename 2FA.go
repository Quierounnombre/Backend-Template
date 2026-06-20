package main

import (
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
)

func create_2FA_table(db *Db_data) {
	var sql		string
	var err		error

	sql = `
	CREATE TABLE IF NOT EXISTS 2fa_pending (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		time TIMESTAMPTZ DEFAULT NOW()
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
	DELETE FROM 2fa_pending WHERE created_at < NOW() - $1::interval
	`
	ticker := time.NewTicker(D_Reset_check_time)
	defer ticker.Stop()
	for range ticker.C {
		ctx, cancel := db.ctx()
		_, err := db.pool.Exec(ctx, sql, D_2FA_time.String())
		if err != nil {
			log.Fatalf("error at checking table: %s", err.Error())
		}
		cancel()
	}
}

func create_a_2FA(db *Db_data, email string) (string, error) {
	var sql		string
	var err		error

	if email == "" {
		return "", errors.New("email is empty")
	}
	sql = `
	INSERT INTO 2fa_pending (id, email) VALUES ($1, $2)
	`
	id := uuid.New().String()
	ctx, cancel := db.ctx()
	defer cancel()
	_, err = db.pool.Exec(ctx, sql, id, email)
	return id, err
}

func delete_a_2FA(db *Db_data, id string) error {
	var sql		string
	var err		error

	sql = `
	DELETE FROM 2fa_pending WHERE id=$1
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
	SELECT email FROM 2fa_pending WHERE id=$1
	`
	ctx, cancel := db.ctx()
	defer cancel()
	row := db.pool.QueryRow(ctx, sql, email)
	err = row.Scan(&email)
	return email, err
}
