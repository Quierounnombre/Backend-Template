package main

//Manage Auth process(OAuth only atm) and endpoints

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	g_jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/appleboy/gin-jwt/v3/core"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
)

//JWT MANUAL: https://pkg.go.dev/github.com/golang-jwt/jwt/v5#section-documentation
//HOLY BIBLE: https://developers.google.com/identity/protocols/oauth2/web-server#httprest_1

//ADD STATE TOKEN
func generate_state_token() string {
	state := rand.Text()
	return (state)
}

func oauthlogin_url_with_query(s *Settings, state string) (*url.URL, error) {
	u, err := url.Parse(s.OAuth.Provider)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("client_id", s.Client_id)
	q.Set("redirect_uri", s.OAuth.Redirect_uri)
	q.Set("response_type", "code")
	q.Set("scope", "openid email profile")
	q.Set("state", state)
	q.Set("access_type", "online")
	u.RawQuery = q.Encode()
	return u, nil
}

func OAuthLogin(s *Settings) gin.HandlerFunc {
	return func(c *gin.Context) {
		oauth_cookies := sessions.Default(c)
		state := generate_state_token()
		oauth_cookies.Set("state", state)
		err := oauth_cookies.Save()
		if err != nil {
			c.JSON(400, gin.H{"error saving cookies": err.Error()})
			return
		}
		auth_url, err := oauthlogin_url_with_query(s, state)
		if err != nil {
			c.JSON(400, gin.H{"Error obtaining url": err.Error()})
		}
		c.Redirect(307, auth_url.String())
	}
}

func oauthcallback_url_with_query(s *Settings, code string) (*url.URL, error) {
	u, err := url.Parse(s.OAuth.Token_provider)
	if err != nil {
		return u, err
	}
	q := u.Query()
	q.Set("client_id", s.Client_id)
	q.Set("client_secret", s.Client_secret)
	q.Set("code", code)
	q.Set("grant_type", "authorization_code")
	q.Set("redirect_uri", s.OAuth.Redirect_uri)
	u.RawQuery = q.Encode()
	return u, nil
}

func OAuthCallback(s *Settings, db *Db_data, authMiddleware *g_jwt.GinJWTMiddleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		var state			string
		var response_state	string
		var response_code	string
		var id_token		string
		var err				error
		var ok				bool
		var oauth_cookies	sessions.Session
		var token_url		*url.URL
		var resp			*http.Response
		var body			map[string]interface{}
		var decoder			*json.Decoder
		var parser			jwt.Parser
		var claims			jwt.MapClaims
		var jwt_token		*jwt.Token
		var user			*User
		var storage_data	User

		oauth_cookies = sessions.Default(c)
		state, ok = oauth_cookies.Get("state").(string)
		if !ok {
			c.JSON(400, gin.H{"error in assertion": err.Error()})
			return
		}
    	oauth_cookies.Delete("state")
		err = oauth_cookies.Save()
		if err != nil {
			c.JSON(400, gin.H{"error saving cookies": err.Error()})
			return
		}
		response_state = c.Query("state")
		if response_state == "" {
			c.JSON(401, gin.H{"error": "missing state"})
			return
		}
		if state != response_state {
			c.JSON(400, gin.H{"error": "mismatch state token"})
			return
		}
		response_code = c.Query("code")
		if response_code == "" {
			c.JSON(400, gin.H{"error": "Missing auth code"})
			return
		}
		token_url, err = oauthcallback_url_with_query(s, response_code)
		if err != nil {
			c.JSON(400, gin.H{"error in response": err.Error()})
			return
		}
		resp, err = http.Post(token_url.String(), "", nil)
		if err != nil {
			c.JSON(400, gin.H{"error in response": err.Error()})
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			c.JSON(400, gin.H{"error": "bad status code"})
			return
		}
		decoder = json.NewDecoder(resp.Body)
		err = decoder.Decode(&body)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		id_token, ok = body["id_token"].(string)
		if !ok {
			c.JSON(400, gin.H{"error": "Missing id token"})
			return
		}
		jwt_token, _, err = parser.ParseUnverified(id_token, jwt.MapClaims{})
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		claims, ok = jwt_token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(400, gin.H{"error": "Missing claims"})
			return
		}
		storage_data.Name = fmt.Sprintf("%v", claims["given_name"])
		storage_data.Email = fmt.Sprintf("%v", claims["email"])
		storage_data.Picture = fmt.Sprintf("%v", claims["picture"])
		user, err = Login_or_ADD_User(db, &storage_data)
		if err != nil {
			c.JSON(400, gin.H{"Error creating user": err.Error()})
			return
		}
		err = handleOAuthSuccess(c, authMiddleware, user, "google")
		if err != nil {
			c.JSON(400, gin.H{"Error authenticating user": err.Error()})
			return
		}
	}
}

func handleOAuthSuccess(
	c				*gin.Context,
	authMiddleware	*g_jwt.GinJWTMiddleware,
	user			*User,
	provider		string,
) error {
	var token		*core.Token
	var err			error

	c.Set(authMiddleware.IdentityKey, user)
	c.Set("oauth_provider", provider)
	token, err = authMiddleware.TokenGenerator(c.Request.Context(), user)
	if err != nil {
		return err
	}
	authMiddleware.SetCookie(c, token.AccessToken)
	authMiddleware.SetRefreshTokenCookie(c, token.RefreshToken)
	if authMiddleware.LoginResponse != nil {
		authMiddleware.LoginResponse(c, token)
	}
	slog.Info("oauth login", "provider", "google", "user_id", user.UserID)
	return nil
}

func Expose_pub_key(s *Settings) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{"pub_key": s.Jwt_pub_key})
	}
}

func Pass_Auth(
	db				*Db_data,
	authMiddleware	*g_jwt.GinJWTMiddleware,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req		LoginRequest
		var err		error
		var match	bool
		var token	*core.Token
		var user	*User

		err = c.ShouldBindJSON(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		if req.Email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing email"})
			return
		}
		match, err = CheckUserPassword(db, req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error checking password"})
			return
		}
		if match == false {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password don't match"})
			return
		}
		user, err = GetUser(db, req.Email)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.Set(authMiddleware.IdentityKey, user)
		token, err = authMiddleware.TokenGenerator(c.Request.Context(), user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		authMiddleware.SetCookie(c, token.AccessToken)
		authMiddleware.SetRefreshTokenCookie(c, token.RefreshToken)
		if authMiddleware.LoginResponse != nil {
			authMiddleware.LoginResponse(c, token)
		}
		slog.Info("password login", "user_id", user.UserID)
	}
}

func Pass_Singup(
	s				*Settings,
	db				*Db_data,
	authMiddleware	*g_jwt.GinJWTMiddleware,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req		SignUpRequest
		var err		error
		var user	User

		err = c.ShouldBindJSON(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		if req.Email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing email"})
			return
		}
		if req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing name"})
			return
		}
		if req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing password"})
			return
		}
		user.Email = req.Email
		user.Name = req.Name
		id, err := create_a_2FA(db, &user, req.Password)
		if err != nil {
			slog.Warn("2FA creating user", "err", err)
			c.JSON(500, gin.H{"Error:": " Error in 2FA"})
			return
		}
		err = TwoFA_Mail(s, db, req.Email, id)
		if err != nil {
			slog.Warn("2FA sending email", "err", err)
			c.JSON(500, gin.H{"Error:": " Error in 2FA"})
			return
		}
		c.JSON(200, gin.H{"result": "Check your email"})
	}
}

func Validate_2FA(
	s					*Settings,
	db					*Db_data,
	authMiddleware		*g_jwt.GinJWTMiddleware,
) gin.HandlerFunc {
	return func (c *gin.Context) {
		var err			error
		var db_email	string
		
		id := c.Param("id")
		db_email, err = Get2FA(db, id)
		if err != nil {
			slog.Warn("2FA getting 2fa_pending", "err", err)
			c.JSON(500, gin.H{"Error:": " Error in 2FA"})
			return 
		}
		_, err = GetUser(db, db_email)
		if err != pgx.ErrNoRows {
			slog.Warn("2FA user may exists", "err", err)
			c.JSON(500, gin.H{"Error:": " Error in 2FA"})
			return
		}
		err = Move_2FA_to_users(db, id)
		if err != nil {
			slog.Warn("2FA moving users", "err", err)
			c.JSON(500, gin.H{"Error:": " Error in 2FA"})
			return
		}
		err = delete_a_2FA(db, id)
		if err != nil {
			slog.Warn("2FA deleting users", "err", err)
			c.JSON(500, gin.H{"Error:": " Error in 2FA"})
			return
		}
		user, err := GetUser(db, db_email)
		if err != nil {
			slog.Warn("2FA getting user", "err", err)
			c.JSON(500, gin.H{"Error:": " Error in 2FA"})
			return
		}
		c.Set(authMiddleware.IdentityKey, user)
		token, err := authMiddleware.TokenGenerator(c.Request.Context(), user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		authMiddleware.SetCookie(c, token.AccessToken)
		authMiddleware.SetRefreshTokenCookie(c, token.RefreshToken)
		if authMiddleware.LoginResponse != nil {
			authMiddleware.LoginResponse(c, token)
		}
		slog.Info("user registered", "user_id", user.UserID)
	}
}
