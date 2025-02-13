package cdn

import (
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"strconv"
)

func CreateRoutes(r *mux.Router) {
	r.HandleFunc("/file/{userid}/{fileid}", getFile).Methods("GET")
	// TODO: Add other file endpoints
	r.HandleFunc("/upload", uploadHandler).Methods("POST")
}

// TODO: Actually use this
func msg(success bool, msg string) string {
	return "{\"success\": " + strconv.FormatBool(success) + ", \"" + msg + "\"}"
}

func getFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userid := vars["userid"]
	fileid := vars["fileid"]
	filePath := bucket + "/" + userid + "/" + fileid + ".png"
	// TODO: Pull filedata from Mongo and verify publicity
	obj, err := GetObject(filePath)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to get file", http.StatusInternalServerError)
		return
	}
	defer obj.Body.Close()
	w.Header().Set("Content-Disposition", "attachment; filename="+fileid)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", *obj.ContentLength))
	_, err = io.Copy(w, obj.Body)
	if err != nil {
		http.Error(w, "Error sending file", http.StatusInternalServerError)
		return
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	// TODO: Set memory max correctly
	err := r.ParseMultipartForm(10000)
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}
	userid := r.FormValue("userid")
	tokenContent := r.FormValue("tokenContent")
	if userid == "" || tokenContent == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	err = file.Close()
	if err != nil {
		http.Error(w, "Error closing file", http.StatusBadRequest)
		return
	}
	// TODO: Sanitize file name and verify file type
	filePath := bucket + "/" + userid + "/" + fileHeader.Filename
	err = UploadToS3(file, filePath)
	if err != nil {
		http.Error(w, "Failed to upload file", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
