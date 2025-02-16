package search

import "HypernexCDN/api"

type AvatarMeta struct {
	Id          string      `json:"Id"`
	OwnerId     string      `json:"OwnerId"`
	Publicity   int         `json:"Publicity"`
	Name        string      `json:"Name"`
	Description string      `json:"Description"`
	Tags        []string    `json:"Tags"`
	ImageURL    string      `json:"ImageURL"`
	Builds      []Build     `json:"Builds"`
	Tokens      []api.Token `json:"Tokens"`
}

func (m AvatarMeta) ValidateToken(content string) bool {
	for _, token := range m.Tokens {
		if token.Content == content {
			return token.IsValid()
		}
	}
	return false
}
