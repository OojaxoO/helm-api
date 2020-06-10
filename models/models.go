package models

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"

	"helm-api/setting"
)

var db *gorm.DB

func Setup () {
	var err error
	db, err = gorm.Open(setting.DatabaseSetting.Type, fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
    		    setting.DatabaseSetting.User,
    		    setting.DatabaseSetting.Password,
    		    setting.DatabaseSetting.Host,
				setting.DatabaseSetting.Name))

	if err != nil {
      	fmt.Println(err)
      	return
  	}else {
  	    fmt.Println("connection succedssed")
	}
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return setting.DatabaseSetting.TablePrefix + defaultTableName
	}

	db.SingularTable(true)
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
}

func CloseDB() {
	defer db.Close()
}
