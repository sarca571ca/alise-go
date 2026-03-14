package config

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Channels struct {
	HNMTimes    string
	BotCommands string
	CampPings   string
	BotLogs     string
}

type Categories struct {
	HNMCategoryID        string
	AwaitingProcessingID string
	DKPReviewID          string
	AttendanceArchiveID  string
	VIPID                string
}
type Config struct {
	Token      string
	GuildID    string
	Channels   Channels
	Categories Categories
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
		HNMTimes:    os.Getenv("HNMTIMES"),
		BotCommands: os.Getenv("BOTCOMMANDS"),
		CampPings:   os.Getenv("CAMPPINGS"),
		BotLogs:     os.Getenv("BOTLOGS"),
	}
	categories := Categories{
		HNMCategoryID:        os.Getenv("HNMCATEGORYID"),
		AwaitingProcessingID: os.Getenv("AWAITINGPROCESSINGID"),
		DKPReviewID:          os.Getenv("DKPREVIEWID"),
		AttendanceArchiveID:  os.Getenv("ATTENDANCEARCHIVEID"),
		VIPID:                os.Getenv("VIPCATID"),
	}
	return Config{
			Token:      token,
			GuildID:    os.Getenv("GUILDID"),
			Channels:   channels,
			Categories: categories,
		},
		nil
}

// // .env imports
// // Required
// var Token = GetString("TOKEN")
// var GuildID = GetInt("GUILDID")
//
// // Channels
// var HNMTimes = GetInt("HNMTIMES")
// var BotCommands = GetInt("BOTCOMMANDS")
// var CampPings = GetInt("CAMPPINGS")
// var BotLogs = GetInt("BOTLOGS")
//
// // Categories
// var HNMCategoryID = GetInt("HNMCATEGORYID")
// var AwaitingProcessingID = GetInt("AWAITINGPROCESSINGID")
// var DKPReviewID = GetInt("DKPREVIEWID")
// var AttendanceArchiveID = GetInt("ATTENDANCEARCHIVEID")
// var VIPID = GetInt("VIPCATID")
//
// // hnm aliases
// var Fafnir = []string{"faf", "fafnir"}
// var Adamantoise = []string{"ad", "ada", "adam", "adamantoise"}
// var Behemoth = []string{"beh", "behe", "behemoth"}
// var KingArthro = []string{"ka", "kinga"}
// var Simurgh = []string{"sim"}
// var ShikigamiWeapon = []string{"shi", "shiki", "shikigami"}
// var KingVinegarroon = []string{"kv", "kingv", "kingvine"}
// var Bloodsucker = []string{"bs", "blo", "bloodsucker"}
// var Vrtra = []string{"vrt", "vrtr", "vrtra"}
// var Tiamat = []string{"tia", "tiam", "tiamat"}
// var Jormungand = []string{"jor", "jorm", "jormungand"}
//
// // Date Time formats
//
// var DateFormats = []string{
// 	"%Y-%m-%d %I%M%S %p",
// 	"%Y%m%d %I%M%S %p",
// 	"%y%m%d %I%M%S %p",
// 	"%m%d%Y %I%M%S %p",
// 	"%m%d%Y %I%M%S %p",
// 	"%Y-%m-%d %I:%M:%S %p",
// 	"%Y%m%d %I:%M:%S %p",
// 	"%y%m%d %I:%M:%S %p",
// 	"%m%d%Y %I:%M:%S %p",
// 	"%m%d%Y %I:%M:%S %p",
// 	"%Y-%m-%d %H%M%S",
// 	"%Y%m%d %H%M%S",
// 	"%y%m%d %H%M%S",
// 	"%m%d%Y %H%M%S",
// 	"%m%d%Y %H%M%S",
// 	"%Y-%m-%d %H:%M:%S",
// 	"%Y%m%d %H:%M:%S",
// 	"%y%m%d %H:%M:%S",
// 	"%m%d%Y %H:%M:%S",
// 	"%m%d%Y %H:%M:%S",
// 	"%Y-%m-%d %h%M%S",
// 	"%Y%m%d %h%M%S",
// 	"%y%m%d %h%M%S",
// 	"%m%d%Y %h%M%S",
// 	"%m%d%Y %h%M%S",
// 	"%Y-%m-%d %h:%M:%S",
// 	"%Y%m%d %h:%M:%S",
// 	"%y%m%d %h:%M:%S",
// 	"%m%d%Y %h:%M:%S",
// 	"%m%d%Y %h:%M:%S",
// }
//
// var TimeFormats = []string{"%I%M%S %p", "%I:%M:%S %p", "%H%M%S", "%H:%M:%S", "%h:%M:%S"}
//
// var TimeZone = "America/Los_Angeles"
//
// // Wait time vars
// var WindowLeadTimeMinutes = 20
// var ArchiveMoveTimeMinutes = 5
//
// // Window messages
// var GeneralWindowMessage = "```" + `
// Note:
//     - Channel will be open for 5-Minutes after pop/last window.
//     - Channel is moved to Awaiting Processing category.
//     - Late x-in's (within reason) or corrections to your camp status can be made after its moved.
//     - DO NOT X-IN before arriving to camp. This means in position and buffed.
// ` + "```"
// var KingVinegarroonWindowMessage = "```" + `
// Note:
//     - x         - used when you are at kv with the window open prior to pop
//     - x-pop     - used when you are present when KV pops and we do NOT claim
//     - x-claim   - used when you are present when KV pops and we DO claim
//     - x-kill    - used when you are present for the kill of KV
//
//     x-pop and x-claim are mutually exclusive
// ` + "```"
// var GrandWyrmWindowMessage = "```" + `
// Note:
//     - A valid hold party must be present for dkp.
//     - Conditions for valid hold party are: Tank (w/ Resist Set), BRD, WHM, 2 Sleeps
//     - The !pop command will work within this channel.
//     - Windows will be opened 5-Minutes prior to window and closed 1-Minute after window.
//     - Late x-in's won't be allowed due to the nature of this camp.
// ` + "```"
//
// // HNM Globals
// var GrandWyrms = []string{"Jormungand", "Tiamat", "Vrtra"}
// var CampsGreaterThanDay = []string{"Bloodsucker", "Jormungand", "Tiamat", "Vrtra"}
// var Kings = []string{"Fafnir", "Adamantoise", "Behemoth"}
// var HighQualityKings = map[string]string{"Fafnir": "Nidhogg", "Adamantoise": "Aspidochelone", "Behemoth": "King Behemoth"}
// var SpawnGroupOne = []string{"Fafnir", "Adamantoise", "Behemoth", "King Arthro", "Simurgh"}
// var SpawnGroupTwo = []string{"Shikigami Weapon", "King Vinegarroon"}
//
// var HNMInfo = map[string]string{
// 	"faf": "Can't miss him.",
// 	"ada": "https://media.discordapp.net/attachments/1175180691550523452/1261704293166092368/image.png?ex=66953eb0&is=6693ed30&hm=ba8558476bd46f7996b6a93233a2b8548ca8528eaa2ffe828ac02b997b3dfbdb&=&format=webp&quality=lossless",
// 	"beh": "https://cdn.discordapp.com/attachments/1175180586030202940/1261506701153275926/image.png?ex=6695d82b&is=669486ab&hm=20d64d9dae37e141db381b5fe9db3653474971e1d6dadecd49b56a4441bf4bb3&",
// 	"kv":  "---> [Weather Forecast - HorizonXI Wiki](https://horizonffxi.wiki/Special:WeatherForecast?weatherTypeDropDown=8&zoneNameDropDown=Western_Altepa_Desert) <--- Please check if there is upcoming weather.",
// }
//
// var HNMFormating = map[string]map[string]string{
// 	"n": {
// 		"Fafnir":           ":dragon_face: (****):",
// 		"Adamantoise":      ":turtle: (****):",
// 		"Behemoth":         ":zap: (****):",
// 		"King Arthro":      ":crab::",
// 		"King Vinegarroon": ":scorpion::",
// 		"Bloodsucker":      ":drop_of_blood::",
// 		"Shikigami Weapon": ":japanese_ogre::",
// 		"Simurgh":          ":bird::",
// 		"Jormungand":       ":ice_cube::chicken::ice_cube::",
// 		"Tiamat":           ":fire::chicken::fire::",
// 		"Vrtra":            ":skull::chicken::skull::",
// 	},
// 	"a": {
// 		"Fafnir":           ":grey_question::dragon_face::grey_question: (****):",
// 		"Adamantoise":      ":grey_question::turtle::grey_question: (****):",
// 		"Behemoth":         ":grey_question::zap::grey_question: (****):",
// 		"King Arthro":      ":grey_question::crab::grey_question::",
// 		"King Vinegarroon": ":grey_question::scorpion::grey_question::",
// 		"Bloodsucker":      ":grey_question::drop_of_blood::grey_question::",
// 		"Shikigami Weapon": ":grey_question::japanese_ogre::grey_question::",
// 		"Simurgh":          ":grey_question::bird::grey_question::",
// 		"Jormungand":       ":grey_question::ice_cube::chicken::ice_cube::grey_question::",
// 		"Tiamat":           ":grey_question::fire::chicken::fire::grey_question::",
// 		"Vrtra":            ":grey_question::skull::chicken::skull::grey_question::",
// 	},
// 	"d": {
// 		"Fafnir":      ":moneybag::dragon_face::moneybag: (****):",
// 		"Adamantoise": ":moneybag::turtle::moneybag: (****):",
// 		"Behemoth":    ":moneybag::zap::moneybag: (****):",
// 	},
// 	"t": {
// 		"Fafnir":      ":gem::dragon_face::gem: (****):",
// 		"Adamantoise": ":gem::turtle::gem: (****):",
// 		"Behemoth":    ":gem::zap::gem: (****):",
// 	},
// }
