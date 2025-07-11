package cdn

import (
	"HypernexCDN/api"
	"HypernexCDN/api/api_responses"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"math/rand"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	MAX_UPLOAD_SIZE = 1 << 30
	MAX_MEM         = 32 << 20
)

var allowAnyGameServer = false
var validExtensions = [][]string{
	{".jpg", ".jpeg", ".gif", ".png", ".mp4"},
	{".hna"},
	{".hnw"},
	{".js", ".lua"},
}
var pics_bucket string
var public_pics string
var reg *regexp.Regexp

func CreateRoutes(r *mux.Router, b string, p string) {
	a, err := api.GetApiResponse[api_responses.AllowAnyGameServer]("allowAnyGameServer")
	if err != nil {
		panic(err)
	}
	allowAnyGameServer = a.AllowAnyGameServer
	r.Use(enableCORS)
	r.HandleFunc("/", root).Methods("GET")
	r.HandleFunc("/file/{userid}/{fileid}", getFile).Methods("GET")
	r.HandleFunc("/file/{userid}/{fileid}/{filetoken}", getFileToken).Methods("GET")
	r.HandleFunc("/file/{userid}/{fileid}/{gameServerId}/{gameServerToken}", getServerScript).Methods("GET")
	r.HandleFunc("/upload", uploadHandler).Methods("POST")
	pics_bucket = b
	public_pics = p
	if pics_bucket != "" {
		r.HandleFunc("/randomImage", randomImageHandler).Methods("GET")
		r.HandleFunc("/picture/{picture}", pictureHandler).Methods("GET")
	}
	reg = regexp.MustCompile(`^[\w\s-]+(\.[A-Za-z0-9]+)+$`)
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func msg(success bool, msg string) string {
	return "{\"success\": " + strconv.FormatBool(success) + ", \"message\": \"" + msg + "\"}"
}

func msg_result(success bool, msg string, result string) string {
	return "{\"success\": " + strconv.FormatBool(success) + ", \"message\": \"" + msg + "\", \"result\": " + result + "}"
}

func root(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, errf := fmt.Fprintf(w, msg(true, "Server running!"))
	if errf != nil {
		fmt.Println(errf)
	}
}

func returnFile(w http.ResponseWriter, fileMeta api.FileUpload) {
	obj, err := GetObject(fileMeta.Key)
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

func validExtension(extension string) int {
	for i, exts := range validExtensions {
		for _, ext := range exts {
			if ext == extension {
				return i
			}
		}
	}
	return -1
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)
	if r.Method != http.MethodPost {
		http.Error(w, msg(false, "Invalid request method"), http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseMultipartForm(MAX_MEM)
	if err != nil {
		http.Error(w, msg(false, "Request too large!"), http.StatusRequestEntityTooLarge)
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
	userUploads := GetUploadData(userdata.Id)
	if userUploads == nil {
		http.Error(w, msg(false, "No user uploads!"), http.StatusInternalServerError)
		return
	}
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, msg(false, "Error retrieving file"), http.StatusBadRequest)
		return
	}
	if fileHeader.Size > MAX_UPLOAD_SIZE {
		http.Error(w, msg(false, "File too large"), http.StatusRequestEntityTooLarge)
		return
	}
	extension := strings.ToLower(filepath.Ext(fileHeader.Filename))
	uploadType := validExtension(extension)
	if uploadType < 0 {
		http.Error(w, msg(false, "Invalid extension"), http.StatusBadRequest)
		return
	}
	success, fileUpload := userUploads.CreateUpload(uploadType, extension, file)
	if !success {
		http.Error(w, msg(false, "Failed to compute file"), http.StatusInternalServerError)
		return
	}
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		http.Error(w, msg(false, "Failed to seek file"), http.StatusInternalServerError)
		return
	}
	fileName := fileUpload.FileName
	filePath := userdata.Id + "/" + fileName
	err = UploadToS3(file, filePath)
	if err != nil {
		http.Error(w, msg(false, "Failed to upload file"), http.StatusInternalServerError)
		return
	}
	UpdateUploadData(userUploads)
	err = file.Close()
	if err != nil {
		http.Error(w, msg(false, "Error closing file"), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, errf := fmt.Fprintf(w, msg_result(true, "File uploaded!", "{\"UploadData\": "+fileUpload.ToJSON()+"}"))
	if errf != nil {
		fmt.Println(errf)
	}
}

func randomImageHandler(w http.ResponseWriter, r *http.Request) {
	pics, err := GetAllObjects(pics_bucket, public_pics)
	if err != nil {
		http.Error(w, msg(false, "Failed to get pictures"), http.StatusInternalServerError)
		return
	}
	keys := len(pics.Contents)
	i := rand.Intn(keys)
	picInfo := pics.Contents[i]
	fileKeySplit := strings.Split(*picInfo.Key, "/")
	fileName := fileKeySplit[len(fileKeySplit)-1]
	http.Redirect(w, r, "picture/"+fileName, http.StatusFound)
}

func pictureHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	picName := vars["picture"]
	if strings.Contains(picName, "?") {
		http.Error(w, msg(false, "Invalid path"), http.StatusBadRequest)
		return
	}
	if strings.Contains(picName, "$") {
		http.Error(w, msg(false, "Invalid path"), http.StatusBadRequest)
		return
	}
	if strings.Contains(picName, "%") {
		http.Error(w, msg(false, "Invalid path"), http.StatusBadRequest)
		return
	}
	if strings.Contains(picName, "..") {
		http.Error(w, msg(false, "Invalid path"), http.StatusBadRequest)
		return
	}
	if strings.Contains(picName, "/") {
		http.Error(w, msg(false, "Invalid path"), http.StatusBadRequest)
		return
	}
	valid := reg.MatchString(picName)
	if !valid {
		http.Error(w, msg(false, "Invalid pic name"), http.StatusBadRequest)
		return
	}
	obj, err := GetExactObject(pics_bucket, public_pics+picName)
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
	w.Header().Set("Content-Disposition", "attachment; filename="+picName)
	w.Header().Set("Content-Type", "application/octet-stream")
	_, err = io.Copy(w, obj.Body)
	if err != nil {
		http.Error(w, msg(false, "Error sending file"), http.StatusInternalServerError)
		return
	}
}
