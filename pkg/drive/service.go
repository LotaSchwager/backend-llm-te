package drive

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Service maneja la interacción con Google Drive API
type Service struct {
	service *drive.Service
}

// NewService crea una nueva instancia del servicio de Google Drive
// Requiere un archivo de credenciales JSON válido para autenticación
func NewService(credentialsFile string) (*Service, error) {
	ctx := context.Background()

	// Leer el archivo de credenciales
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %v", err)
	}

	// Crear configuración de credenciales
	config, err := google.JWTConfigFromJSON(b, drive.DriveFileScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	// Crear el servicio de Drive
	srv, err := drive.NewService(ctx, option.WithHTTPClient(config.Client(ctx)))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Drive client: %v", err)
	}

	return &Service{service: srv}, nil
}

// FindFolderIdByName busca un ID de carpeta por su nombre
func (s *Service) FindFolderIdByName(folderName string) (string, error) {
	query := fmt.Sprintf("mimeType='application/vnd.google-apps.folder' and name='%s' and trashed=false", folderName)
	fileList, err := s.service.Files.List().Q(query).Fields("files(id, name)").Do()
	if err != nil {
		return "", fmt.Errorf("unable to find folder: %v", err)
	}

	for _, file := range fileList.Files {
		if file.Name == folderName {
			return file.Id, nil
		}
	}

	return "", fmt.Errorf("folder '%s' not found", folderName)
}

// UploadFile sube un archivo a Google Drive y retorna el ID del archivo
func (s *Service) UploadFile(filepath string, parentFolderId string) (string, error) {
	// Abrir el archivo
	f, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("unable to open file: %v", err)
	}
	defer f.Close()

	// Obtener información del archivo
	fileInfo, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("unable to get file info: %v", err)
	}

	// Crear metadatos del archivo
	file := &drive.File{
		Name:    fileInfo.Name(),
		Parents: []string{parentFolderId}, // Especificar la carpeta padre
	}

	// Subir el archivo
	res, err := s.service.Files.Create(file).Media(f).Do()
	if err != nil {
		return "", fmt.Errorf("unable to upload file: %v", err)
	}

	return res.Id, nil
}
