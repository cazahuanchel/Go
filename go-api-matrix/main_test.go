package main

import (
	"errors"
	"reflect" // Necesario para comparar slices de slices (matrices)
	"testing"

	"github.com/gofiber/fiber/v2" // Necesario para verificar errores específicos de Fiber
)

// Definición de una estructura para los casos de prueba de rotación
type rotateTest struct {
	name          string
	input         Matrix
	expected      Matrix
	expectedError bool
	errorMessage  string // Mensaje de error esperado para verificación
}

func TestRotateMatrix(t *testing.T) {
	// Definimos varios casos de prueba para la función rotateMatrix
	tests := []rotateTest{
		{
			name: "Matriz 2x3 básica (Rotación 90 grados Correcta)",
			input: Matrix{
				{1, 2, 3},
				{4, 5, 6},
			},
			// Esperado: La matriz 2x3 se convierte en una 3x2, rotada 90° horario
			expected: Matrix{
				{4, 1},
				{5, 2},
				{6, 3},
			},
			expectedError: false,
		},
		{
			name: "Matriz Cuadrada 3x3",
			input: Matrix{
				{1, 2, 3},
				{4, 5, 6},
				{7, 8, 9},
			},
			expected: Matrix{
				{7, 4, 1},
				{8, 5, 2},
				{9, 6, 3},
			},
			expectedError: false,
		},
		{
			name: "Matriz con Números Negativos y Cero",
			input: Matrix{
				{-1, 0},
				{5, -50},
			},
			expected: Matrix{
				{5, -1},
				{-50, 0},
			},
			expectedError: false,
		},
		{
			name:          "Matriz Vacía (Error Esperado)",
			input:         Matrix{},
			expected:      nil,
			expectedError: true,
			errorMessage:  "La matriz no puede estar vacía",
		},
		{
			name: "Matriz No Rectangular (Error Esperado)",
			input: Matrix{
				{1, 2, 3},
				{4, 5}, // Fila corta, viola el requisito rectangular
			},
			expected:      nil,
			expectedError: true,
			errorMessage:  "La matriz debe ser rectangular. La fila 1 tiene 2 columnas, pero se esperaban 3.",
		},
	}

	// Ejecutamos cada caso de prueba
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := rotateMatrix(tt.input)

			// 1. Manejo de Errores
			if (err != nil) != tt.expectedError {
				t.Errorf("rotateMatrix() error = %v, expectedError %v", err, tt.expectedError)
				return
			}

			if tt.expectedError {
				// Si esperábamos un error, verificamos que sea el mensaje correcto de Fiber
				var ferr *fiber.Error
				if errors.As(err, &ferr) {
					if ferr.Message != tt.errorMessage {
						t.Errorf("Error inesperado. Se esperaba el mensaje: '%s', se obtuvo: '%s'", tt.errorMessage, ferr.Message)
					}
				} else {
					t.Errorf("Error de tipo incorrecto. Se esperaba fiber.Error, se obtuvo %T", err)
				}
				return
			}

			// 2. Comparación de Resultados (Solo si no se esperaba un error)
			// reflect.DeepEqual es crucial para comparar matrices (slices de slices)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("rotateMatrix() resultado = \n%v, \nquería \n%v", result, tt.expected)
			}
		})
	}
}
