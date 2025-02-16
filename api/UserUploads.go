package api

type UserUploads struct {
	UserId  string       `json:"UserId"`
	Uploads []FileUpload `json:"Uploads"`
}
