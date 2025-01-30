package internal

import (
	"gorm.io/gorm"
)

type Genre struct {
	gorm.Model
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Platform struct {
	gorm.Model
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Game struct {
	gorm.Model
	ID          int    `json:"id"`
	Name        string `json:"name"`
	SteamAppID  string `json:"steam_app_id"`
	ReleaseYear int    `json:"release_year"`
	ReviewScore int    `json:"review_score"`
}

type GameGenre struct {
	gorm.Model
	ID      int `json:"id"`
	GameID  int `json:"game_id"`
	GenreID int `json:"genre_id"`

	Game  Game
	Genre Genre
}

type GamePlatform struct {
	gorm.Model
	ID         int `json:"id"`
	GameID     int `json:"game_id"`
	PlatformID int `json:"platform_id"`
	TimeToBeat int `json:"time_to_beat" comment:"in minutes"`

	Game     Game
	Platform Platform
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&Genre{}, &Platform{}, &Game{}, &GameGenre{}, &GamePlatform{})
}
