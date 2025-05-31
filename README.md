# Ollama Backend API

Este es un backend hecho en golang para el ProyectoTE.

## Prerequisitos

1. **Instalación de Go**
   - Instalar Go versión 1.21 o más actual
   - Verificacion de la instalación: `go version`

2. **Instalación de Ollama**
   - Install Ollama from https://ollama.ai/download
   - Start Ollama service

3. **Modelos requeridos**
   ```bash
   ollama pull hf.co/Ainxz/qwen2.5-pucv-gguf:latest
   ollama pull hf.co/Ainxz/llama3.2-pucv-gguf:latest
   ollama pull hf.co/Ainxz/gemma3-pucv-gguf:latest
   ```
   
## Setup

1. Clonar el repositorio:
   ```bash
   git clone https://github.com/LotaSchwager/backend-llm-te
   cd backend-llm-te
   ```

2. Instala las depedencias:
   ```bash
   go mod tidy
   ```

3. Ejecuta el servidor:
   ```bash
   go run main.go
   ```

Se necesita un .env con ciertos valores para funcionar.

## Endpoints de la API

### Generación de la respuesta de los modelos

**Endpoint:** `POST /generate`

**Request Body:**
```json
{
    "prompt": "el prompt del usuario"
}
```

**Response:**
```json
{
    "status": http status,
    "content": [{
         "model": "modelo",
         "response": "respuesta del modelo",
         "error": "error, si es que hay uno",
         "id": "id de la respuesta dentro de la tabla respuesta en la base de datos"
      }]
}
```
### Guardar la decision del usuario en la base de datos

**Endpoint:** `POST /save-result`

**Request Body:**
```json
{
    "prompt": "el prompt del usuario",
    "respuesta_id_1": id_1,
    "respuesta_id_2": id_2,
    "respuesta_id_3": id_3,
    "respuesta_elegida_id": id elegido, debe ser uno de los 3,
}
```

**Response:**
```json
{
    "status": http status,
    "content": "exitoso o no"
}
```
