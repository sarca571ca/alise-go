package bot

import (
	"alise-go/internal/commands"
	"alise-go/internal/config"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (b *Bot) buildLinkshellCommand(cfg config.Config) commands.Command {
	linkshellCmd := commands.LinkshellCommand{
		List: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			guildID := i.GuildID
			recs, err := b.store.ListLinkshellRecords(guildID)
			if err != nil {
				respondEphemeral(s, i, "Failed to list linkshell records.")
			}

			msg := formatLinkshellListPlain(b.store.NewLinkshellListFromRecords(recs))
			respondEphemeral(s, i, msg)
		},
		Add: func(s *discordgo.Session, i *discordgo.InteractionCreate, ls string) {
			guildID := i.GuildID
			_, ok, err := b.store.GetLinkshellRecord(guildID, ls)
			if err != nil {
				respondEphemeral(s, i, err.Error())
				return
			}
			if ok {
				respondEphemeral(s, i, "The linkshell already exists.")
				return
			}
			rec := b.store.NewRecordFromLinkshellEntry(guildID, ls)
			if _, err := b.store.UpsertLinkshellRecord(rec); err != nil {
				respondEphemeral(s, i, "Failed to save Linkshell")
				return
			}

			var sb strings.Builder
			fmt.Fprintf(&sb, "Successfully added %s to the linkshells list.", ls)
			respondEphemeral(s, i, sb.String())

		},
		Remove: func(s *discordgo.Session, i *discordgo.InteractionCreate, ls string) {
			guildID := i.GuildID
			_, ok, err := b.store.GetLinkshellRecord(guildID, ls)
			if err != nil {
				respondEphemeral(s, i, err.Error())
				return
			}
			if !ok {
				respondEphemeral(s, i, "The linkshell doesn't exist.")
				return
			}
			if err := b.store.DeleteLinkshellRecord(guildID, ls); err != nil {
				respondEphemeral(s, i, "Failed to remove Linkshell")
				return
			}

			var sb strings.Builder
			fmt.Fprintf(&sb, "Successfully removed %s to the linkshells list.", ls)
			respondEphemeral(s, i, sb.String())
		},
		Restore: func(s *discordgo.Session, i *discordgo.InteractionCreate, ls string) {
			guildID := i.GuildID
			_, ok, err := b.store.GetLinkshellRecord(guildID, ls)
			if err != nil {
				respondEphemeral(s, i, err.Error())
				return
			}
			if ok {
				respondEphemeral(s, i, "The linkshell already exist.")
				return
			}
			if _, err := b.store.RestoreLinkshellRecord(guildID, ls); err != nil {
				respondEphemeral(s, i, "Failed to restore Linkshell")
				return
			}

			var sb strings.Builder
			fmt.Fprintf(&sb, "Successfully restored %s to the linkshells list.", ls)
			respondEphemeral(s, i, sb.String())
		},
		// NOTE: I'm not going to implement this yet there is to much of a concern of manipulating data.
		// Update modal that self populates the the fields with current values
		Update: func(s *discordgo.Session, i *discordgo.InteractionCreate, ls string) {},
	}
	return linkshellCmd
}
