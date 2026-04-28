package commands

import "github.com/bwmarrin/discordgo"

type CommandFunc struct {
	NameFn        func() string
	DescriptionFn func() string
	SlashDefFn    func() *discordgo.ApplicationCommand
	HandleFn      func(s *discordgo.Session, i *discordgo.InteractionCreate)
	AutoFn        func(
		s *discordgo.Session,
		i *discordgo.InteractionCreate,
		focused *discordgo.ApplicationCommandInteractionDataOption,
	)
}

func (c CommandFunc) Name() string {
	return c.NameFn()
}

func (c CommandFunc) Description() string {
	return c.DescriptionFn()
}

func (c CommandFunc) SlashDef() *discordgo.ApplicationCommand {
	return c.SlashDefFn()
}

func (c CommandFunc) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	c.HandleFn(s, i)
}

func (c CommandFunc) HandleAutoCompleteCommand(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	focused *discordgo.ApplicationCommandInteractionDataOption,
) {
	if c.AutoFn != nil {
		c.AutoFn(s, i, focused)
	}
}
