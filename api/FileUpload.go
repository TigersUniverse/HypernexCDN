package api

type FileUpload struct {
	UserID     string `bson:"UserId"`
	FileId     string `bson:"FileId"`
	FileName   string `bson:"FileName"`
	UploadType int    `bson:"UploadType"`
	Key        string `bson:"Key"`
}
