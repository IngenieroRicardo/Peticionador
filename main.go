package main

import (
	"time"
	"fmt"
	"Peticionador/Peticionador" // Asumiendo que el paquete está en esta ruta
)

func main() {

		rm := Peticionador.NewRequestManager()

		// Cargar configuración desde JSON
		err := rm.LoadJSON("config.json")
		if err != nil {
			fmt.Println("Error al cargar JSON:", err)
			return
		}

	
		// Configurar endpoint con delay suficiente para cancelar
		// Modificar un header si es necesario
		rm.SetHeader("Authorization", "Bearer token123")
	
		// Modificar el body si es necesario
		rm.SetBody("data", "Golang Peticionador")
		rm.SetBody("nuevoCampo", "valor")
		rm.SetBody("erreglado.0", 100)
		
		// Variable para verificar si la petición terminó
		var completed bool
	
		go func() {
			err := rm.Response(func(body string, status int) {
				completed = true
				fmt.Printf("Status: %d\n", status)
				fmt.Printf("Response: %s\n", body)
			})
			
			if err != nil {
				fmt.Println("Error:", err)
				completed = true
			}
		}()
	
		// Esperar 1 segundo y luego cancelar
		time.Sleep(1 * time.Second)
		rm.Cancel()
		fmt.Println("Petición cancelada")
	
		// Esperar un tiempo suficiente para ver el resultado
		time.Sleep(4 * time.Second)
		
		if !completed {
			fmt.Println("La petición fue cancelada antes de completarse")
		}

		time.Sleep(8 * time.Second)
}
