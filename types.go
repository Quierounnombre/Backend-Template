package main

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"
	"context"
	"time"
)

//----------------------------------------------------------------------------------------------SETTINGS

/*
github.com/jackc/pgx/v5/pgxpool

DB config for pgxpool
*/
type Db_settings struct {
	Max_cons					int32				`yaml:"max_cons"`
	Min_cons					int32				`yaml:"min_cons"`
	Min_idle_cons				int32				`yaml:"min_idle_cons"`
	Health_check_period			time.Duration		`yaml:"health_check_period"`
	Max_con_lifetime			time.Duration		`yaml:"max_con_lifetime"`
	Max_con_life_time_jitter	time.Duration		`yaml:"max_con_lifetime_jitter"`
	Max_con_idle_time			time.Duration		`yaml:"max_con_idle_time"`
	Ctx_timeout					time.Duration		`yaml:"ctx_timeout"`
}

/*
https://pkg.go.dev/github.com/gin-contrib/cors#Config

Cors config
*/
type Cors_settings struct {
	AllowAllOrigins			bool			`yaml:"allow_all_origins"`
	AllowPrivateNetwork		bool			`yaml:"allow_private_network"`
	AllowCredentials        bool			`yaml:"allow_credentials"`
	AllowWildcard			bool			`yaml:"allow_wildcard"`
	AllowBrowserExtensions	bool			`yaml:"allow_browser_extensions"`
	AllowWebsockets			bool			`yaml:"allow_websockets"`
	AllowFiles				bool			`yaml:"allow_files"`
	AllowOrigins			[]string		`yaml:"allow_origins"`
	AllowMethods			[]string		`yaml:"allow_methods"`
	AllowHeaders			[]string		`yaml:"allow_headers"`
	MaxAge					time.Duration	`yaml:"max_age"`
	OptionsResponseStatus	int				`yaml:"options_response_status"`
}

/*
"github.com/appleboy/gin-jwt/v3"
"github.com/appleboy/gin-jwt/v3/core"
"github.com/golang-jwt/jwt/v5"

JWT configs
*/
type Jwt_settings struct {
	Realm				string				`yaml:"realm"`
	TokenLookup			string				`yaml:"token_lookup"`
	TokenHeadName		string				`yaml:"token_head_name"`
	CookieDomain		string				`yaml:"cookie_domain"`
	SendCookie			bool				`yaml:"send_cookie"`
	SecureCookie		bool				`yaml:"secure_cookie"`
	SendAuthorization	bool				`yaml:"send_authorization"`
	CookieHTTPOnly		bool				`yaml:"cookie_http_only"`
	CookieSameSite		int					`yaml:"cookie_same_site"`
	CookieMaxAge		time.Duration		`yaml:"cookie_max_age"`
	Timeout				time.Duration		`yaml:"timeout"`
	MaxRefresh			time.Duration		`yaml:"max_refresh"`
}

type OAuth_settings struct {
	Provider			string				`yaml:"provider"`
	Redirect_uri		string				`yaml:"redirect_uri"`
	Token_provider		string				`yaml:"token_provider"`
}

type Settings struct {
	Release_mode		string				`yaml:"release_mode"`
	Frontend			string				`yaml:"frontend"`
	Port				string				`yaml:"port"`
	Db_set				Db_settings			`yaml:"db"`
	Cors				Cors_settings		`yaml:"cors"`
	Jwt					Jwt_settings		`yaml:"jwt"`
	OAuth				OAuth_settings		`yaml:"oauth"`

	// injected from .env

	Session_key			string
	Jwt_priv_key		string
	Jwt_pub_key			string
	DB_url				string
	Client_id			string
	Client_secret		string
}

//------------------------------------------------------------------------------------------------------DATA

type Db_data struct {
	pool			*pgxpool.Pool
	cancel			context.CancelFunc
	ctx_timeout		time.Duration
}

type User struct {
	Name			string		`json:"name"`
	Email			string		`json:"email"`
	UserID			uuid.UUID	`json:"id"`
	Picture			string		`json:"picture"`
}
