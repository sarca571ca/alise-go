package bot

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type LogEntry struct {
	Name    string
	MsgType string
	Msg     string
}

func (b *Bot) NewLogEntry(logEntry LogEntry) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "[%-8s] %s: %s", logEntry.Name, logEntry.MsgType, logEntry.Msg)
	return sb.String()
}

func (b *Bot) logCommandUsage(s *discordgo.Session, i *discordgo.InteractionCreate, cmdName, subName string) {
	var sb strings.Builder

	ch, err := s.Channel(i.ChannelID)
	if err != nil {
		return
	}

	if subName == "" {
		fmt.Fprintf(
			&sb,
			"{%s} User: %s(%s) Channel: %s",
			cmdName,
			i.Member.Nick,
			i.Member.User.Username,
			ch.Name,
		)
	} else {
		fmt.Fprintf(
			&sb,
			"{%s->%s} User: %s(%s) Channel: %s",
			cmdName,
			subName,
			i.Member.Nick,
			i.Member.User.Username,
			ch.Name,
		)
	}

	logEntry := b.NewLogEntry(LogEntry{
		Name:    "Alise",
		MsgType: "Command",
		Msg:     sb.String(),
	})

	log.Printf(logEntry)
}

func (b *Bot) logErrorMessage(group string, err error) {
	var sb strings.Builder

	fmt.Fprintf(&sb, "{%s} -> %s", group, err.Error())

	logEntry := b.NewLogEntry(LogEntry{
		Name:    "Alise",
		MsgType: "Error",
		Msg:     sb.String(),
	})

	log.Printf(logEntry)
}

func (b *Bot) logBasicMessage(group, msg string) {
	var sb strings.Builder

	fmt.Fprintf(&sb, "{%s} -> %s", group, msg)

	logEntry := b.NewLogEntry(LogEntry{
		Name:    "Alise",
		MsgType: "Basic",
		Msg:     sb.String(),
	})

	log.Printf(logEntry)
}
