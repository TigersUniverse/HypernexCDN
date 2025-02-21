package api

import (
	"HypernexCDN/tools"
)

type UserUploads struct {
	UserId  string       `json:"UserId" bson:"UserId"`
	Uploads []FileUpload `json:"Uploads" bson:"Uploads"`
}

func (u *UserUploads) Contains(fileid string) bool {
	contains := false
	for _, upload := range u.Uploads {
		if upload.FileId == fileid {
			contains = true
			break
		}
	}
	return contains
}

func (u *UserUploads) CreateUpload(uploadType int, ext string) FileUpload {
	upload := FileUpload{
		UserID:     u.UserId,
		UploadType: uploadType,
	}
	upload.FileId = tools.NewId(4)
	for {
		if !u.Contains(upload.FileId) {
			break
		}
		upload.FileId = tools.NewId(4)
	}
	upload.FileName = upload.FileId + ext
	upload.Key = u.UserId + "/" + upload.FileId + ext
	u.Uploads = append(u.Uploads, upload)
	return upload
}
