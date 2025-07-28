package main

import (
	"time"
	"fmt"
	"Peticionador/Peticionador" // Asumiendo que el paquete está en esta ruta
)

var (
	manager   *Peticionador.RequestManager
	err       error
)


func main() {
	manager.SetHeader("Authorization", "Bearer token123")

	manager.SetBody("data", "Golang Peticionador")
	manager.SetBody("nuevoCampo", "valor")
	manager.SetBody("erreglado.0", 100)
	
	go func() {
		err := manager.Response(respuesta)
		
		if err != nil {
			fmt.Println("Error:", err)
		}
	}()
	
	// Esperar 1 segundo y luego cancelar
	//cancelar()


	// Esperar 8 segundo y cerrar
	time.Sleep(8 * time.Second)	
}


func init() {
	manager, err = Peticionador.NewRequestManager("./config.json")
	if err != nil {
		panic(err)
	}
}

func respuesta(body string, status int) {
	fmt.Printf("Status: %d\n", status)
	fmt.Printf("Response: %s\n", body)
}

func cancelar() {
	time.Sleep(1 * time.Second)
	manager.Cancel()
	fmt.Println("Petición cancelada")
}