package api

import (
	"HypernexCDN/tools"
)

type UserUploads struct {
	UserId  string       `json:"UserId"`
	Uploads []FileUpload `json:"Uploads"`
}

func (u UserUploads) Contains(fileid string) bool {
	contains := false
	for _, upload := range u.Uploads {
		if upload.FileId == fileid {
			contains = true
			break
		}
	}
	return contains
}

func (u UserUploads) CreateUpload(fileName string, uploadType int, ext string) {
	upload := FileUpload{
		UserID:     u.UserId,
		FileName:   fileName,
		UploadType: uploadType,
	}
	upload.FileId = tools.NewId(4)
	for {
		if !u.Contains(upload.FileId) {
			break
		}
		upload.FileId = tools.NewId(4)
	}
	upload.Key = u.UserId + "/" + upload.FileId + ext
}
