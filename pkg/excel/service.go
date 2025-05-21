package excel

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/xuri/excelize/v2"
)

// Answer representa una respuesta individual de un modelo
type Answer struct {
	Model      string  `json:"model"`
	Answer     string  `json:"answer"`
	IsSelected bool    `json:"is_selected"`
	TimeSpend  float64 `json:"time_spend"`
}

// Interaction representa una interacción completa con múltiples respuestas
type Interaction struct {
	Question string   `json:"question"`
	Name     string   `json:"name"`
	Answers  []Answer `json:"answers"`
}

// Service maneja la creación y manipulación de archivos Excel
type Service struct {
	outputDir string
}

// NewService crea una nueva instancia del servicio de Excel
// Requiere un directorio de salida donde se guardarán los archivos
func NewService(outputDir string) *Service {
	return &Service{
		outputDir: outputDir,
	}
}

// CreateExcelFile crea un archivo Excel con las interacciones proporcionadas
// Retorna la ruta del archivo creado o un error si algo falla
func (s *Service) CreateExcelFile(interactions []Interaction) (string, error) {
	// Crear un nuevo archivo Excel
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// Crear encabezados
	headers := []string{"Name", "Question", "Model", "Answer", "Selected", "Time Spent (s)"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue("Sheet1", cell, header)
	}

	// Agregar datos
	row := 2
	for _, interaction := range interactions {
		for _, answer := range interaction.Answers {
			// Escribir cada campo en su columna correspondiente
			f.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), interaction.Name)
			f.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), interaction.Question)
			f.SetCellValue("Sheet1", fmt.Sprintf("C%d", row), answer.Model)
			f.SetCellValue("Sheet1", fmt.Sprintf("D%d", row), answer.Answer)
			f.SetCellValue("Sheet1", fmt.Sprintf("E%d", row), answer.IsSelected)
			f.SetCellValue("Sheet1", fmt.Sprintf("F%d", row), answer.TimeSpend)
			row++
		}
	}

	// Generar nombre de archivo con timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.xlsx", interactions[0].Name, timestamp)
	filepath := filepath.Join(s.outputDir, filename)

	// Guardar el archivo
	if err := f.SaveAs(filepath); err != nil {
		return "", fmt.Errorf("error saving excel file: %v", err)
	}

	return filepath, nil
}
