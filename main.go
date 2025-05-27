package main

import (
	"fmt"
	"log"
	"net/http"
	"ollama-backend/pkg/db"
	"ollama-backend/pkg/ollama"
	"os"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// =========== Struct para el metodo de Respuesta ===============

// GenerateRequest estructura para las peticiones de generación
type GenerateRequest struct {
	Prompt string `json:"prompt"`
}

type ModelResult struct {
	Model    string `json:"model"`
	Response string `json:"response,omitempty"`
	Error    string `json:"error,omitempty"`
	ID       int64  `json:"id,omitempty"`
}

// Respuesta de ollama con multiples respuestas de multiples modelos
type MultiResponse struct {
	Status    int           `json:"status"`
	Responses []ModelResult `json:"responses,omitempty"`
}

// ===========================================================

// =========== Struct para el metodo de Resultado ===============

type ResultadoResponse struct {
	Status  int    `json:"status"`
	Content string `json:"content"`
}

// ============================================================

// ============================================================

type PongResponse struct {
	Status  int    `json:"status"`
	Content string `json:"content"`
}

// ============================================================

func main() {
	// Inicializar servicios
	ollamaService := ollama.NewService()

	// Cargar variables de entorno
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Error al cargar el archivo .env")
		return
	}

	// Obtener variables de entorno
	user := os.Getenv("USER_DB")
	password := os.Getenv("PASSWORD")
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	dbname := os.Getenv("DB_NAME")

	// Inicializar servicio de DB
	dbservice := db.NewService(&user, &password, &host, &port, &dbname)

	// Conexion con la base de datos
	ctx, err := dbservice.ConnectDB()
	if err != nil {
		log.Fatalln("Error al conectar con la base de datos")
		return
	}
	log.Println("Conectado con la base de datos......")

	// Configurar el router de Gin
	r := gin.Default()

	// Configurar CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Custom-Header"}
	// config.AllowCredentials = true

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

		// Imprimir el prompt
		//log.Println(req.Prompt)

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
		resultsChan := make(chan db.Respuesta, len(modelNames))

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
				result := db.Respuesta{Model: modelKeyName}
				if err != nil {
					log.Fatalln("Error: $w", err)
				} else {
					result.Done = content.Done
					result.Model = content.Model
					result.Message = content.Message
					result.EvalCount = content.EvalCount
					result.EvalDuration = content.EvalDuration
					result.LoadDuration = content.LoadDuration
					result.TotalDuration = content.TotalDuration
					result.PromptEvalCount = content.PromptEvalCount
					result.PromptEvalDuration = content.PromptEvalDuration
					result.Modelo_id = content.ID
				}
				//fmt.Printf("\nTermino el modelo: %s\n", modelKeyName)
				//fmt.Printf("\nModelo: %s ===> %s\n", modelKeyName, result.Content)

				// Retornar el resultado al canal
				resultsChan <- result
			}(modelNameCurrent, req.Prompt)
		}

		// Esperar a que terminen las goroutines
		//log.Println("Esperando a que terminen las goroutines")

		// Crear otra goroutine par cerrar los canales
		go func() {
			wg.Wait()
			close(resultsChan)
		}()

		// Obtener la lista en formato respuesta
		var respuestas_struct []db.Respuesta
		for res := range resultsChan {
			respuestas_struct = append(respuestas_struct, res)
		}

		// Imprimir el numero de respuestas
		log.Println("Termino de obtener las respuestas")

		// Recolectar los resultados
		ids, err := dbservice.InsertRespuesta(ctx, &respuestas_struct)
		if err != nil {
			//log.Fatalln(err)
			c.JSON(http.StatusConflict, MultiResponse{
				Status:    http.StatusConflict,
				Responses: []ModelResult{},
			})
			return
		}

		log.Println("Termino de insertar las respuestas")

		// Obtener los valores para devolverlos
		var multiResponse []ModelResult
		for _, val := range ids.Modelos {
			value := ModelResult{
				Model:    val.Model,
				Response: val.Message,
				ID:       val.Id,
			}
			multiResponse = append(multiResponse, value)
		}
		log.Println("Termino de obtener los valores para devolver")
		// Enviar los resultados
		c.JSON(http.StatusOK, MultiResponse{
			Status:    http.StatusOK,
			Responses: multiResponse,
		})
	})

	// Endpoint para guardar el resultado
	r.POST("/save-result", func(c *gin.Context) {
		var req db.Resultado
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ResultadoResponse{
				Status:  http.StatusBadRequest,
				Content: "Formato invalido",
			})
			return
		}

		err := dbservice.InsertResultado(ctx, &req)
		if err != nil {
			c.JSON(http.StatusConflict, ResultadoResponse{
				Status:  http.StatusConflict,
				Content: "Error al insertar el resultado",
			})
			return
		}

		c.JSON(http.StatusOK, ResultadoResponse{
			Status:  http.StatusOK,
			Content: "Resultado insertado correctamente",
		})

	})

	// Para verificar si funciona el backend
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, PongResponse{
			Status:  http.StatusOK,
			Content: "pong",
		})
	})

	fmt.Println("Server starting on :8080")
	r.Run(":8080")
}
