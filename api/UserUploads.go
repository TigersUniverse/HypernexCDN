package api

import (
	"HypernexCDN/tools"
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
)

type UserUploads struct {
	UserId  string       `json:"UserId" bson:"UserId"`
	Uploads []FileUpload `json:"Uploads" bson:"Uploads"`
}

func getSize(stream io.Reader) (bool, int) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(stream)
	if err != nil {
		return false, 0
	}
	return true, buf.Len()
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

func (u *UserUploads) CreateUpload(uploadType int, ext string, file io.Reader) (bool, *FileUpload) {
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
	hash := md5.New()
	_, err := io.Copy(hash, file)
	if err != nil {
		return false, nil
	}
	upload.Hash = fmt.Sprintf("%x", hash.Sum(nil))
	success, size := getSize(file)
	if !success {
		return false, nil
	}
	upload.Size = size
	u.Uploads = append(u.Uploads, upload)
	return true, &upload
}
