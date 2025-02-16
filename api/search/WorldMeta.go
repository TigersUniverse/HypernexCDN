package search

import "HypernexCDN/api"

type WorldMeta struct {
	Id           string      `json:"Id"`
	OwnerId      string      `json:"OwnerId"`
	Publicity    int         `json:"Publicity"`
	Name         string      `json:"Name"`
	Description  string      `json:"Description"`
	Tags         []string    `json:"Tags"`
	ThumbnailURL string      `json:"ThumbnailURL"`
	IconURLs     []string    `json:"IconURLs"`
	Builds       []Build     `json:"Builds"`
	Tokens       []api.Token `json:"Tokens"`
}

func (m WorldMeta) ValidateToken(content string) bool {
	for _, token := range m.Tokens {
		if token.Content == content {
			return token.IsValid()
		}
	}
	return false
}
