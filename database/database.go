package database

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var datastore *gorm.DB

// InitDatabase Create a db in the dbName path/filename, and migrate it with the supplied models
func InitDatabase(dbName *string, dst ...interface{}) {
	fmt.Println("Loading database")
	var err error
	datastore, err = gorm.Open(sqlite.Open(fmt.Sprintf("%v.db", *dbName)), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	err = datastore.AutoMigrate(dst...)
	if err != nil {
		log.Panicf("Migration failed! Please check the logs! %v", err)
	}
}

// FetchDatabase Return the current database pointer
func FetchDatabase() *gorm.DB {
	return datastore
}

// Create a new instance of the supplied interface, this is a helper wrapper around the database so you don't need to check FetchDatabase
func Create(dst interface{}) {
	if datastore != nil {
		datastore.Create(dst)
	} else {
		log.Printf("No database configured, not creeating %v", dst)
	}
}

// Save Update an instance of the supplied interface, this is a helper wrapper around the database so you don't need to check FetchDatabase
func Save(dst interface{}) {
	if datastore != nil {
		datastore.Save(dst)
	} else {
		log.Printf("No database configured, not saving %v", dst)
	}
}