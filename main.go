package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	var db		Db_data
	var eng		*gin.Engine
	var set		Settings


	load_settings_from_env(&set)
	set_logger(&set)
	gin.SetMode(set.Release_mode) //Needs to be before creating a engine
	Set_db(&set, &db)
	defer db.pool.Close()
	Set_db_tables(&db)
	eng = Set_gin(&set, &db)
	eng.Run(set.Port)
}
