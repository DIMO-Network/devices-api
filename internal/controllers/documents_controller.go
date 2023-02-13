package controllers

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
)

type DocumentsController struct {
	settings *config.Settings
	s3Client *s3.Client
	DBS      func() *db.ReaderWriter
	logger   *zerolog.Logger
}

// NewDocumentsController constructor
func NewDocumentsController(settings *config.Settings, z *zerolog.Logger, s3Client *s3.Client, dbs func() *db.ReaderWriter) DocumentsController {
	return DocumentsController{settings: settings, s3Client: s3Client, DBS: dbs, logger: z}
}

// GetDocuments godoc
// @Description gets all documents associated with current user - pulled from token
// @Tags        documents
// @Produce     json
// @Accept      json
// @Success     200 {object} []controllers.DocumentResponse
// @Security    BearerAuth
// @Router      /documents [get]
func (udc *DocumentsController) GetDocuments(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	udi := c.Query("user_device_id")

	folder := userID
	if len(udi) > 0 {
		folder = fmt.Sprintf("%s/%s", userID, udi)
	}

	response, err := udc.s3Client.ListObjectsV2(c.Context(), &s3.ListObjectsV2Input{
		Bucket: aws.String(udc.settings.AWSDocumentsBucketName),
		Prefix: aws.String(folder),
	})

	documents := []DocumentResponse{}

	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "the bucket does not exist")
	}

	for _, item := range response.Contents {
		responseItem, err := udc.s3Client.GetObject(c.Context(),
			&s3.GetObjectInput{
				Bucket: aws.String(udc.settings.AWSDocumentsBucketName),
				Key:    aws.String(*item.Key),
			})

		if err != nil {
			return helpers.ErrorResponseHandler(c, err, fiber.StatusInternalServerError)
		}

		documents = append(documents, DocumentResponse{
			ID:           responseItem.Metadata[MetadataDocumentID],
			Name:         responseItem.Metadata[MetadataDocumentName],
			Ext:          responseItem.Metadata[MetadataDocumentFileExtension],
			UserDeviceID: responseItem.Metadata[MetadataDocumentUserDeviceID],
			CreatedAt:    *responseItem.LastModified,
			URL:          fmt.Sprintf("%s/v1/documents/%s/download", udc.settings.DeploymentBaseURL, responseItem.Metadata[MetadataDocumentID]),
			Type:         DocumentTypeEnum(responseItem.Metadata[MetadataDocumentType]),
		})

	}

	return c.JSON(documents)
}

// GetDocumentByID godoc
// @Description get document by id associated with current user - pulled from token
// @Tags        documents
// @Produce     json
// @Accept      json
// @Param       id  path     string true "Document ID"
// @Success     200 {object} controllers.DocumentResponse
// @Security    BearerAuth
// @Router      /documents/{id} [get]
func (udc *DocumentsController) GetDocumentByID(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	fileID := c.Params("id")
	folder := resolveFolderKey(userID, fileID)

	response, err := udc.s3Client.GetObject(c.Context(),
		&s3.GetObjectInput{
			Bucket: aws.String(udc.settings.AWSDocumentsBucketName),
			Key:    aws.String(folder),
		})

	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("no document with id %s found", fileID))
	}

	document := DocumentResponse{
		ID:           response.Metadata[MetadataDocumentID],
		Name:         response.Metadata[MetadataDocumentName],
		Ext:          response.Metadata[MetadataDocumentFileExtension],
		UserDeviceID: response.Metadata[MetadataDocumentUserDeviceID],
		CreatedAt:    *response.LastModified,
		URL:          fmt.Sprintf("%s/v1/documents/%s/download", udc.settings.DeploymentBaseURL, response.Metadata[MetadataDocumentID]),
		Type:         DocumentTypeEnum(response.Metadata[MetadataDocumentType]),
	}

	return c.JSON(document)
}

// PostDocument godoc
// @Description post document by id associated with current user - pulled from token
// @Tags        documents
// @Produce     json
// @Accept      multipart/form-data
// @Param       file         formData file   true  "The file to upload. file is required"
// @Param       name         formData string true  "The document name. name is required"
// @Param       type         formData string true  "The document type. type is required"
// @Param       userDeviceID formData string false "The user device ID, optional"
// @Success     201          {object} controllers.DocumentResponse
// @Security    BearerAuth
// @Router      /documents [post]
func (udc *DocumentsController) PostDocument(c *fiber.Ctx) error {

	userID := helpers.GetUserID(c)
	file, err := c.FormFile("file")
	documentName := c.FormValue("name")
	documentType := c.FormValue("type")
	udi := c.FormValue("userDeviceID")

	if len(udi) > 0 {
		// Validate if user devices exists
		exists, err := models.UserDevices(
			models.UserDeviceWhere.UserID.EQ(userID),
			models.UserDeviceWhere.ID.EQ(udi),
		).Exists(c.Context(), udc.DBS().Writer)

		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		if !exists {
			return fiber.NewError(fiber.StatusNotFound, "Device not found.")
		}
	}

	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid file.")
	}

	if err := DocumentTypeEnum(documentType).IsValid(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid document type.")
	}
	// Get Buffer from file
	fileObj, err := file.Open()

	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "document cannot be read.")
	}
	defer fileObj.Close()

	// Validate file type
	filetype := file.Header.Get("content-type")

	if err := FileTypeAllowedEnum(filetype).IsValid(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "the provided file format is not allowed.")
	}

	// Create an uploader with the session and default options
	uploader := manager.NewUploader(udc.s3Client)

	// Unique Id
	id := ksuid.New().String()
	documentID := buildUniqueID(id, udi)

	metadata := map[string]string{}
	metadata[MetadataDocumentID] = documentID
	metadata[MetadataDocumentName] = documentName
	metadata[MetadataDocumentFile] = file.Filename
	metadata[MetadataDocumentFileExtension] = filepath.Ext(file.Filename)
	metadata[MetadataDocumentType] = documentType
	metadata[MetadataDocumentUserDeviceID] = udi

	fileID := buildFileID(udi, id)
	awsPathKey := getAwsFilePath(userID, fileID)

	// Upload the file to S3.
	result, err := uploader.Upload(c.Context(), &s3.PutObjectInput{
		Bucket:             aws.String(udc.settings.AWSDocumentsBucketName),
		Key:                aws.String(awsPathKey),
		Body:               fileObj,
		ContentDisposition: aws.String("attachment"),
		ContentType:        aws.String(filetype),
		Metadata:           metadata,
	})

	if err != nil {
		udc.logger.Err(err).Msg("failed to upload glovebox document")
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	_ = result

	url := fmt.Sprintf("%s/v1/documents/%s/download", udc.settings.DeploymentBaseURL, id)
	if len(udi) > 0 {
		url = fmt.Sprintf("%s/v1/documents/%s-%s/download", udc.settings.DeploymentBaseURL, udi, id)
	}
	udc.logger.Info().Msgf("succesfully uploaded glovebox document %s", documentName)
	return c.JSON(DocumentResponse{
		ID:           id,
		Name:         documentName,
		Ext:          filepath.Ext(file.Filename),
		CreatedAt:    time.Now().UTC(),
		UserDeviceID: udi,
		URL:          url,
		Type:         DocumentTypeEnum(documentType),
	})
}

// DeleteDocument godoc
// @Description delete document associated with current user - pulled from token
// @Tags        documents
// @Produce     json
// @Accept      json
// @Param       id path string true "Document ID"
// @Success     204
// @Security    BearerAuth
// @Router      /documents/{id} [delete]
func (udc *DocumentsController) DeleteDocument(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	fileID := c.Params("id")
	folder := resolveFolderKey(userID, fileID)

	_, err := udc.s3Client.DeleteObject(c.Context(), &s3.DeleteObjectInput{
		Bucket: aws.String(udc.settings.AWSDocumentsBucketName),
		Key:    aws.String(folder),
	})

	if err != nil {
		return helpers.ErrorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// DownloadDocument godoc
// @Description download document associated with current user - pulled from token
// @Tags        documents
// @Produce     octet-stream
// @Produce     png
// @Produce     jpeg
// @Param       id path string true "Document ID"
// @Success     200
// @Security    BearerAuth
// @Router      /documents/{id}/download [get]
func (udc *DocumentsController) DownloadDocument(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	fileID := c.Params("id")
	folder := resolveFolderKey(userID, fileID)

	buffer := manager.NewWriteAtBuffer([]byte{})

	downloader := manager.NewDownloader(udc.s3Client)

	numBytes, err := downloader.Download(c.Context(), buffer,
		&s3.GetObjectInput{
			Bucket: aws.String(udc.settings.AWSDocumentsBucketName),
			Key:    aws.String(folder),
		})

	if err != nil {
		return helpers.ErrorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	if numBytes == 0 {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("no document with id %s found", fileID))
	}

	data := buffer.Bytes()

	return c.SendStream(bytes.NewReader(data))
}

func getAwsFilePath(userID, fileID string) string {
	return fmt.Sprintf("%s/%s", userID, fileID)
}

// Build Unique ID
func buildUniqueID(id string, userDeviceID string) string {
	if userDeviceID != "" {
		return fmt.Sprintf("%s-%s", userDeviceID, id)
	}
	return id
}

// Build file ID
func buildFileID(userDeviceID string, uniqueID string) string {
	if userDeviceID != "" {
		return fmt.Sprintf("%s/%s", userDeviceID, uniqueID)
	}
	return uniqueID
}

func resolveFolderKey(userID string, fileID string) string {
	return fmt.Sprintf("%s/%s", userID, strings.Replace(fileID, "-", "/", 1))
}

type DocumentResponse struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	URL          string           `json:"url"`
	Ext          string           `json:"ext"`
	UserDeviceID string           `json:"userDeviceId"`
	CreatedAt    time.Time        `json:"createdAt"`
	Type         DocumentTypeEnum `json:"type"`
}

type FileTypeAllowedEnum string

const (
	jpeg FileTypeAllowedEnum = "image/jpeg"
	png  FileTypeAllowedEnum = "image/png"
	pdf  FileTypeAllowedEnum = "application/pdf"
)

func (r FileTypeAllowedEnum) IsValid() error {
	switch r {
	case jpeg, png, pdf:
		return nil
	}
	return errors.New("invalid document type")
}

type DocumentTypeEnum string

const (
	DriversLicense      DocumentTypeEnum = "DriversLicense"
	Other               DocumentTypeEnum = "Other"
	VehicleTitle        DocumentTypeEnum = "VehicleTitle"
	VehicleRegistration DocumentTypeEnum = "VehicleRegistration"
	VehicleInsurance    DocumentTypeEnum = "VehicleInsurance"
	VehicleMaintenance  DocumentTypeEnum = "VehicleMaintenance"
	VehicleCustomImage  DocumentTypeEnum = "VehicleCustomImage"
)

func (r DocumentTypeEnum) String() string {
	return string(r)
}

func (r DocumentTypeEnum) IsValid() error {
	switch r {
	case DriversLicense, VehicleMaintenance, VehicleRegistration, VehicleInsurance, VehicleTitle, Other, VehicleCustomImage:
		return nil
	}
	return errors.New("invalid document type")
}

const (
	MetadataDocumentID            = "document-id"
	MetadataDocumentName          = "document-name"
	MetadataDocumentFile          = "document-file"
	MetadataDocumentType          = "document-type"
	MetadataDocumentFileExtension = "document-file-ext"
	MetadataDocumentUserDeviceID  = "document-user-device-id"
)
