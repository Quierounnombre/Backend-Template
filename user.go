package main

//Manage user endpoints(creation, deletion, list)

//TODO: A way to update user changes

import (
	"log"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	g_jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/golang-jwt/jwt/v5"
)

func create_table_user(db *Db_data) {
	var err		error
	var sql		string
	var ctx		context.Context
	var cancel	context.CancelFunc

	sql = `
	CREATE EXTENSION IF NOT EXISTS pgcrypto;

	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		picture TEXT
	);`
	ctx, cancel = db.ctx()
	defer cancel()
	_, err = db.pool.Exec(ctx, sql)
	if err != nil {
		log.Fatalf("error creating table: %s", err.Error())
	}
}

func AddUser(db *Db_data, user *User) error {
	var err		error
	var sql		string
	var ctx		context.Context
	var cancel	context.CancelFunc

	sql = `
	INSERT INTO users (name, email, picture) VALUES ($1, $2, $3)
	`
	ctx, cancel = db.ctx()
	defer cancel()
	_, err = db.pool.Exec(ctx, sql, user.Name, user.Email, user.Picture)
	return err
}

func GetUser(db *Db_data, email string) (*User, error) {
	var err		error
	var sql		string
	var ctx		context.Context
	var cancel	context.CancelFunc
	var row		pgx.Row
	var user	*User

	user = new(User)
	sql = `
	SELECT name, email, id, picture FROM users WHERE email=$1
	`
	ctx, cancel = db.ctx()
	defer cancel()
	row = db.pool.QueryRow(ctx, sql, email)
	err = row.Scan(&user.Name, &user.Email, &user.UserID, &user.Picture)
	return user, err
}

func CheckUserPassword(db *Db_data, req LoginRequest) (bool, error) {
	var match		bool
	var err			error
	var sql			string
	var ctx			context.Context
	var cancel		context.CancelFunc
	var row			pgx.Row

	sql = `
	SELECT password_hash = crypt(1$, password_hash) FROM users WHERE email=2$
	`
	ctx, cancel = db.ctx()
	defer cancel()
	row = db.pool.QueryRow(ctx, sql, req.Password, req.Email)
	err = row.Scan(&match)
	return match, err
}

func Login_or_ADD_User(db *Db_data, storage_data *User) (*User, error) {
	var tmp_email	string
	var sql			string
	var err			error
	var row			pgx.Row
	var ctx			context.Context
	var cancel		context.CancelFunc
	var user		*User

	sql = `
	SELECT email FROM users WHERE email=$1
	`
	ctx, cancel = db.ctx()
	defer cancel()
	row = db.pool.QueryRow(ctx, sql, storage_data.Email)
	err = row.Scan(&tmp_email)
	if err == pgx.ErrNoRows {
		err = nil
		err = AddUser(db, storage_data)
	} else if err != nil {
		return user, err
	}
	user, err = GetUser(db, storage_data.Email)
	return user, err
}

func EraseUser(db *Db_data) gin.HandlerFunc {
	return func(c *gin.Context) {
		var claims	jwt.MapClaims
		var sql		string
		var email	string
		var err		error
		var ctx		context.Context
		var cancel	context.CancelFunc

		sql = `
		DELETE FROM users WHERE email=$1
		`
		claims = g_jwt.ExtractClaims(c)
		email = claims["email"].(string)
		if email == "" {
			c.JSON(400, gin.H{"Error:": " retrieving claims from jwt"})
			return
		}
		ctx, cancel = db.ctx()
		defer cancel()
		_, err = db.pool.Exec(ctx, sql, email)
		if err != nil {
			c.JSON(400, gin.H{"Error deleting user:": err.Error()})
			return
		}
		c.JSON(200, gin.H{})
	}
}

func GetProfile(db *Db_data) gin.HandlerFunc {
	return func(c *gin.Context) {
		var claims		jwt.MapClaims
		var err			error
		var email		string
		var user		*User

		claims = g_jwt.ExtractClaims(c)
		email = claims["email"].(string)
		if email == "" {
			c.JSON(400, gin.H{"Error:": " retrieving claims from jwt"})
			return
		}
		user, err = GetUser(db, email)
		if err != nil {
			c.JSON(400, gin.H{"Error:": " obtaining user from db"})
			return
		}
		c.JSON(200, gin.H{
			"email": user.Email,
			"name": user.Name,
			"picture": user.Picture,
		})
	}
}

func ResetPass(s *Settings) gin.HandlerFunc {
	return func (c *gin.Context) {
		var err		error

		err = Mail_Reset_Pass(s, "vicengandrade@gmail.com")
		if err != nil {
			c.JSON(500, gin.H{"Error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"result": "Check your email"})
	}
}
