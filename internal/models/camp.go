package models

import (
	"alise-go/embedded"
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func GetCampInfo(hnm HNM, timer HNMTimer, firstWindow time.Time) *discordgo.MessageSend {
	var fields []*discordgo.MessageEmbedField
	hnmInfo := fmt.Sprintf(
		// TODO: need to display any double or triple dkp stuff here
		"First Window: <t:%d:T> <t:%d:R>",
		firstWindow.Unix(),
		firstWindow.Unix(),
	)
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "Info",
		Value:  hnmInfo,
		Inline: false,
	})
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "Note",
		Value:  hnm.Note,
		Inline: false,
	})
	if hnm.ID == "kv" {
		weatherBlob := os.Getenv("BLOB")

		wb, err := NewWeatherBlob(weatherBlob)
		if err != nil {
			log.Fatal(err)
		}

		vt, err := FetchCurrentVanaTime()
		var day DayWeather

		if err == nil {
			day, err = wb.WeatherForVanaDay(AbsoluteVanaDay(vt.Year, vt.Month, vt.Day))
		} else {
			day, err = wb.WeatherForNowFallback(time.Now())
		}

		if day == (DayWeather{}) {
			log.Println("Failed to load VanaDay")
		}
		if err != nil {
			log.Println(err)
		}
		weatherForecast, err := wb.BuildWeatherForecast(10, time.Now())
		if err != nil {
			log.Fatal(err)
		}

		formatedWeatherForecast := FormatWeatherForcastPlain(weatherForecast)
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Weather",
			Value:  formatedWeatherForecast,
			Inline: false,
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:  hnm.Name,
		Fields: fields,
	}

	msg := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{embed},
	}

	bMap, err := embedded.GetHNMMapBytes(hnm.ID)
	if err == nil && len(bMap) > 0 {
		embed.Image = &discordgo.MessageEmbedImage{
			URL: "attachment://map.png",
		}
		msg.Files = append(msg.Files, &discordgo.File{
			Name:   "map.png",
			Reader: bytes.NewReader(bMap),
		})
	}

	bThumbnail, err := embedded.GetHNMThumbnailBytes(hnm.ID)
	if err == nil && len(bThumbnail) > 0 {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: "attachment://thumbnail.png",
		}
		msg.Files = append(msg.Files, &discordgo.File{
			Name:   "thumbnail.png",
			Reader: bytes.NewReader(bThumbnail),
		})
	}
	return msg
}

func GetRSVPInput(hnm HNM, firstWindow time.Time) string {
	var sb strings.Builder
	for w := range hnm.WindowCount {
		sb.WriteString(
			fmt.Sprintf(
				"%s (%d/%d)|||%s|||%s|||%s",
				hnm.Name,
				w,
				hnm.WindowCount,
				firstWindow.Unix(),
			),
		)
	}
	return sb.String()
}
