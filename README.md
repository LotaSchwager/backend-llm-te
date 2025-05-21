# Ollama Backend API

This is a Go backend service that interfaces with Ollama to provide access to specific AI models.

## Prerequisites

1. **Go Installation**
   - Install Go 1.21 or later
   - Verify installation: `go version`

2. **Ollama Installation**
   - Install Ollama from https://ollama.ai/download
   - Start Ollama service

3. **Required Models**
   Pull these models in Ollama:
   ```bash
   ollama pull hf.co/Ainxz/qwen2.5-pucv-gguf:latest
   ollama pull hf.co/Ainxz/llama3.2-pucv-gguf:latest
   ollama pull hf.co/Ainxz/phi3.5-pucv-gguf:latest
   ```

## Setup

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd <repository-name>
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Run the server:
   ```bash
   go run main.go
   ```

The server will start on port 8080.

## API Usage

### Generate Text

**Endpoint:** `POST /generate`

**Request Body:**
```json
{
    "model": "phi3.5",  // or "llama3.2" or "qwen2.5"
    "prompt": "Your prompt here"
}
```

**Response:**
```json
{
    "status": 200,
    "content": ["Generated text response"]
}
```

## Error Responses

The API will return appropriate error messages with status codes:
- 400: Invalid request format or model name
- 500: Server errors or Ollama connection issues

## Testing the API

You can test the API using curl:
```bash
curl -X POST http://localhost:8080/generate \
-H "Content-Type: application/json" \
-d '{"model": "phi3.5", "prompt": "Quien es el jefe de carrera?"}'
``` 