package Database

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	Conf "MavlinkProject/Server/backend/Config"
	Objects "MavlinkProject/Server/backend/Shared"
)

func InitMysql() (*gorm.DB, error) {
	setting := Conf.GetSetting()
	cfg := setting.Database.MySQL

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.Charset,
	)

	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(mysql.Open(dsn), config)
	if err != nil {
		return nil, err
	}

	mysqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	mysqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	mysqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	mysqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	if err := mysqlDB.Ping(); err != nil {
		return nil, err
	}

	log.Println("MavlinkProject - MysqlService : 自动化迁移开始...")
	for _, model := range Objects.ObjectModels {
		var target interface{} = model
		if ptr, ok := model.(interface{ GetModel() interface{} }); ok {
			target = ptr.GetModel()
		} else if reflect.TypeOf(model).Kind() == reflect.Ptr {
			target = reflect.ValueOf(model).Elem().Interface()
		}
		if err = db.AutoMigrate(target); err != nil {
			return nil, err
		}
		log.Printf("MavlinkProject - MysqlService : 表 %s 迁移完成", reflect.TypeOf(target).Name())
	}
	return db, nil
}
