package main

import (
	"time"
)

const D_config_path				= "config.yaml"
const D_JWT_identity_key		= "email"
const D_User_ID					= "user_id"
const D_Reset_pass_time			= 5 * time.Minute
const D_Reset_check_time		= 5 * time.Minute
const D_2FA_time				= 5 * time.Minute
