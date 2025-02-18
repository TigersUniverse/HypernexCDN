package tools

import "github.com/google/uuid"

func NewId(idType int) string {
	idprefix := "_"
	switch idType {
	case 0:
		idprefix = "user_"
		break
	case 1:
		idprefix = "avatar_"
		break
	case 2:
		idprefix = "world_"
		break
	case 3:
		idprefix = "post_"
		break
	case 4:
		idprefix = "file_"
		break
	case 5:
		idprefix = "gameserver_"
		break
	case 6:
		idprefix = "instance_"
		break
	}
	return idprefix + uuid.New().String()
}
