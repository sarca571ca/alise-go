package models

import (
	"alise-go/embedded"
	"bytes"
	"fmt"
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

// name (currentwindow/windows)|||start|||incrementfromstart|||name
// Adamantoise (1/7)|||1777772698|||1777701600|||Adamantoise
