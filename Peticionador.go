package Peticionador

import (
	"bytes"
	"context"
	"encoding/json"
//	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"net/url"
	
)
// Config representa la estructura del archivo JSON de configuración
type Config struct {
	Body    interface{} 	`json:"Body"`
	Header  []Header        `json:"Header"`
	Method  string          `json:"Method"`
	URL     string          `json:"URL"`
	Timeout int             `json:"Timeout,omitempty"` // Opcional: tiempo de espera en segundos
}

// Header representa la estructura de los headers en el JSON
type Header struct {
	Comentario string `json:"Comentario"`
	Nombre     string `json:"Nombre"`
	Valor      string `json:"Valor"`
}

// RequestManager gestiona las peticiones HTTP
type RequestManager struct {
	config  Config
	client  *http.Client
	ctx     context.Context
	cancel  context.CancelFunc
	headers map[string]string
	body    interface{}
	mu      sync.Mutex
}


// NewRequestManager crea una nueva instancia de RequestManager con configuración obligatoria
// El parámetro configInput puede ser:
// 1. Ruta a un archivo JSON (ej: "./config.json")
// 2. String con el contenido JSON (ej: `{"Method":"GET","URL":"https://api.com"}`)
func NewRequestManager(configInput string) (*RequestManager, error) {
    rm := &RequestManager{
        headers: make(map[string]string),
        client:  &http.Client{},
    }

    // Primero verificar si es un archivo existente
    if fileContent, err := os.ReadFile(configInput); err == nil {
        // Es un archivo válido, cargar desde el archivo
        if err := rm.loadConfig(fileContent); err != nil {
            return nil, fmt.Errorf("error cargando archivo JSON: %v", err)
        }
        return rm, nil
    }

    // Si no es archivo, asumir que es string JSON
    if err := rm.loadConfig([]byte(configInput)); err != nil {
        return nil, fmt.Errorf("error cargando string JSON: %v", err)
    }

    return rm, nil
}

// loadConfig carga la configuración desde bytes JSON (usado tanto para archivo como string)
func (rm *RequestManager) loadConfig(jsonData []byte) error {
    var config Config
    if err := json.Unmarshal(jsonData, &config); err != nil {
        return err
    }

    rm.config = config

    // Inicializar headers
    for _, h := range config.Header {
        rm.headers[h.Nombre] = h.Valor
    }

    // Inicializar body
    rm.body = config.Body

    return nil
}


// SetHeader establece o modifica un header
func (rm *RequestManager) SetHeader(nombre, valor string) {
	rm.headers[nombre] = valor
}

// SetBody establece o modifica un valor en el body
func (rm *RequestManager) SetBody(key string, value interface{}) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Verificar si es una petición SOAP
	if _, isSoap := rm.headers["SOAPAction"]; isSoap {
		return rm.setSoapBody(key, value)
	}

	// Verificar si es JSON (por Content-Type o tipo de body)
	contentType := rm.headers["Content-Type"]
	isJson := strings.Contains(contentType, "json") || (contentType == "" && (rm.body == nil || isJsonCompatible(rm.body)))

	if isJson {
		return rm.setJsonPath(key, value)
	}

	// Si no es JSON ni SOAP, manejar como URL-encoded
	return rm.setUrlEncodedBody(key, value)
}

func isJsonCompatible(body interface{}) bool {
	switch body.(type) {
	case map[string]interface{}, []interface{}:
		return true
	default:
		return false
	}
}

func (rm *RequestManager) setUrlEncodedBody(key string, value interface{}) error {
    // Convertir el body a map[string]string si es necesario
    var bodyMap map[string]string

    switch body := rm.body.(type) {
    case string:
        // Parsear el string URL-encoded
        values, err := url.ParseQuery(body)
        if err != nil {
            return fmt.Errorf("error parsing URL-encoded body: %v", err)
        }
        
        // Convertir a map[string]string
        bodyMap = make(map[string]string)
        for k, v := range values {
            if len(v) > 0 {
                bodyMap[k] = v[0]
            }
        }
        rm.body = bodyMap
        
    case map[string]interface{}:
        // Convertir map[string]interface{} a map[string]string
        bodyMap = make(map[string]string)
        for k, v := range body {
            bodyMap[k] = fmt.Sprintf("%v", v)
        }
        rm.body = bodyMap
        
    case map[string]string:
        bodyMap = body
        
    default:
        return fmt.Errorf("unsupported body type for URL encoding: %T", body)
    }

    // Asignar el nuevo valor
    bodyMap[key] = fmt.Sprintf("%v", value)
    return nil
}


func (rm *RequestManager) setJsonPath(path string, value interface{}) error {
	parts := strings.Split(path, ".")
	current := rm.body

	for i, part := range parts {
		last := i == len(parts)-1

		switch v := current.(type) {
		case map[string]interface{}:
			if last {
				v[part] = value
				return nil
			}
			// Si no es el último, seguir navegando
			if next, exists := v[part]; exists {
				current = next
			} else {
				// Crear nuevo mapa si no existe
				newMap := make(map[string]interface{})
				v[part] = newMap
				current = newMap
			}

		case []interface{}:
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(v) {
				return fmt.Errorf("índice de array inválido: %s", part)
			}
			if last {
				v[index] = value
				return nil
			}
			current = v[index]

		default:
			// Si es el primer elemento y es un número, convertirlo a array
			if i == 0 {
				if index, err := strconv.Atoi(part); err == nil {
					// Convertir el body a array si es necesario
					if rm.body == nil {
						rm.body = make([]interface{}, index+1)
						current = rm.body
					} else if arr, ok := rm.body.([]interface{}); ok {
						if index >= len(arr) {
							// Expandir array si es necesario
							newArr := make([]interface{}, index+1)
							copy(newArr, arr)
							rm.body = newArr
							current = newArr
						} else {
							current = arr
						}
					}
					
					if arr, ok := current.([]interface{}); ok {
						if last {
							arr[index] = value
							return nil
						}
						if arr[index] == nil {
							arr[index] = make(map[string]interface{})
						}
						current = arr[index]
						continue
					}
				}
			}
			return fmt.Errorf("no se puede navegar por el path: %s (tipo: %T)", path, current)
		}
	}
	return nil
}



func (rm *RequestManager) setSoapBody(path string, value interface{}) error {
	bodyStr, ok := rm.body.(string)
	if !ok {
		return errors.New("SOAP body must be a string XML")
	}

	// Buscar la posición del tag <soapenv:Body>
	bodyStart := strings.Index(bodyStr, "<soapenv:Body>")
	if bodyStart == -1 {
		bodyStart = strings.Index(bodyStr, "<soap:Body>")
		if bodyStart == -1 {
			return errors.New("no se encontró el tag soapenv:Body o soap:Body")
		}
		bodyStart += len("<soap:Body>")
	} else {
		bodyStart += len("<soapenv:Body>")
	}

	bodyEnd := strings.Index(bodyStr[bodyStart:], "</soapenv:Body>")
	if bodyEnd == -1 {
		bodyEnd = strings.Index(bodyStr[bodyStart:], "</soap:Body>")
		if bodyEnd == -1 {
			return errors.New("no se encontró el tag de cierre soapenv:Body o soap:Body")
		}
	}
	bodyEnd += bodyStart

	// Procesar path para SOAP (soporte básico para arrays)
	parts := strings.Split(path, ".")
	tagName := parts[len(parts)-1] // Usar el último componente como tag name

	// Buscar la etiqueta específica solo dentro del Body
	bodySection := bodyStr[bodyStart:bodyEnd]
	startTag := fmt.Sprintf("<%s", tagName)
	tagStart := strings.Index(bodySection, startTag)
	if tagStart == -1 {
		return fmt.Errorf("tag <%s no encontrado dentro del soapenv:Body", tagName)
	}

	// Buscar el cierre de la etiqueta de apertura
	gtIdx := strings.Index(bodySection[tagStart:], ">") + tagStart
	if gtIdx == -1 {
		return errors.New("etiqueta XML mal formada")
	}

	// Buscar la etiqueta de cierre
	endTag := fmt.Sprintf("</%s>", tagName)
	tagEnd := strings.Index(bodySection[gtIdx:], endTag) + gtIdx
	if tagEnd == -1 {
		return fmt.Errorf("etiqueta de cierre </%s> no encontrada", tagName)
	}

	// Reconstruir el XML con el nuevo valor
	newBodySection := bodySection[:gtIdx+1] + fmt.Sprintf("%v", value) + bodySection[tagEnd:]
	newBody := bodyStr[:bodyStart] + newBodySection + bodyStr[bodyEnd:]

	rm.body = newBody
	return nil
}




func (rm *RequestManager) prepareRequestBody() (io.Reader, string, error) {
	// Obtener el Content-Type de los headers (si está definido)
	contentType, exists := rm.headers["Content-Type"]
	
	// Si no hay Content-Type definido, determinar el tipo
	if !exists {
		if _, isSoap := rm.headers["SOAPAction"]; isSoap {
			contentType = "application/xml"
		} else {
			switch rm.body.(type) {
			case map[string]interface{}, []interface{}:
				contentType = "application/json"
			case string:
				// Si es string pero no es SOAP, asumir URL-encoded
				contentType = "application/x-www-form-urlencoded"
			default:
				contentType = "application/x-www-form-urlencoded"
			}
		}
	}

	// Preparar el cuerpo según el tipo de contenido
	switch {
	case strings.Contains(contentType, "xml"):
		// Para contenido XML/SOAP
		bodyStr, ok := rm.body.(string)
		if !ok {
			return nil, "", errors.New("XML body must be a string")
		}
		return strings.NewReader(bodyStr), contentType, nil

	case strings.Contains(contentType, "json"):
		// Para contenido JSON
		jsonBody, err := json.Marshal(rm.body)
		if err != nil {
			return nil, "", err
		}
		return bytes.NewReader(jsonBody), contentType, nil

    default:
        // Para URL-encoded parameters
        var bodyStr string
        
        switch body := rm.body.(type) {
        case map[string]string:
            values := url.Values{}
            for key, val := range body {
                values.Add(key, val)
            }
            bodyStr = values.Encode()
            
        case map[string]interface{}:
            values := url.Values{}
            for key, val := range body {
                values.Add(key, fmt.Sprintf("%v", val))
            }
            bodyStr = values.Encode()
            
        case string:
            // Verificar que el string sea válido URL-encoded
            if _, err := url.ParseQuery(body); err != nil {
                return nil, "", fmt.Errorf("invalid URL-encoded string: %v", err)
            }
            bodyStr = body
            
        default:
            return nil, "", fmt.Errorf("unsupported body type for URL encoding: %T", body)
        }
        
        return strings.NewReader(bodyStr), contentType, nil
    }
}



// Response realiza la petición HTTP y llama al callback con la respuesta
func (rm *RequestManager) Response() (string, int) {
    rm.mu.Lock()
    rm.ctx, rm.cancel = context.WithCancel(context.Background())
    rm.mu.Unlock()

    // Preparar el cuerpo de la petición
    bodyReader, contentType, err := rm.prepareRequestBody()
    if err != nil {
        return err.Error(), 0
    }

    // Configurar headers
    if _, exists := rm.headers["Content-Type"]; !exists {
        rm.headers["Content-Type"] = contentType
    }

    req, err := http.NewRequestWithContext(
        rm.ctx,
        strings.ToUpper(rm.config.Method),
        rm.config.URL,
        bodyReader,
    )
    if err != nil {
        return err.Error(), 0
    }

    // Añadir headers
    for nombre, valor := range rm.headers {
        req.Header.Add(nombre, valor)
    }

    resp, err := rm.client.Do(req)
    if err != nil {
        if rm.ctx.Err() == context.Canceled {
            return "request canceled", 0
        }
        return err.Error(), 0
    }
    defer resp.Body.Close()

    responseBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return err.Error(), 0
    }

    return string(responseBody), resp.StatusCode
}


// Cancel cancela la petición HTTP en curso
func (rm *RequestManager) Cancel() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	if rm.cancel != nil {
		rm.cancel()
	}
}

