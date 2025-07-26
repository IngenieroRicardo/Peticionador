package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	// Imprimir método y ruta
	fmt.Println("Método:", r.Method)
	fmt.Println("Ruta:", r.URL.Path)

	// Imprimir headers
	fmt.Println("Headers:")
	for name, values := range r.Header {
		for _, value := range values {
			fmt.Printf("  %s: %s\n", name, value)
		}
	}

	// Leer y mostrar body
	fmt.Println("Body:")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Error leyendo el body:", err)
	} else {
		fmt.Println(string(body))
	}

	// Responder al cliente
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Datos recibidos. Ver consola."))
}

func main() {
	http.HandleFunc("/", handler)

	fmt.Println("Servidor escuchando en http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Error al iniciar el servidor:", err)
	}
}
