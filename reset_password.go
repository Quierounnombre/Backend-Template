package main

import (
	"log"
	"time"
	"github.com/google/uuid"
)

func create_password_reset_table(db *Db_data) {
	var sql		string
	var err		error

	sql = `
	CREATE TABLE IF NOT EXISTS reset_pass (
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
	go cleanup_password_reset_table(db)
}

func cleanup_password_reset_table(db *Db_data) {
	var sql				string

	sql = `
	DELETE FROM reset_pass WHERE created_at < NOW() - $1::interval
	`
	ticker := time.NewTicker(D_Reset_check_time)
	defer ticker.Stop()
	for range ticker.C {
		ctx, cancel := db.ctx()
		_, err := db.pool.Exec(ctx, sql, D_Reset_pass_time.String())
		if err != nil {
			log.Fatalf("error at checking table: %s", err.Error())
		}
		cancel()
	}
}

func create_a_password_reset(db *Db_data, email string) (string, error) {
	var sql		string
	var err		error

	sql = `
	INSERT INTO reset_pass (id, email) VALUES ($1, $2)
	`
	id := uuid.New().String()
	ctx, cancel := db.ctx()
	defer cancel()
	_, err = db.pool.Exec(ctx, sql, id, email)
	return id, err
}

func delete_a_password_reset(db *Db_data, id string) error {
	var sql		string
	var err		error

	sql = `
	DELETE FROM users WHERE id=$1
	`
	ctx, cancel := db.ctx()
	defer cancel()
	_, err = db.pool.Exec(ctx, sql, id)
	return err
}
