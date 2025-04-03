package handlers

import (
	"encoding/json"
	"expo-open-ota/internal/branch"
	"expo-open-ota/internal/bucket"
	"expo-open-ota/internal/helpers"
	"expo-open-ota/internal/services"
	"expo-open-ota/internal/types"
	"expo-open-ota/internal/update"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"
)

// UpdateUploadRequest represents the metadata about an update to be uploaded
type UpdateUploadRequest struct {
	Branch         string   `json:"branch"`
	Channel        string   `json:"channel"`
	Platform       string   `json:"platform"`
	RuntimeVersion string   `json:"runtimeVersion"`
	FileNames      []string `json:"fileNames"`
	CommitHash     string   `json:"commitHash"`
}

// UpdateUploadResponse is returned with information about the created update
type UpdateUploadResponse struct {
	UpdateId      string                     `json:"updateId"`
	Branch        string                     `json:"branch"`
	Platform      string                     `json:"platform"`
	UploadUrls    []bucket.FileUploadRequest `json:"uploadUrls"`
	RuntimeVersion string                    `json:"runtimeVersion"`
}

// UploadFileRequest handles the upload of a file for an update
type UploadFileRequest struct {
	Branch         string `json:"branch"`
	UpdateId       string `json:"updateId"`
	RuntimeVersion string `json:"runtimeVersion"`
	FileName       string `json:"fileName"`
}

// CompleteUpdateRequest is used to mark an update as completed
type CompleteUpdateRequest struct {
	Branch         string `json:"branch"`
	UpdateId       string `json:"updateId"`
	RuntimeVersion string `json:"runtimeVersion"`
	Platform       string `json:"platform"`
}

// CompleteUpdateResponse contains the result of marking an update as complete
type CompleteUpdateResponse struct {
	Status string `json:"status"` // "deployed", "identical", or "error"
}

// InitiateUpdateHandler handles the initial request to create a new update
func InitiateUpdateHandler(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()
	
	// Authenticate the request
	expoAuth := helpers.GetExpoAuth(r)
	expoAccount, err := services.FetchExpoUserAccountInformations(expoAuth)
	if err != nil {
		log.Printf("[RequestID: %s] Error fetching expo account information: %v", requestID, err)
		http.Error(w, "Error fetching expo account information", http.StatusUnauthorized)
		return
	}
	
	if expoAccount == nil {
		log.Printf("[RequestID: %s] No expo account found", requestID)
		http.Error(w, "No expo account found", http.StatusUnauthorized)
		return
	}
	
	currentExpoUsername := services.FetchSelfExpoUsername()
	if expoAccount.Username != currentExpoUsername {
		log.Printf("[RequestID: %s] Invalid expo account", requestID)
		http.Error(w, "Invalid expo account", http.StatusUnauthorized)
		return
	}
	
	// Parse the request
	var uploadRequest UpdateUploadRequest
	err = json.NewDecoder(r.Body).Decode(&uploadRequest)
	if err != nil {
		log.Printf("[RequestID: %s] Invalid request body: %v", requestID, err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate request data
	if uploadRequest.Branch == "" {
		log.Printf("[RequestID: %s] Branch name is required", requestID)
		http.Error(w, "Branch name is required", http.StatusBadRequest)
		return
	}
	
	if uploadRequest.Platform == "" || (uploadRequest.Platform != "ios" && uploadRequest.Platform != "android") {
		log.Printf("[RequestID: %s] Invalid platform: %s", requestID, uploadRequest.Platform)
		http.Error(w, "Invalid platform", http.StatusBadRequest)
		return
	}
	
	if uploadRequest.RuntimeVersion == "" {
		log.Printf("[RequestID: %s] Runtime version is required", requestID)
		http.Error(w, "Runtime version is required", http.StatusBadRequest)
		return
	}
	
	if len(uploadRequest.FileNames) == 0 {
		log.Printf("[RequestID: %s] File names are required", requestID)
		http.Error(w, "File names are required", http.StatusBadRequest)
		return
	}
	
	// Create or ensure branch exists
	err = branch.UpsertBranch(uploadRequest.Branch)
	if err != nil {
		log.Printf("[RequestID: %s] Error upserting branch: %v", requestID, err)
		http.Error(w, "Error upserting branch", http.StatusInternalServerError)
		return
	}
	
	// Generate update ID (using current time in milliseconds)
	updateId := fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond))
	
	// Request upload URLs for the files
	uploadUrls, err := bucket.RequestUploadUrlsForFileUpdates(
		uploadRequest.Branch,
		uploadRequest.RuntimeVersion,
		updateId,
		uploadRequest.FileNames,
	)
	if err != nil {
		log.Printf("[RequestID: %s] Error requesting upload URLs: %v", requestID, err)
		http.Error(w, "Error requesting upload URLs", http.StatusInternalServerError)
		return
	}
	
	// Prepare and send response
	response := UpdateUploadResponse{
		UpdateId:      updateId,
		Branch:        uploadRequest.Branch,
		Platform:      uploadRequest.Platform,
		UploadUrls:    uploadUrls,
		RuntimeVersion: uploadRequest.RuntimeVersion,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UploadFileHandler handles the upload of a single file
func UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()
	
	// Authenticate the request
	expoAuth := helpers.GetExpoAuth(r)
	expoAccount, err := services.FetchExpoUserAccountInformations(expoAuth)
	if err != nil {
		log.Printf("[RequestID: %s] Error fetching expo account information: %v", requestID, err)
		http.Error(w, "Error fetching expo account information", http.StatusUnauthorized)
		return
	}
	
	if expoAccount == nil {
		log.Printf("[RequestID: %s] No expo account found", requestID)
		http.Error(w, "No expo account found", http.StatusUnauthorized)
		return
	}
	
	currentExpoUsername := services.FetchSelfExpoUsername()
	if expoAccount.Username != currentExpoUsername {
		log.Printf("[RequestID: %s] Invalid expo account", requestID)
		http.Error(w, "Invalid expo account", http.StatusUnauthorized)
		return
	}
	
	// Get parameters from the request
	vars := mux.Vars(r)
	branch := vars["BRANCH"]
	updateId := vars["UPDATE_ID"]
	runtimeVersion := vars["RUNTIME_VERSION"]
	fileName := r.URL.Query().Get("fileName")
	
	if branch == "" || updateId == "" || runtimeVersion == "" || fileName == "" {
		log.Printf("[RequestID: %s] Missing required parameters", requestID)
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}
	
	// Parse multipart form
	err = r.ParseMultipartForm(100 << 20) // 100MB max
	if err != nil {
		log.Printf("[RequestID: %s] Error parsing multipart form: %v", requestID, err)
		http.Error(w, "Error parsing multipart form", http.StatusBadRequest)
		return
	}
	
	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("[RequestID: %s] Error retrieving file from form: %v", requestID, err)
		http.Error(w, "Error retrieving file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()
	
	// Create update object
	updateObj, err := update.GetUpdate(branch, runtimeVersion, updateId)
	if err != nil {
		log.Printf("[RequestID: %s] Error getting update: %v", requestID, err)
		http.Error(w, "Error getting update", http.StatusInternalServerError)
		return
	}
	
	// Upload file to bucket
	resolvedBucket := bucket.GetBucket()
	err = resolvedBucket.UploadFileIntoUpdate(*updateObj, fileName, file)
	if err != nil {
		log.Printf("[RequestID: %s] Error uploading file: %v", requestID, err)
		http.Error(w, "Error uploading file", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"success":true,"fileName":"%s","size":%d}`, header.Filename, header.Size)))
}

// CompleteUpdateHandler marks an update as completed
func CompleteUpdateHandler(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()
	
	// Authenticate the request
	expoAuth := helpers.GetExpoAuth(r)
	expoAccount, err := services.FetchExpoUserAccountInformations(expoAuth)
	if err != nil {
		log.Printf("[RequestID: %s] Error fetching expo account information: %v", requestID, err)
		http.Error(w, "Error fetching expo account information", http.StatusUnauthorized)
		return
	}
	
	if expoAccount == nil {
		log.Printf("[RequestID: %s] No expo account found", requestID)
		http.Error(w, "No expo account found", http.StatusUnauthorized)
		return
	}
	
	currentExpoUsername := services.FetchSelfExpoUsername()
	if expoAccount.Username != currentExpoUsername {
		log.Printf("[RequestID: %s] Invalid expo account", requestID)
		http.Error(w, "Invalid expo account", http.StatusUnauthorized)
		return
	}
	
	// Parse the request
	var completeRequest CompleteUpdateRequest
	err = json.NewDecoder(r.Body).Decode(&completeRequest)
	if err != nil {
		log.Printf("[RequestID: %s] Invalid request body: %v", requestID, err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate request data
	if completeRequest.Branch == "" {
		log.Printf("[RequestID: %s] Branch name is required", requestID)
		http.Error(w, "Branch name is required", http.StatusBadRequest)
		return
	}
	
	if completeRequest.Platform == "" || (completeRequest.Platform != "ios" && completeRequest.Platform != "android") {
		log.Printf("[RequestID: %s] Invalid platform: %s", requestID, completeRequest.Platform)
		http.Error(w, "Invalid platform", http.StatusBadRequest)
		return
	}
	
	if completeRequest.RuntimeVersion == "" {
		log.Printf("[RequestID: %s] Runtime version is required", requestID)
		http.Error(w, "Runtime version is required", http.StatusBadRequest)
		return
	}
	
	if completeRequest.UpdateId == "" {
		log.Printf("[RequestID: %s] Update ID is required", requestID)
		http.Error(w, "Update ID is required", http.StatusBadRequest)
		return
	}
	
	// Get the update
	currentUpdate, err := update.GetUpdate(completeRequest.Branch, completeRequest.RuntimeVersion, completeRequest.UpdateId)
	if err != nil {
		log.Printf("[RequestID: %s] Error getting update: %v", requestID, err)
		http.Error(w, "Error getting update", http.StatusInternalServerError)
		return
	}
	
	// Verify the update files
	resolvedBucket := bucket.GetBucket()
	errorVerify := update.VerifyUploadedUpdate(*currentUpdate)
	if errorVerify != nil {
		// Delete folder and throw error
		log.Printf("[RequestID: %s] Invalid update, deleting folder...", requestID)
		err := resolvedBucket.DeleteUpdateFolder(completeRequest.Branch, completeRequest.RuntimeVersion, completeRequest.UpdateId)
		if err != nil {
			log.Printf("[RequestID: %s] Error deleting update folder: %v", requestID, err)
			http.Error(w, "Error deleting update folder", http.StatusInternalServerError)
			return
		}
		log.Printf("[RequestID: %s] Invalid update, folder deleted", requestID)
		http.Error(w, fmt.Sprintf("Invalid update %s", errorVerify), http.StatusBadRequest)
		return
	}
	
	// Check if this update is identical to the latest one
	latestUpdate, err := update.GetLatestUpdateBundlePathForRuntimeVersion(completeRequest.Branch, completeRequest.RuntimeVersion)
	if err != nil || latestUpdate == nil {
		err = update.MarkUpdateAsChecked(*currentUpdate)
		if err != nil {
			log.Printf("[RequestID: %s] Error marking update as checked: %v", requestID, err)
			http.Error(w, "Error marking update as checked", http.StatusInternalServerError)
			return
		}
		
		response := CompleteUpdateResponse{
			Status: "deployed",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}
	
	// Compare updates
	areUpdatesIdentical, err := update.AreUpdatesIdentical(*currentUpdate, *latestUpdate, completeRequest.Platform)
	if err != nil {
		log.Printf("[RequestID: %s] Error comparing updates: %v", requestID, err)
		http.Error(w, "Error comparing updates", http.StatusInternalServerError)
		return
	}
	
	if !areUpdatesIdentical {
		err = update.MarkUpdateAsChecked(*currentUpdate)
		if err != nil {
			log.Printf("[RequestID: %s] Error marking update as checked: %v", requestID, err)
			http.Error(w, "Error marking update as checked", http.StatusInternalServerError)
			return
		}
		
		response := CompleteUpdateResponse{
			Status: "deployed",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}
	
	// Updates are identical, delete the folder
	log.Printf("[RequestID: %s] Updates are identical, delete folder...", requestID)
	err = resolvedBucket.DeleteUpdateFolder(completeRequest.Branch, completeRequest.RuntimeVersion, completeRequest.UpdateId)
	if err != nil {
		log.Printf("[RequestID: %s] Error deleting update folder: %v", requestID, err)
		http.Error(w, "Error deleting update folder", http.StatusInternalServerError)
		return
	}
	
	response := CompleteUpdateResponse{
		Status: "identical",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// BulkUploadHandler handles uploading multiple files in a single request
func BulkUploadHandler(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()
	
	// Authenticate the request
	expoAuth := helpers.GetExpoAuth(r)
	expoAccount, err := services.FetchExpoUserAccountInformations(expoAuth)
	if err != nil {
		log.Printf("[RequestID: %s] Error fetching expo account information: %v", requestID, err)
		http.Error(w, "Error fetching expo account information", http.StatusUnauthorized)
		return
	}
	
	if expoAccount == nil {
		log.Printf("[RequestID: %s] No expo account found", requestID)
		http.Error(w, "No expo account found", http.StatusUnauthorized)
		return
	}
	
	currentExpoUsername := services.FetchSelfExpoUsername()
	if expoAccount.Username != currentExpoUsername {
		log.Printf("[RequestID: %s] Invalid expo account", requestID)
		http.Error(w, "Invalid expo account", http.StatusUnauthorized)
		return
	}
	
	// Get parameters from the request
	vars := mux.Vars(r)
	branch := vars["BRANCH"]
	updateId := vars["UPDATE_ID"]
	runtimeVersion := vars["RUNTIME_VERSION"]
	
	if branch == "" || updateId == "" || runtimeVersion == "" {
		log.Printf("[RequestID: %s] Missing required parameters", requestID)
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}
	
	// Parse multipart form
	err = r.ParseMultipartForm(100 << 20) // 100MB max
	if err != nil {
		log.Printf("[RequestID: %s] Error parsing multipart form: %v", requestID, err)
		http.Error(w, "Error parsing multipart form", http.StatusBadRequest)
		return
	}
	
	// Create update object
	updateObj, err := update.GetUpdate(branch, runtimeVersion, updateId)
	if err != nil {
		log.Printf("[RequestID: %s] Error getting update: %v", requestID, err)
		http.Error(w, "Error getting update", http.StatusInternalServerError)
		return
	}
	
	// Get all files from form
	form := r.MultipartForm
	files := form.File
	
	type UploadResult struct {
		FileName string `json:"fileName"`
		Success  bool   `json:"success"`
		Error    string `json:"error,omitempty"`
	}
	
	results := make([]UploadResult, 0, len(files))
	resolvedBucket := bucket.GetBucket()
	
	for formField, fileHeaders := range files {
		for _, fileHeader := range fileHeaders {
			fileName := fileHeader.Filename
			if formField != "files[]" {
				// If formField is not the default "files[]", use it as the fileName
				fileName = formField
			}
			
			file, err := fileHeader.Open()
			if err != nil {
				results = append(results, UploadResult{
					FileName: fileName,
					Success:  false,
					Error:    fmt.Sprintf("Error opening file: %v", err),
				})
				continue
			}
			
			err = resolvedBucket.UploadFileIntoUpdate(*updateObj, fileName, file)
			file.Close()
			
			if err != nil {
				results = append(results, UploadResult{
					FileName: fileName,
					Success:  false,
					Error:    fmt.Sprintf("Error uploading file: %v", err),
				})
				continue
			}
			
			results = append(results, UploadResult{
				FileName: fileName,
				Success:  true,
			})
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
	})
} 