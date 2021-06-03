package database

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var datastore *gorm.DB
var dbFile gorm.Dialector

// InitDatabase Create a db in the dbName path/filename, and migrate it with the supplied models
func InitDatabase(dbName *string, dst ...interface{}) {
	log.Info().Msgf("Loading database %v", *dbName)
	var err error
	dbFile = sqlite.Open(fmt.Sprintf("%v.db", *dbName))
	datastore, err = gorm.Open(dbFile, &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	err = datastore.AutoMigrate(dst...)
	if err != nil {
		log.Fatal().Err(err).Msg("Migration failed! Please check the logs!")
	}
}

// FetchDatabase Return the current database pointer
func FetchDatabase() *gorm.DB {
	return datastore
}

// Close Closes the underlying DB
func Close() {
	sqlDB, err := datastore.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("No DB to close!")
	}
	sqlDB.Close()
	datastore = nil
}

// Create a new instance of the supplied interface, this is a helper wrapper around the database so you don't need to check FetchDatabase
func Create(dst interface{}) {
	if datastore != nil {
		datastore.Create(dst)
	} else {
		log.Printf("No database configured, not creating %v", dst)
	}
}

// Save Update an instance of the supplied interface, this is a helper wrapper around the database so you don't need to check FetchDatabase
func Save(dst interface{}) {
	if datastore != nil {
		result := datastore.Session(&gorm.Session{FullSaveAssociations: true}).Debug().Save(dst)
		if result.Error != nil {
			log.Error().Err(result.Error).Msg("Failed to save")
		}
	} else {
		log.Info().Msgf("No database configured, not saving %v", dst)
	}
}
