### Usar JSON en ves del archivo, para mayor portabilidad:
```golang
package main

import (
	"fmt"
	"github.com/IngenieroRicardo/Peticionador"
)

func main() {

	manager, _ := Peticionador.NewRequestManager(`{
	"Method":"POST",
	"URL":"https://httpbin.org/delay/5",
	"Header":[
		{
			"Comentario":"",
			"Nombre":"User-Agent",
			"Valor":"PeticionadorJSON/1.0"
		},{
			"Comentario":"",
			"Nombre":"Content-Type",
			"Valor":"application/json"
		}
	],
	"Body": {
		"modificaCampo":"Hola mundo",
		"id":101,
		"erreglo": [ 1,2,3 ]
	}
}`	)


	manager.SetHeader("Authorization", "Bearer token123")

	manager.SetBody("modificaCampo", "Golang Peticionador")
	manager.SetBody("nuevoCampo", "valor")
	manager.SetBody("erreglo.0", 100)
	
	body, status := manager.Response()

	fmt.Println("Estado: ",status,"Cuerpo:",body)
}
```

### Hacer una peticion cancelable, modo 1:
```golang
package main

import (
	"fmt"
	"time"
	"github.com/IngenieroRicardo/Peticionador"
)

func main() {

  	manager, err := Peticionador.NewRequestManager("./config.json")
 	if err != nil {
 		panic(err)
 	}
 	go manager.Response(repuesta)

 	time.Sleep(1* time.Second)
 	manager.Cancel()
 	time.Sleep(5* time.Second)
}

func repuesta(body string, status int) {
    if status != 0 {
        fmt.Println("Estado:", status, "Cuerpo:", body)
    }
}

```



### Hacer una peticion cancelable, modo 2:
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
			cancelar() //Cancelar la peticion despues de 3 segundos
	}
}
```

