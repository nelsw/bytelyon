package model

type SearchBot struct {
	Bot
	Headless bool `json:"headless" dynamodbav:"Headless,boolean"`
}
