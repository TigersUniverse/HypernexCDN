package api

import "strconv"

type FileUpload struct {
	UserID     string `json:"UserID" bson:"UserId"`
	FileId     string `json:"FileId" bson:"FileId"`
	FileName   string `json:"FileName" bson:"FileName"`
	UploadType int    `json:"UploadType" bson:"UploadType"`
	Key        string `json:"Key" bson:"Key"`
}

func (upload FileUpload) ToJSON() string {
	return "{\"UserId\": \"" + upload.UserID + "\", \"FileId\": \"" + upload.FileId + "\", \"FileName\": \"" + upload.FileName + "\", \"UploadType\": " + strconv.Itoa(upload.UploadType) + ", \"Key\": \"" + upload.Key + "\"}"
}
