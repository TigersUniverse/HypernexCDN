package api

import "time"

type Token struct {
	Content     string `json:"content"`
	DateCreated int64  `json:"dateCreated"`
	DateExpire  int64  `json:"dateExpire"`
	App         string `json:"app"`
}

func (t Token) IsValid() bool {
	if t.DateExpire == 0 {
		return true
	}
	timeNow := time.Now().Unix()
	return t.DateExpire > timeNow
}
