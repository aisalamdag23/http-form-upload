package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	db  *sql.DB
	err error
)

type uploaded_image_metadata struct {
	fileName    string
	filePath    string
	contentType string
	size        int
	ipAddress   string
	userAgent   string
}

func main() {
	// load env
	log.Println("Read env files")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Unable to load .env file")
	}
	// setup db
	log.Println("Connect sql DB")
	dbConn()

	// setup routes
	http.HandleFunc("/", showForm)
	http.HandleFunc("/upload", upload)

	log.Println("Starting server...")
	http.ListenAndServe(":8080", nil)
}

func dbConn() {
	dbURL := os.Getenv("DATABASE_URL")
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("Error verifying connection to the database: %v", err)
	}
}

func showForm(w http.ResponseWriter, req *http.Request) {
	tmpl, err := template.ParseFiles("form.html")
	if err != nil {
		http.Error(w, "Unable to load form", http.StatusInternalServerError)
		return
	}

	formAuthValue := struct {
		AuthToken string
	}{
		AuthToken: os.Getenv("AUTH_TOKEN"),
	}

	tmpl.Execute(w, formAuthValue)
}

func upload(w http.ResponseWriter, req *http.Request) {
	// ensure post method is used to satisfy the requirement:
	// The form should POST data to the /upload handler, which should write the received file data to a temporary file.
	if req.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	// check image size ensure to only accept images that are less than or equal to 8MB
	if !isImageSizeLessThanOrEqualTo8MB(w, req) {
		http.Error(w, "Invalid request: file size too large", http.StatusForbidden)
		return
	}
	// get the file from the form
	uploadedFile, uploadedFileHeader, err := req.FormFile("data")
	if err != nil {
		log.Println("Error getting file from form: ", err)
		http.Error(w, "Invalid file. Check logs", http.StatusForbidden)
		return
	}
	defer uploadedFile.Close()

	// check if the uploaded file is an image and if the auth token matches env auth token
	if !isAnImage(uploadedFileHeader) || !doesAuthTokenMatch(req.FormValue("auth")) {
		http.Error(w, "Invalid request: file is not an image or invalid auth token", http.StatusForbidden)
		return
	}

	// write the received file data to a temporary file
	filePath, err := writeReceivedFileToTempFile(uploadedFile, uploadedFileHeader)
	if err != nil {
		log.Println("Error writing file to temp file: ", err)
		http.Error(w, "Unable to write file. Check logs", http.StatusInternalServerError)
		return
	}
	// write the image metadata (content type, size, etc) to a database of your choice, including all relevant HTTP information.
	err = recordUploadedImage(uploaded_image_metadata{
		fileName:    uploadedFileHeader.Filename,
		filePath:    *filePath,
		contentType: uploadedFileHeader.Header.Get("Content-Type"),
		size:        int(uploadedFileHeader.Size),
		ipAddress:   req.RemoteAddr,
		userAgent:   req.UserAgent(),
	})
	if err != nil {
		log.Println("Error recording uploaded file: ", err)
		http.Error(w, "Unable to save upload. Check logs", http.StatusInternalServerError)
		return
	}

	// return a success message
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File uploaded successfully"))
}

// isImageSizeLessThanOrEqualTo8MB checks that the size of the uploaded file is less than or equal to 8MB
func isImageSizeLessThanOrEqualTo8MB(w http.ResponseWriter, req *http.Request) bool {
	// get the max upload size from the environment variable
	maxUploadSizeStr := os.Getenv("MAX_UPLOAD_SIZE")
	maxUploadSizeInt, err := strconv.Atoi(maxUploadSizeStr)
	if err != nil {
		// default to 8MB if MAX_UPLOAD_SIZE is not set
		maxUploadSizeInt = 8388608
	}
	maxUploadSize := int64(maxUploadSizeInt)

	req.Body = http.MaxBytesReader(w, req.Body, maxUploadSize)

	if err := req.ParseForm(); err != nil {
		log.Println("Error parsing form: ", err)
		return false
	}
	return true
}

// isAnImage checks that the content type of the uploaded file is an image
func isAnImage(uploadedFileHeader *multipart.FileHeader) bool {
	fileContentType := uploadedFileHeader.Header.Get("Content-Type")

	log.Println("fileContentType: ", fileContentType)

	return strings.HasPrefix(strings.ToLower(fileContentType), "image/")
}

// doesAuthTokenMatch checks that auth token matches
func doesAuthTokenMatch(authTokenRequest string) bool {
	return authTokenRequest == os.Getenv("AUTH_TOKEN")
}

// writeReceivedFileToTempFile writes the received file to a temporary file
func writeReceivedFileToTempFile(uploadedFile multipart.File, uploadedFileHeader *multipart.FileHeader) (*string, error) {
	uploadDir := os.Getenv("UPLOAD_DIR")

	// Create the uploads directory if it doesn't exist
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, os.ModePerm)
	}

	// Create the destination file
	uploadDst := fmt.Sprintf("%s/%s", uploadDir, uploadedFileHeader.Filename)
	dst, err := os.Create(uploadDst)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	// Copy the uploaded file to the destination
	if _, err := io.Copy(dst, uploadedFile); err != nil {
		return nil, err
	}
	return &uploadDst, nil
}

// recordUploadedImage writes the image metadata to the database
func recordUploadedImage(metadata uploaded_image_metadata) error {
	_, err := db.Exec("INSERT INTO uploaded_image_metadata (filename, filepath, content_type, size, ip_address, user_agent) VALUES ($1, $2, $3, $4, $5, $6)", metadata.fileName, metadata.filePath, metadata.contentType, metadata.size, metadata.ipAddress, metadata.userAgent)
	if err != nil {
		return err
	}

	return nil
}
