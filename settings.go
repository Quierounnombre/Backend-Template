package main

import (
	"log"
	"os"
	"strings"

	g_jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-yaml"
)

func load_settings_from_env(s *Settings) {
	content, err := os.ReadFile(D_config_path)
	if err != nil {
		log.Fatalf("config: %s", err.Error())
	}
	err = yaml.Unmarshal(content, s)
	if err != nil {
		log.Fatalf("config: %s", err.Error())
	}
	s.Session_key = parse_string_env("SESSION_KEY")
	s.Jwt_priv_key = parse_string_env("JWT_PRIV_KEY")
	s.Jwt_priv_key = strings.ReplaceAll(s.Jwt_priv_key, `\n`, "\n")
	s.Jwt_pub_key = parse_string_env("JWT_PUB_KEY")
	s.Jwt_pub_key = strings.ReplaceAll(s.Jwt_pub_key, `\n`, "\n")
	s.DB_url = parse_string_env("DATABASE_URL")
	s.Client_id = parse_string_env("CLIENT_ID")
	s.Client_secret = parse_string_env("CLIENT_SECRET")
}

func parse_string_env(key string) (string) {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("config: %s: missing", key)
	}
	return v
}

func Set_cors_config(s *Settings) cors.Config {
	config := cors.DefaultConfig()
	config.AllowAllOrigins				= s.Cors.AllowAllOrigins
	config.AllowPrivateNetwork			= s.Cors.AllowPrivateNetwork
	config.AllowCredentials				= s.Cors.AllowCredentials
	config.AllowWildcard				= s.Cors.AllowWildcard
	config.AllowBrowserExtensions		= s.Cors.AllowBrowserExtensions
	config.AllowWebSockets				= s.Cors.AllowWebsockets
	config.AllowFiles					= s.Cors.AllowFiles
	config.AllowOrigins					= s.Cors.AllowOrigins
	config.AllowMethods					= s.Cors.AllowMethods
	config.AllowHeaders					= s.Cors.AllowHeaders
	config.MaxAge						= s.Cors.MaxAge
	config.OptionsResponseStatusCode	= s.Cors.OptionsResponseStatus
	return config
}

func Set_endpoints(
	s			*Settings,
	eng			*gin.Engine,
	db			*Db_data,
	handle		*g_jwt.GinJWTMiddleware,
) {
	eng.GET("/OAuthLogin", OAuthLogin(s))
	eng.GET("/OAuthCallback", OAuthCallback(s, db, handle))
	eng.POST("/auth/refresh", handle.RefreshHandler)
	eng.GET("/auth/public-key", Expose_pub_key(s))
	eng.NoRoute(handle.MiddlewareFunc(), handleNoRoute())
	auth := eng.Group("/user/", handle.MiddlewareFunc())
	{
		auth.GET("/profile", GetProfile(db))
		auth.GET("/logout", handle.LogoutHandler)
		auth.GET("/erase_user", EraseUser(db))
	}
}

func Set_JWT(s *Settings) *g_jwt.GinJWTMiddleware {
	Middleware, err := g_jwt.New(init_jwt_params(s))
	if err != nil {
		log.Fatalf("JWT Error:" + err.Error())
	}
	err = Middleware.MiddlewareInit()
	if err != nil {
		log.Fatalf("Middleware init Error:" + err.Error())
	}
	return Middleware
}

func Set_RateLimiter(s *Settings) *RateLimiter {
	rl := RateLimiter {
		reqs: make(map[string]uint),
		Max_reqs: s.Limiter.Max_request,
		Reset_time: s.Limiter.Reset_time,
	}
	go rl.Cleanup()
	return (&rl)
}

func Set_gin(s *Settings, db *Db_data) *gin.Engine {
	Middleware := Set_JWT(s)
	store := cookie.NewStore([]byte(s.Session_key))
	config := Set_cors_config(s)
	rl :=  Set_RateLimiter(s)
	eng := gin.New()
	eng.Use(gin.Recovery())
	eng.Use(gin.Logger())
	eng.Use(cors.New(config))
	eng.Use(rl.Middleware())
	eng.Use(sessions.Sessions("state_session", store))
	Set_endpoints(s, eng, db, Middleware)
	return eng
}
