package main

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
)

type ErrorResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message"`
}

type DataResponse struct {
	Data any `json:"data"`
}

type FileCreated struct {
	Filename string `json:"filename"`
	Message  string `json:"message"`
	Link     string `json:"link"`
}

type UploadedFile struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Url  string `json:"url"`
}

// Create a web server that listens on port 5890 using echo framework
func main() {
	e := echo.New()
	e.HidePort = true
	e.HideBanner = true
	//log := log.New()
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {
			log.WithFields(log.Fields{
				"URI":    values.URI,
				"status": values.Status,
			}).Info("request")

			return nil
		},
	}))

	e.Use(middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
		Validator: func(username, password string, c echo.Context) (bool, error) {
			if username == "william" && password == "3EzV8osmCog4dfp2CPj2" {
				return true, nil
			}

			return false, nil
		},
		Skipper: func(c echo.Context) bool {
			return c.Request().RequestURI != "/" && c.Request().Method == "GET"
		},
	}))

	log.Infof("Starting server on port 5890")

	e.GET("/", func(c echo.Context) error {
		// Get a list of files in the current directory
		files, err := os.ReadDir("images/")
		if err != nil {
			log.Error("Error reading directory")
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   err.Error(),
				Message: "Error reading directory",
			})
		}

		// Create a slice of strings to hold the filenames
		var filenames []UploadedFile
		for _, file := range files {
			// Append the filename to the slice
			filenames = append(filenames, UploadedFile{
				Name: file.Name(),
				Type: mime.TypeByExtension(filepath.Ext(file.Name())),
				Url:  "http://localhost:5890/" + file.Name(),
			})
		}

		// Return the filenames as a JSON response
		return c.JSON(http.StatusOK, filenames)
	})

	e.GET("/:file_name", func(c echo.Context) error {
		fileName := c.Param("file_name")
		filePath := filepath.Join("images", fileName)

		// Check if file exists on the server
		_, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
		if err != nil {
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Message: "File not found",
			})
		}

		// Return the file to the client
		return c.File(filePath)
	})

	// Create a route which allows images to be uploaded
	e.POST("/", func(c echo.Context) error {
		// Read the file from the request
		file, err := c.FormFile("file")
		if err != nil {
			log.Errorf("Error reading file: %v", err)
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   err.Error(),
				Message: "Error reading file",
			})
		}

		fileType := filepath.Ext(file.Filename)
		mimeType := mime.TypeByExtension(fileType)

		if !isAllowedFile(fileType) {
			log.WithFields(
				log.Fields{
					"file_name": file.Filename,
					"file_type": fileType,
					"mime_type": mimeType,
				},
			).Error("Tried to upload a file that is not an image/video")

			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Message: "File is not an image",
			})
		}

		// Get the bytes from the file
		src, err := file.Open()
		if err != nil {
			log.Errorf("Error opening file: %v", err)
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   err.Error(),
				Message: "Error opening file",
			})
		}

		defer func(src multipart.File) {
			err := src.Close()
			if err != nil {
				log.Errorf("Error closing file: %v", err)
			}
		}(src)

		// Generate a new file name which is only 6 characters long

		newFilename, err := GenerateRandomString(6)
		if err != nil {
			log.Errorf("Error generating random string: %v", err)
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   err.Error(),
				Message: "Error generating random string",
			})
		}

		// Append the file extension to the new file name
		newFilename = newFilename + fileType

		// Create a new file
		dst, err := os.Create("images/" + newFilename)
		if err != nil {
			log.Errorf("Error creating file: %v", err)
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   err.Error(),
				Message: "Error creating file",
			})
		}

		defer func(dst *os.File) {
			err := dst.Close()
			if err != nil {
				log.Errorf("Error closing file: %v", err)
			}
		}(dst)

		// Copy the bytes from the file to the new file
		if _, err = io.Copy(dst, src); err != nil {
			log.Errorf("Error copying file: %v", err)
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   err.Error(),
				Message: "Error copying file",
			})
		}

		return c.JSON(http.StatusCreated, DataResponse{FileCreated{
			Filename: newFilename,
			Message:  "File uploaded successfully",
			Link:     "http://localhost:5890/" + newFilename,
		}})
	})

	e.Logger.Fatal(e.Start(":5890"))
}

// GenerateRandomString Randomly generate a string with i characters
func GenerateRandomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func isAllowedFile(fileType string) bool {
	mimeType := mime.TypeByExtension(fileType)

	if mimeType == "" {
		return false
	}

	allowedMimeTypes := []string{"image", "video"}

	for _, allowedMimeType := range allowedMimeTypes {
		if strings.Contains(mimeType, allowedMimeType) {
			return true
		}
	}

	return false
}
