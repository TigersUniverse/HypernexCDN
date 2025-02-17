package data

import "HypernexCDN/api"

type UserData struct {
	Id            string      `json:"Id"`
	AccountTokens []api.Token `json:"AccountTokens"`
}
