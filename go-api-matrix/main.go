package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Define una estructura para la matriz de entrada.
type Matrix [][]int

// Definimos una estructura para la respuesta que esperamos de la API de Node.js
type StatsResponse struct {
	Message string `json:"message"`
	Stats   struct {
		ValorMaximo    int     `json:"valorMaximo"`
		ValorMinimo    int     `json:"valorMinimo"`
		Promedio       float64 `json:"promedio"`
		SumaTotal      int     `json:"sumaTotal"`
		MatrizDiagonal bool    `json:"matrizDiagonal"`
	} `json:"stats"`
}

// URL del endpoint de la API de Node.js (asegúrate de que Node.js esté corriendo en este puerto)
// const nodeApiUrl = "http://localhost:3001/stats/calculateMatrixStats" // Para local
const nodeApiUrl = "http://node-api:3001/calculate-stats" // Para docker

func main() {
	app := fiber.New()
	app.Post("/rotate-and-send", handleRotateAndSend)
	// Escuchar en el puerto 3000
	log.Fatal(app.Listen(":3000"))
}

// Clave secreta para JWT
const jwtSecret = "OU2vp5HZsx0Mw0Xy9HLJBn7nTmVM9MKatJUvpS9IKPa"

// Función para generar un JWT
func generateJWT(secret string) (string, error) {
	claims := jwt.MapClaims{
		"authorized": true,
		"user_id":    "go_api_service",
		"exp":        time.Now().Add(time.Hour * 1).Unix(), // Expira en 1 hora
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return t, nil
}

// Handler para la solicitud
func handleRotateAndSend(c *fiber.Ctx) error {
	// 1. Recibir la matriz del cuerpo de la solicitud
	var originalMatrix Matrix
	if err := c.BodyParser(&originalMatrix); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Solicitud inválida. Asegúrate de enviar un array de arrays de números.",
		})
	}

	// 2. Realizar la Rotación de la Matriz (incluye validación de rectangularidad)
	rotatedMatrix, err := rotateMatrix(originalMatrix)
	if err != nil {
		// handleRotateAndSend devuelve un error, lo mapeamos a una respuesta Fiber
		var ferr *fiber.Error
		if errors.As(err, &ferr) {
			return c.Status(ferr.Code).JSON(fiber.Map{"error": ferr.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error interno al rotar la matriz"})
	}

	// 3. Generar JWT
	token, err := generateJWT(jwtSecret)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Fallo al generar el token JWT.",
		})
	}

	// 4. Enviar la Matriz Rotada a la API de Node.js
	client := resty.New()
	var result StatsResponse

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+token).
		SetBody(rotatedMatrix).
		SetResult(&result).
		Post(nodeApiUrl)

	// Manejo de errores de red o conexión
	if err != nil {
		log.Printf("Error al conectar con la API de Node.js (%s): %v", nodeApiUrl, err)
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "No se pudo comunicar con la API de Node.js. Asegúrate de que esté corriendo en el puerto 3001.",
		})
	}

	// Manejo de códigos de estado HTTP de Node.js
	if resp.StatusCode() != http.StatusOK {
		return c.Status(resp.StatusCode()).JSON(fiber.Map{
			"error":   "La API de Node.js devolvió un error.",
			"details": string(resp.Body()),
		})
	}

	// 5. Devolver la respuesta de la API de Node.js
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":        "Proceso completado: Rotación en Go y Estadísticas en Node.js.",
		"original_rows":  len(originalMatrix),
		"original_cols":  len(originalMatrix[0]),
		"matrix_rotated": rotatedMatrix,
		"statistics":     result.Stats,
	})
}

// Función central para la rotación de la matriz 90 grados en sentido horario
func rotateMatrix(matrix Matrix) (Matrix, error) {
	if len(matrix) == 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "La matriz no puede estar vacía")
	}

	// M: número de filas (alto)
	M := len(matrix)
	// N: número de columnas (ancho). Se asume que la primera fila no está vacía.
	N := len(matrix[0])

	// 🚩 VERIFICACIÓN CRUCIAL DE RECTANGULARIDAD
	// Esto previene el "index out of range"
	for i, row := range matrix {
		if len(row) != N {
			// Añadimos contexto para saber qué fila falló
			errMsg := fmt.Sprintf("La matriz debe ser rectangular. La fila %d tiene %d columnas, pero se esperaban %d.", i, len(row), N)
			return nil, fiber.NewError(fiber.StatusBadRequest, errMsg)
		}
	}

	// La matriz rotada tendrá N filas y M columnas
	rotatedMatrix := make(Matrix, N)
	for i := range rotatedMatrix {
		rotatedMatrix[i] = make([]int, M)
	}

	// Mapeo de coordenadas: (i, j) -> (j, M - 1 - i)
	for i := 0; i < M; i++ {
		for j := 0; j < N; j++ {
			// Nueva fila (i_new) es la columna original j
			iNew := j
			// Nueva columna (j_new) es (M - 1 - i)
			jNew := M - 1 - i

			rotatedMatrix[iNew][jNew] = matrix[i][j]
		}
	}

	return rotatedMatrix, nil
}
