package config

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Channels struct {
	HNMTimes          string
	ClaimsLeaderBoard string
	BotCommands       string
	CampPings         string
	BotLogs           string
}

type Categories struct {
	HNMCategoryID        string
	AwaitingProcessingID string
	DKPReviewID          string
	AttendanceArchiveID  string
	VIPID                string
}
type Config struct {
	Token         string
	GuildID       string
	Channels      Channels
	Categories    Categories
	KVWeatherBlob string
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("Failed to load .env")
	}
	token := os.Getenv("TOKEN")
	if token == "" {
		return Config{}, errors.New("TOKEN is required.")
	}

	channels := Channels{
		HNMTimes:          os.Getenv("HNMTIMES"),
		ClaimsLeaderBoard: os.Getenv("CLAIMSLEADERBOARD"),
		BotCommands:       os.Getenv("BOTCOMMANDS"),
		CampPings:         os.Getenv("CAMPPINGS"),
		BotLogs:           os.Getenv("BOTLOGS"),
	}
	categories := Categories{
		HNMCategoryID:        os.Getenv("HNMCATEGORYID"),
		AwaitingProcessingID: os.Getenv("AWAITINGPROCESSINGID"),
		DKPReviewID:          os.Getenv("DKPREVIEWID"),
		AttendanceArchiveID:  os.Getenv("ATTENDANCEARCHIVEID"),
		VIPID:                os.Getenv("VIPCATID"),
	}
	kVWeatherBlob := os.Getenv("BLOB")
	return Config{
			Token:         token,
			GuildID:       os.Getenv("GUILDID"),
			Channels:      channels,
			Categories:    categories,
			KVWeatherBlob: kVWeatherBlob,
		},
		nil
}
