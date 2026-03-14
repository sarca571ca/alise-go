package commands

import (
	"github.com/bwmarrin/discordgo"
)

type Command interface {
	Name() string
	Description() string
	SlashDef() *discordgo.ApplicationCommand
	HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate)
}

type AutoCompleteCommand interface {
	HandleAutoCompleteCommand(
		s *discordgo.Session,
		i *discordgo.InteractionCreate,
		focused *discordgo.ApplicationCommandInteractionDataOption,
	)
}

type Registery struct {
	cmds map[string]Command
}

func NewRegistry() *Registery {
	return &Registery{cmds: make(map[string]Command)}
}

func (r *Registery) Register(c Command) {
	r.cmds[c.Name()] = c
}

func (r *Registery) Get(name string) (Command, bool) {
	c, ok := r.cmds[name]
	return c, ok
}

func (r *Registery) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		data := i.ApplicationCommandData()
		cmd, ok := r.Get(data.Name)
		if !ok {
			return
		}
		cmd.HandleInteraction(s, i)

	case discordgo.InteractionApplicationCommandAutocomplete:
		r.handleCommandAutcomplete(s, i)

	case discordgo.InteractionMessageComponent:
		r.handleComponentInteractions(s, i)

	case discordgo.InteractionModalSubmit:
		r.handleModalSubmit(s, i)
	}
}

func (r *Registery) SlashDefs() []*discordgo.ApplicationCommand {
	out := make([]*discordgo.ApplicationCommand, 0, len(r.cmds))
	for _, c := range r.cmds {
		out = append(out, c.SlashDef())
	}
	return out
}

func (r *Registery) handleCommandRun(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	if cmd, ok := r.Get(data.Name); ok {
		cmd.HandleInteraction(s, i)
	}
}

func (r *Registery) handleCommandAutcomplete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

	cmd, ok := r.Get(data.Name)
	if !ok {
		return
	}

	ac, ok := cmd.(AutoCompleteCommand)
	if !ok {
		return
	}

	var focused *discordgo.ApplicationCommandInteractionDataOption

	for _, opt := range data.Options {
		if opt.Focused {
			focused = opt
			break
		}
	}
	if focused == nil {
		return
	}

	ac.HandleAutoCompleteCommand(s, i, focused)
}

func (r *Registery) handleComponentInteractions(s *discordgo.Session, i *discordgo.InteractionCreate) {
}

func (r *Registery) handleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
}
