package ollama

import (
	"fmt"
	"time"

	//"log"

	"github.com/parakeet-nest/parakeet/completion"
	"github.com/parakeet-nest/parakeet/enums/option"
	"github.com/parakeet-nest/parakeet/llm"
)

// Constantes para la configuración del servicio
const (
	OllamaEndpoint = "http://localhost:11434"
	SystemPrompt   = "You are a virtual assistant for the Pontificia Universidad Católica de Valparaíso, also known as PUCV. Your role is to help students of the Computer Engineering (Ingeniería Civil Informática) program with their academic and administrative queries related exclusively to their program or the university. You must always communicate in Spanish, be coherent in your responses, and maintain a friendly and helpful tone. It is very important that you strictly answer questions about PUCV and the Computer Engineering program; if asked about external topics, you should politely state that you cannot provide that information. Additionally, strive for concise and direct answers to avoid overly long or repetitive text."
)

// Configuracion de los modelos
type ModelConfig struct {
	Path        string  `json:"path"`
	Temperature float64 `json:"temperature"`
}

// Service maneja la interacción con los modelos de Ollama
type Service struct {
	validModels map[string]ModelConfig
}

// NewService crea una nueva instancia del servicio de Ollama
// Inicializa el mapa de modelos válidos con sus rutas correspondientes
func NewService() *Service {
	return &Service{
		validModels: map[string]ModelConfig{
			"qwen2.5": {
				Path:        "hf.co/Ainxz/qwen2.5-pucv-gguf:latest",
				Temperature: 0.4,
			},
			"llama3.2": {
				Path:        "hf.co/Ainxz/llama3.2-pucv-gguf:latest",
				Temperature: 0.7,
			},
			"phi3.5": {
				Path:        "hf.co/Ainxz/phi3.5-pucv-gguf:latest",
				Temperature: 0.6,
			},
		},
	}
}

// OllamaData estructura para almacenar los datos de la interacción con Ollama
type OllamaData struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Message            string    `json:"message"`
	Done               bool      `json:"done"`
	TotalDuration      int64     `json:"total_duration"`
	LoadDuration       int       `json:"load_duration"`
	PromptEvalCount    int       `json:"prompt_eval_count"`
	PromptEvalDuration int       `json:"prompt_eval_duration"`
	EvalCount          int       `json:"eval_count"`
	EvalDuration       int64     `json:"eval_duration"`
}

// GenerateResponse genera una respuesta usando el modelo especificado
// Retorna la respuesta del modelo o un error si algo falla
func (s *Service) GenerateResponse(modelName, prompt string) (*OllamaData, error) {
	// Verificar si el modelo existe
	modelConfig, exists := s.validModels[modelName]
	if !exists {
		return nil, fmt.Errorf("invalid model name: %s", modelName)
	}

	// Configurar las opciones del modelo
	modelOpts := llm.SetOptions(map[string]interface{}{
		option.Temperature: modelConfig.Temperature,
	})

	// Crear la consulta para el modelo
	query := llm.Query{
		Model: modelConfig.Path,
		Messages: []llm.Message{
			{Role: "system", Content: SystemPrompt},
			{Role: "user", Content: prompt},
		},
		Stream:  false,
		Options: modelOpts,
	}

	// Obtener la respuesta del modelo
	answer, err := completion.Chat(OllamaEndpoint, query)
	if err != nil {
		return nil, fmt.Errorf("error al obtener la respuesta de ollama %w", err)
	}

	ollamaData := OllamaData{
		Model:              answer.Model,
		CreatedAt:          answer.CreatedAt,
		Message:            answer.Message.Content,
		Done:               answer.Done,
		TotalDuration:      answer.TotalDuration,
		LoadDuration:       answer.LoadDuration,
		PromptEvalCount:    answer.PromptEvalCount,
		PromptEvalDuration: answer.PromptEvalCount,
		EvalCount:          answer.EvalCount,
		EvalDuration:       answer.EvalDuration,
	}

	// log.Printf("Datos (%%+v): %+v", ollamaData)

	return &ollamaData, nil
}

// Funcion para obtener los nombres de los modelos
func (s *Service) GetModelNames() []string {
	keys := make([]string, 0, len(s.validModels))
	for k := range s.validModels {
		keys = append(keys, k)
	}
	return keys
}
