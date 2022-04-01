package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"github.com/gocolly/colly"
	"os"
)

//Estructura que tendrá la información guardada
type Datos struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
}

func main() {
	//Se crea un slice de tipo Datos vacío que vaya guardando todo lo que se recolecte
	Slice_hechos := make([]Datos, 0)

	//Se crea un nuevo recolector
	collector := colly.NewCollector(
		//Se permiten los siguientes dominios:
		colly.AllowedDomains("factretriever.com", "www.factretriever.com"),
	)

	//Se utiliza para ejecutar una función cada vez que se encuentre un query que coincida con el parámetro indicado
	collector.OnHTML(".factsList li", func(element *colly.HTMLElement) {
		//Se obtiene el ID de cada hecho y se transforma a int, el segundo parámetro que retorna es un error
		Id_hecho, err := strconv.Atoi(element.Attr("id"))

		//Se maneja cualquier posible error
		if err != nil {
			log.Println("No se pudo obtener el ID")
		}

		//Se obtiene la descripción
		Desc_hecho := element.Text

		//Se crea una estructura para cada hecho
		hecho := Datos{
			ID: Id_hecho,
			Description: Desc_hecho,
		}

		Slice_hechos = append(Slice_hechos, hecho)
	})

	//Función que se ejecuta cada vez que el usuario realice una request
	collector.OnRequest(func(request *colly.Request) {
		fmt.Println("Visiting", request.URL.String())
	})

	//Se indica la URL que visitará el collector
	collector.Visit("https://www.factretriever.com/rhino-facts")

	//Impresión en consola
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent(""," ")
	enc.Encode(Slice_hechos)

	//Se crea el json con el struct de datos
	writeJSON(Slice_hechos)
}

//Función para escritura del JSON que recibe un parámetro de tipo Datos
func writeJSON(data []Datos) {
	//Se transforma a tipo JSON con el metodo MarshalIndent
	file, err := json.MarshalIndent(data, "", " ")

	//Se maneja el error
	if err != nil {
		log.Println("Unable to create json file")
		return
	}

	//Se crea un archivo que almacena toda la información
	_ = ioutil.WriteFile("hechos_rhyno.json", file, 0644)
}