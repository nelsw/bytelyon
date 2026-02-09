package model

type Search struct {
	Model
	Bot   *Bot `json:",omitempty"`
	BotID uint
	Pages []*SearchPage
}
