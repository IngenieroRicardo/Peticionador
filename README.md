# Peticionador
Libreria basada en los archivos de https://gitlab.com/RicardoValladares/json


```golang
package main

import (
	"time"
	"fmt"
	"github.com/IngenieroRicardo/Peticionador"
)

var (
	manager   *Peticionador.RequestManager
	err       error
)

type Result struct {
	response string
	status   int
}


func init() {
	manager, err = Peticionador.NewRequestManager("./config.json") //En ves de un archivo puede ser un JSON
	if err != nil {
		panic(err)
	}
}

func cancelar() {
	manager.Cancel()
	fmt.Println("Petición cancelada")
}


func main() {
	manager.SetHeader("Authorization", "Bearer token123")

	manager.SetBody("modificaCampo", "Golang Peticionador")
	manager.SetBody("nuevoCampo", "valor")
	manager.SetBody("erreglo.0", 100)
	
	result := make(chan Result)

	go func() {
		body, status := manager.Response()
		result <- Result{response: body, status: status}
	}()

	select {
		case resultado := <-result: 
			if resultado.status == 0 {
				fmt.Println("Peticion fallida..")
			} else {
				fmt.Println(resultado.response)
			}
		case <-time.After(3 * time.Second): 
			cancelar()
	}
}
```
