package cdn

import (
	"HypernexCDN/api"
	"HypernexCDN/api/api_responses"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
)

var allowAnyGameServer = false
var validExtensions = [][]string{
	{".jpg", ".jpeg", ".gif", ".png", ".mp4"},
	{".hna"},
	{".hnw"},
	{".js", ".lua"},
}

func CreateRoutes(r *mux.Router) {
	a, err := api.GetApiResponse[api_responses.AllowAnyGameServer]("allowAnyGameServer")
	if err != nil {
		panic(err)
	}
	allowAnyGameServer = a.AllowAnyGameServer
	r.HandleFunc("/file/{userid}/{fileid}", getFile).Methods("GET")
	r.HandleFunc("/file/{userid}/{fileid}/{filetoken}", getFileToken).Methods("GET")
	r.HandleFunc("/file/{userid}/{fileid}/{gameServerId}/{gameServerToken}", getServerScript).Methods("GET")
	r.HandleFunc("/upload", uploadHandler).Methods("POST")
}

func msg(success bool, msg string) string {
	return "{\"success\": " + strconv.FormatBool(success) + ", \"" + msg + "\"}"
}

func returnFile(w http.ResponseWriter, fileMeta api.FileUpload) {
	obj, err := GetObject(bucket + "/" + fileMeta.Key)
	if err != nil {
		http.Error(w, msg(false, "Failed to get file"), http.StatusInternalServerError)
		return
	}
	noError := true
	defer func(Body io.ReadCloser) {
		err2 := Body.Close()
		if err2 != nil {
			http.Error(w, msg(false, "Failed to get file"), http.StatusInternalServerError)
			noError = false
			return
		}
	}(obj.Body)
	if !noError {
		return
	}
	if obj.ContentLength != nil {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", *obj.ContentLength))
	}
	w.Header().Set("Content-Disposition", "attachment; filename="+fileMeta.FileName)
	w.Header().Set("Content-Type", "application/octet-stream")
	_, err = io.Copy(w, obj.Body)
	if err != nil {
		http.Error(w, msg(false, "Error sending file"), http.StatusInternalServerError)
		return
	}
}

func getFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userid := vars["userid"]
	fileid := vars["fileid"]
	fileMeta := GetFileMetaById(userid, fileid)
	if fileMeta == nil {
		http.Error(w, msg(false, "Failed to find file"), http.StatusNotFound)
		return
	}
	switch fileMeta.UploadType {
	// media
	case 0:
		returnFile(w, *fileMeta)
		return
	// avatar
	case 1:
		avatarMeta := GetAvatarMetaFromFileId(userid, fileid)
		if avatarMeta == nil {
			http.Error(w, msg(false, "Failed to find avatar"), http.StatusNotFound)
			return
		}
		if avatarMeta.Publicity > 0 {
			break
		}
		popularity := GetOrCreatePopularity(avatarMeta.Id)
		if popularity != nil {
			popularity.AddPopularityUsage()
			UpdatePopularity(popularity)
		}
		returnFile(w, *fileMeta)
		return
	// world
	case 2:
		worldMeta := GetWorldMetaFromFileId(userid, fileid)
		if worldMeta == nil {
			http.Error(w, msg(false, "Failed to find world"), http.StatusNotFound)
			return
		}
		if worldMeta.Publicity > 0 {
			break
		}
		popularity := GetOrCreatePopularity(worldMeta.Id)
		if popularity != nil {
			popularity.AddPopularityUsage()
			UpdatePopularity(popularity)
		}
		returnFile(w, *fileMeta)
		return
	// game server script
	case 3:
		// only allow if any game server is allowed
		if allowAnyGameServer {
			returnFile(w, *fileMeta)
			return
		}
		break
	}
	http.Error(w, msg(false, "No permissions!"), http.StatusForbidden)
}

func getFileToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userid := vars["userid"]
	fileid := vars["fileid"]
	filetoken := vars["filetoken"]
	fileMeta := GetFileMetaById(userid, fileid)
	if fileMeta == nil {
		http.Error(w, msg(false, "Failed to find file"), http.StatusNotFound)
		return
	}
	switch fileMeta.UploadType {
	// avatar
	case 1:
		avatarMeta := GetAvatarMetaFromFileId(userid, fileid)
		if avatarMeta == nil {
			http.Error(w, msg(false, "Failed to find avatar"), http.StatusNotFound)
			return
		}
		if !avatarMeta.ValidateToken(filetoken) {
			break
		}
		returnFile(w, *fileMeta)
		return
	// world
	case 2:
		worldMeta := GetWorldMetaFromFileId(userid, fileid)
		if worldMeta == nil {
			http.Error(w, msg(false, "Failed to find world"), http.StatusNotFound)
			return
		}
		if !worldMeta.ValidateToken(filetoken) {
			break
		}
		returnFile(w, *fileMeta)
		return
	// game server script
	case 3:
		if allowAnyGameServer {
			returnFile(w, *fileMeta)
			return
		}
		// TODO: Upstream API Implementation
		response, err := api.GetApiResponse[api_responses.ValidGameServer]("checkGameServer/" + filetoken)
		if err != nil {
			http.Error(w, msg(false, "Failed to verify GameServer"), http.StatusInternalServerError)
			return
		}
		if !response.Valid {
			break
		}
		returnFile(w, *fileMeta)
		return
	}
	http.Error(w, msg(false, "No permissions!"), http.StatusForbidden)
}

func getServerScript(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userid := vars["userid"]
	fileid := vars["fileid"]
	gameServerId := vars["gameServerId"]
	gameServerToken := vars["gameServerToken"]
	fileMeta := GetFileMetaById(userid, fileid)
	if fileMeta == nil {
		http.Error(w, msg(false, "Failed to find file"), http.StatusNotFound)
		return
	}
	switch fileMeta.UploadType {
	// game server script
	case 3:
		if allowAnyGameServer {
			returnFile(w, *fileMeta)
			return
		}
		// TODO: Upstream API Implementation
		response, err := api.GetApiResponse[api_responses.ValidGameServer]("checkGameServer/" + gameServerId + "/" + gameServerToken)
		if err != nil {
			http.Error(w, msg(false, "Failed to verify GameServer"), http.StatusInternalServerError)
			return
		}
		if !response.Valid {
			break
		}
		returnFile(w, *fileMeta)
		return
	}
	http.Error(w, msg(false, "No permissions!"), http.StatusForbidden)
}

func validExtension(extension string) bool {
	for _, exts := range validExtensions {
		for _, ext := range exts {
			if ext == extension {
				return true
			}
		}
	}
	return false
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, msg(false, "Invalid request method"), http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseMultipartForm(1 << 30)
	if err != nil {
		http.Error(w, msg(false, "Request too large!"), http.StatusBadRequest)
		return
	}
	userid := r.FormValue("userid")
	tokenContent := r.FormValue("tokenContent")
	if userid == "" || tokenContent == "" {
		http.Error(w, msg(false, "Missing required parameters"), http.StatusBadRequest)
		return
	}
	userdata := GetUserData(userid)
	if userdata == nil {
		http.Error(w, msg(false, "Invalid user Id"), http.StatusBadRequest)
		return
	}
	valid := false
	for i := 0; i < len(userdata.AccountTokens); i++ {
		proposedToken := userdata.AccountTokens[i]
		validToken := tokenContent == proposedToken.Content && proposedToken.IsValid()
		if validToken {
			valid = true
			break
		}
	}
	if !valid {
		http.Error(w, msg(false, "Invalid user token"), http.StatusBadRequest)
		return
	}
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, msg(false, "Error retrieving file"), http.StatusBadRequest)
		return
	}
	err = file.Close()
	if err != nil {
		http.Error(w, msg(false, "Error closing file"), http.StatusInternalServerError)
		return
	}
	extension := filepath.Ext(fileHeader.Filename)
	if !validExtension(extension) {
		http.Error(w, msg(false, "Invalid extension"), http.StatusBadRequest)
		return
	}
	filePath := bucket + "/" + userdata.Id + "/" + fileHeader.Filename
	err = UploadToS3(file, filePath)
	if err != nil {
		http.Error(w, "Failed to upload file", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, errf := fmt.Fprintf(w, msg(true, "File uploaded!"))
	if errf != nil {
		fmt.Println(errf)
	}
}
