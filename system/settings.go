package system

import (
	"sync"

	"github.com/dougedey/elsinore/database"
	"gorm.io/gorm"
)

// Settings - The Global settings for this brewery
type Settings struct {
	gorm.Model
	BreweryName string
}

var once sync.Once
var (
	instance Settings
)

// CurrentSettings - Find or create the current system settings
func CurrentSettings() *Settings {
	once.Do(func() {
		instance = Settings{}
		instance.load()
	})

	return &instance
}

func (s *Settings) load() {
	database.FetchDatabase().Debug().First(s)
}

// Save the current settings
func (s *Settings) Save() {
	database.Save(s)
}
