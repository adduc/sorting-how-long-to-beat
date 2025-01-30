/*
Synopsis:

	Script to import data from jsonlines file to SQLite database.

Example Data:

	{"Name": "Borderlands 3", "Stats": {"Additional Content": {"Moxxi's Heist of the Handsome Jackpot": {"Polled": "77%", "Rated": "6h", "Main": "9h", "Main+": "11h", "100%": "9h"}, "Guns, Love, and Tentacles": {"Polled": "74%", "Rated": "6h", "Main": "10h", "Main+": "11h", "100%": "10h"}, "Bounty of Blood": {"Polled": "74%", "Rated": "5h", "Main": "7h", "Main+": "11h", "100%": "8h"}, "Psycho Krieg and the Fantastic Fustercluck": {"Polled": "71%", "Rated": "4h", "Main": "5h", "Main+": "6h", "100%": "5h"}}, "Single-Player": {"Main Story": {"Polled": "436", "Average": "23h 17m", "Median": "22h 20m", "Rushed": "16h 57m", "Leisure": "31h 55m"}, "Main + Extras": {"Polled": "769", "Average": "47h 3m", "Median": "40h 2m", "Rushed": "30h 46m", "Leisure": "270h 46m"}, "Completionist": {"Polled": "175", "Average": "77h 27m", "Median": "66h", "Rushed": "54h 58m", "Leisure": "121h 51m"}, "All PlayStyles": {"Polled": "1.4K", "Average": "43h 24m", "Median": "35h", "Rushed": "25h 9m", "Leisure": "268h 36m"}}, "Multi-Player": {"Co-Op": {"Polled": "90", "Average": "43\u00bd Hours", "Median": "38 Hours", "Least": "21\u00bd Hours", "Most": "81 Hours"}, "Competitive": {"Polled": "1", "Average": "19\u00bd Hours", "Median": "19\u00bd Hours", "Least": "19\u00bd Hours", "Most": "19\u00bd Hours"}}, "Platform": {"Google Stadia": {"Polled": "12", "Main": "28h 40m", "Main +": "65h 40m", "100%": "133h 1m", "Fastest": "25h", "Slowest": "158h"}, "PC": {"Polled": "926", "Main": "22h 57m", "Main +": "46h 14m", "100%": "83h 11m", "Fastest": "11h 6m", "Slowest": "498h"}, "PlayStation 4": {"Polled": "248", "Main": "23h 52m", "Main +": "46h 5m", "100%": "65h 39m", "Fastest": "11h 17m", "Slowest": "250h"}, "PlayStation 5": {"Polled": "50", "Main": "25h 32m", "Main +": "46h 18m", "100%": "61h 13m", "Fastest": "17h 28m", "Slowest": "113h"}, "Xbox One": {"Polled": "116", "Main": "23h 44m", "Main +": "51h 58m", "100%": "73h 22m", "Fastest": "12h 3m", "Slowest": "413h"}, "Xbox Series X/S": {"Polled": "28", "Main": "22h 47m", "Main +": "56h 48m", "100%": "124h 1m", "Fastest": "14h 12m", "Slowest": "172h"}}}, "steam_app_id": "397540", "Release_date": "2019-09-13", "Genres": "First-Person, Action, Shooter", "Review_score": 76}

Table Structures:
  - genres (id, name)
  - platforms (id, name)
  - games (id, name, steam_app_id, release_date, review_score)
  - game_genres (id, game_id, genre_id)
  - game_platforms (id, game_id, platform_id, time_to_beat)

Constraints:
  - gorm should be used to interact with the database

Opportunities for Improvement:
  - Batch process 1k lines at a time to reduce the number of queries
  - e.g. build slices of Game, Genre, Platform, GameGenre, and
    GamePlatform structs, insert games, genres, platforms, and then
    reconcile and insert game_genres and game_platforms
*/
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/adduc/sorting-how-long-to-beat/internal"
)

type GameData struct {
	Name        string `json:"Name"`
	SteamAppID  string `json:"steam_app_id"`
	ReleaseDate string `json:"Release_date"`
	ReviewScore int    `json:"Review_score"`
	Genres      string `json:"Genres"`
	Stats       struct {
		Platform map[string]map[string]string `json:"Platform"`
	} `json:"Stats"`
}

func parseFilePath() *string {
	// Parse command line arguments
	filePath := flag.String("file", "", "Path to the jsonlines file")
	flag.Parse()

	if *filePath == "" {
		log.Fatalln("Please provide the path to the jsonlines file using the -file flag.")
	}

	return filePath
}

func openFile(filePath string) *os.File {
	// Open the jsonlines file
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalln("Error opening file:", err)
	}

	return file
}

func openDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("games.db"), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		log.Fatalln("Error connecting to database:", err)
	}

	// Enable WAL mode
	db.Exec("PRAGMA journal_mode = WAL;")

	// Set busy_timeout to 5 seconds
	db.Exec("PRAGMA busy_timeout = 5000;")

	// Set sync mode to NORMAL
	db.Exec("PRAGMA synchronous = NORMAL;")

	return db
}

func processLine(scanner *bufio.Scanner, db *gorm.DB) {

	var gameData GameData
	if err := json.Unmarshal(scanner.Bytes(), &gameData); err != nil {
		fmt.Println("Error unmarshalling json:", err)
		return
	}

	// Take first 4 characters of the release date to get the year, and convert it to an integer
	releaseYearStr := strings.Split(gameData.ReleaseDate, "-")[0]
	releaseYear := 0

	if releaseYearStr != "" {
		val, err := strconv.Atoi(releaseYearStr)
		releaseYear = val
		if err != nil {
			fmt.Println("Error converting release year to integer:", err)
			fmt.Println("Game:", gameData.Name, "Year:", releaseYearStr)
		}
	}

	// Create or get game
	game := internal.Game{
		Name:        gameData.Name,
		SteamAppID:  gameData.SteamAppID,
		ReleaseYear: releaseYear,
		ReviewScore: gameData.ReviewScore,
	}
	db.FirstOrCreate(&game, internal.Game{Name: gameData.Name})

	// Handle genres
	genres := strings.Split(gameData.Genres, ", ")
	for _, genreName := range genres {
		var genre internal.Genre
		db.FirstOrCreate(&genre, internal.Genre{Name: genreName})
		db.FirstOrCreate(
			&internal.GameGenre{GameID: game.ID, GenreID: genre.ID},
			internal.GameGenre{GameID: game.ID, GenreID: genre.ID},
		)
	}

	// Handle platforms and time to beat
	for platformName, times := range gameData.Stats.Platform {
		var platform internal.Platform
		db.FirstOrCreate(&platform, internal.Platform{Name: platformName})

		timeToBeat := parseTimeToBeat(times["Main"])
		db.FirstOrCreate(
			&internal.GamePlatform{GameID: game.ID, PlatformID: platform.ID, TimeToBeat: timeToBeat},
			internal.GamePlatform{GameID: game.ID, PlatformID: platform.ID},
		)
	}
}

func parseTimeToBeat(timeStr string) int {
	if timeStr == "--" {
		return 0
	}

	time := 0
	timeParts := strings.Split(timeStr, " ")

	// iterate over the time parts
	for _, part := range timeParts {
		// check if the part contains 'h'
		if strings.Contains(part, "h") {
			// remove the 'h' from the part
			part = strings.Replace(part, "h", "", -1)
			// convert the part to an integer
			hours, err := strconv.Atoi(part)
			if err != nil {
				log.Println("Error converting hours to integer:", err)
				log.Println("Time:", timeStr)
			}
			// add the hours to the total time
			time += hours * 60
		} else if strings.Contains(part, "m") {
			// remove the 'm' from the part
			part = strings.Replace(part, "m", "", -1)
			// convert the part to an integer
			minutes, err := strconv.Atoi(part)
			if err != nil {
				fmt.Println("Error converting minutes to integer:", err)
				fmt.Println("Time:", timeStr)
			}
			// add the minutes to the total time
			time += minutes
		}
	}

	return time
}

func main() {
	// Parse the file path
	filePath := parseFilePath()

	// Open the jsonlines file
	file := openFile(*filePath)

	// Initialize the database
	db := openDB()

	// Migrate the schema
	if err := internal.Migrate(db); err != nil {
		fmt.Println("Error migrating database:", err)
		return
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		processLine(scanner, db)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
}
