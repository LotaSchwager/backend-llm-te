package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"ollama-backend/pkg/drive"
	"ollama-backend/pkg/excel"
	"ollama-backend/pkg/ollama"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// =========== Struct para el metodo de Ollama ===============

// GenerateRequest estructura para las peticiones de generación
type GenerateRequest struct {
	Prompt string `json:"prompt"`
}

// Contenedor de la estructura de cada modelo
type ModelResult struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at,omitempty"`
	Message            string    `json:"message,omitempty"`
	Done               bool      `json:"done,omitempty"`
	TotalDuration      int64     `json:"total_duration,omitempty"`
	LoadDuration       int       `json:"load_duration,omitempty"`
	PromptEvalCount    int       `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int       `json:"prompt_eval_duration,omitempty"`
	EvalCount          int       `json:"eval_count,omitempty"`
	EvalDuration       int64     `json:"eval_duration,omitempty"`
	Error              string    `json:"error,omitempty"`
}

// Respuesta de ollama con multiples respuestas de multiples modelos
type MultiResponse struct {
	Status    int           `json:"status"`
	Responses []ModelResult `json:"responses"`
}

// ===========================================================

// =========== Struct para el metodo drive y excel ===========

// GenerateResponse estructura para las respuestas de generación
type GenerateResponse struct {
	Status  int      `json:"status"`
	Content []string `json:"content"`
}

// ExcelRequest estructura para las peticiones de creación de Excel
type ExcelRequest struct {
	Interactions []excel.Interaction `json:"interactions"`
}

// ============================================================

// ============================================================

type PongResponse struct {
	Status  int    `json:"status"`
	Content string `json:"content"`
}

// ============================================================

func main() {
	// Crear directorio de salida para archivos Excel
	outputDir := "output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	// Inicializar servicios
	ollamaService := ollama.NewService()
	excelService := excel.NewService(outputDir)

	// Inicializar servicio de Google Drive si existe el archivo de credenciales
	var driveService *drive.Service
	credentialsFile := "credentials.json"
	if _, err := os.Stat(credentialsFile); err == nil {
		var err error
		driveService, err = drive.NewService(credentialsFile)
		if err != nil {
			fmt.Printf("Error initializing Google Drive service: %v\n", err)
		}
	}

	// Modo release
	gin.SetMode(gin.ReleaseMode)

	// Configurar el router de Gin
	r := gin.Default()

	// Configurar CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"https://prototipo-te-ws1t.vercel.app/"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Custom-Header"}
	config.AllowCredentials = true

	// Usar CORS
	r.Use(cors.New(config))

	// Endpoint para generación de respuestas, usando multitarea
	r.POST("/generate", func(c *gin.Context) {
		var req GenerateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, MultiResponse{
				Status: http.StatusBadRequest,
				Responses: []ModelResult{{
					Model: "Todos",
					Error: "Formato invalido",
				}},
			})
			return
		}

		// Obtener los modelos desde el servicio
		modelNames := ollamaService.GetModelNames()
		if len(modelNames) == 0 {
			c.JSON(http.StatusInternalServerError, MultiResponse{
				Status: http.StatusInternalServerError,
				Responses: []ModelResult{{
					Model: "Todos",
					Error: "Error al cargar los modelos validos",
				}},
			})
			return
		}

		// Desde aqui se usara paralelismo con canales
		var wg sync.WaitGroup

		// Canal para realizar la cantidad de request equivalente a la cantidad de modelos
		resultsChan := make(chan ModelResult, len(modelNames))

		// Recorrer cada modelo dentro de validModels
		for _, modelNameCurrent := range modelNames {

			// Incrementar el contador de waitgroup en 1
			wg.Add(1)

			// Crear una gorouting
			go func(modelKeyName string, currentPrompt string) {

				//fmt.Printf("\nModelo: %s\n", modelKeyName)

				// Terminar y bajar el contador de wg cuando la funcion termine
				defer wg.Done()

				// Llamar al servicio de ollama
				content, err := ollamaService.GenerateResponse(modelKeyName, currentPrompt)

				result := ModelResult{Model: modelKeyName}
				if err != nil {
					result.Error = err.Error()
				} else {
					result.CreatedAt = content.CreatedAt
					result.Message = content.Message
					result.Done = content.Done
					result.TotalDuration = content.TotalDuration
					result.LoadDuration = content.LoadDuration
					result.PromptEvalCount = content.PromptEvalCount
					result.PromptEvalDuration = content.PromptEvalDuration
					result.EvalCount = content.EvalCount
					result.EvalDuration = content.EvalDuration
				}

				//fmt.Printf("\nTermino el modelo: %s\n", modelKeyName)
				//fmt.Printf("\nModelo: %s ===> %s\n", modelKeyName, result.Content)

				// Retornar el resultado al canal
				resultsChan <- result
			}(modelNameCurrent, req.Prompt)
		}

		// Crear otra goroutine par cerrar los canales
		go func() {
			wg.Wait()
			close(resultsChan)
		}()

		// Recolectar los resultados
		var collectedResponses []ModelResult
		for res := range resultsChan {
			collectedResponses = append(collectedResponses, res)
		}

		// Enviar los resultados
		c.JSON(http.StatusOK, MultiResponse{
			Status:    http.StatusOK,
			Responses: collectedResponses,
		})
	})

	// Endpoint para creación de archivos Excel
	r.POST("/excel", func(c *gin.Context) {
		var req ExcelRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  400,
				"message": "Invalid request format",
			})
			return
		}

		// Crear archivo Excel
		filepath, err := excelService.CreateExcelFile(req.Interactions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  500,
				"message": fmt.Sprintf("Error creating Excel file: %v", err),
			})
			return
		}

		// Intentar subir a Google Drive si el servicio está disponible
		llmFolder, err := driveService.FindFolderIdByName("llm")
		if err != nil {
			fmt.Printf("Error finding folder: %v\n", err)
			// Continuar solo con el archivo local
		}

		if driveService != nil {
			fileId, err := driveService.UploadFile(filepath, llmFolder)
			if err != nil {
				fmt.Printf("Error uploading to Google Drive: %v\n", err)
				// Continuar solo con el archivo local
			} else {
				c.JSON(http.StatusOK, gin.H{
					"status":   200,
					"message":  "File created and uploaded to Google Drive",
					"filepath": filepath,
					"fileId":   fileId,
				})
				return
			}
		}

		// Retornar ruta del archivo local si la subida a Google Drive falló o no estaba disponible
		c.JSON(http.StatusOK, gin.H{
			"status":   200,
			"message":  "File created locally",
			"filepath": filepath,
		})
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, PongResponse{
			Status:  http.StatusOK,
			Content: "pong",
		})
	})

	fmt.Println("Server starting on :8080")
	r.Run(":8080")
}
