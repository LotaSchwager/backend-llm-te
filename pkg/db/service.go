package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

// ==================================

// Servicio de DB
type Service struct {
	User     string
	Password string
	Host     string
	Port     string
	DbName   string
}

func NewService(user *string, password *string, host *string, port *string, dbname *string) *Service {
	return &Service{
		User:     *user,
		Password: *password,
		Host:     *host,
		Port:     *port,
		DbName:   *dbname,
	}
}

// ===================================

// Tabla conversacion
type Respuesta struct {
	Modelo_id          int
	Model              string
	Message            string
	Done               bool
	TotalDuration      int64
	LoadDuration       int64
	PromptEvalCount    int64
	PromptEvalDuration int64
	EvalCount          int64
	EvalDuration       int64
}

// Tabla respuesta
type Resultado struct {
	Prompt               string `json:"prompt"`
	Respuesta_id_1       int    `json:"respuesta_id_1"`
	Respuesta_id_2       int    `json:"respuesta_id_2"`
	Respuesta_id_3       int    `json:"respuesta_id_3"`
	Respuesta_elegida_id int    `json:"respuesta_elegida_id"`
}

// ===================================

// Respuesta para InsertRespuesta
type Relacion struct {
	Model   string `json:"modelo"`
	Message string `json:"message"`
	Id      int64  `json:"id"`
}

type ResponseIds struct {
	Modelos []Relacion `json:"relacion"`
}

// Funcion de conectarse
func (s *Service) ConnectDB() (*sql.DB, error) {
	// Formatear la conexion con mysql
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&collation=utf8mb4_unicode_ci", s.User, s.Password, s.Host, s.Port, s.DbName)
	fmt.Println(dsn)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalln("Error: %w", err)
		return nil, fmt.Errorf("Error al abrir la conexion a la base de datos: %w", err)
	}

	// Intenta conectarse y esperar que no de tiempo caducado
	err = db.Ping()
	if err != nil {
		log.Fatalln("Error: %w", err)
		return nil, fmt.Errorf("Error al conectar a la base de datos: %w", err)
	}

	log.Println("Conexion exitosa con la base de datos")
	return db, nil // Devuelve la base de datos
}

func (s *Service) InsertRespuesta(db *sql.DB, respuestas *[]Respuesta) (*ResponseIds, error) {
	ids := &ResponseIds{
		Modelos: []Relacion{},
	} // Lista donde estaran los 3 ids

	// Inicializar la base de datos
	tx, err := db.Begin()
	if err != nil {
		return ids, err
	}
	defer tx.Rollback() // Si ocurre algun error, no se guardara ningun cambio

	// Query
	query := `INSERT INTO respuesta (
		modelo_id, model, message, done,
		total_duration, load_duration, prompt_eval_count,
		prompt_eval_duration, eval_count, eval_duration
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return ids, err
	}
	defer stmt.Close() // Cerrar la conexion cuando se termine de ejecutar

	// Recorrer las 3 respuestas
	for _, r := range *respuestas {
		res, err := stmt.Exec(
			r.Modelo_id, r.Model, r.Message, r.Done, r.TotalDuration,
			r.LoadDuration, r.PromptEvalCount, r.PromptEvalDuration,
			r.EvalCount, r.EvalDuration,
		)
		if err != nil {
			return ids, err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return ids, err
		}

		value := Relacion{
			Model:   r.Model,
			Message: r.Message,
			Id:      id,
		}

		ids.Modelos = append(ids.Modelos, value)
	}

	if err := tx.Commit(); err != nil {
		return ids, err
	}

	return ids, nil // Si todo va bien, se retorna el array de ids
}

func (s *Service) InsertResultado(db *sql.DB, res *Resultado) error {
	// Inicializar la base de datos
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Por si ocurre un error

	// Query
	query := `INSERT INTO resultado (
		prompt, respuesta_id_1, respuesta_id_2,
		respuesta_id_3, respuesta_elegida_id
	) VALUES (?, ?, ?, ?, ?)`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close() // Cerrar la conexion con la base de datos

	// Ejecutar la query
	response, err := stmt.Exec(
		res.Prompt, res.Respuesta_id_1, res.Respuesta_id_2,
		res.Respuesta_id_3, res.Respuesta_elegida_id,
	)
	if err != nil {
		return err
	}

	// Ver cuantas filas fueron afectadas
	rows, err := response.RowsAffected()
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	fmt.Println("Filas afectadas: %i", rows)

	return nil // Retornar nulo si todo sale bien
}
