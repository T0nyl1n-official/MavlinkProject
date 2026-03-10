package Database

import (
	"log"
	"os"
	"reflect"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	Objects "MavlinkProject/Server/backend/Shared"
)

func InitMysql() (*gorm.DB, error) {
	var err error
	dsn := os.Getenv("MavlinkMysqlDSN")
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// 尝试获取mysqlDB
	db, err := gorm.Open(mysql.Open(dsn), config)
	if err != nil {
		return nil, err
	}

	// 测试连通性
	mysqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err := mysqlDB.Ping(); err != nil {
		return nil, err
	}

	// 自动化迁移
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
