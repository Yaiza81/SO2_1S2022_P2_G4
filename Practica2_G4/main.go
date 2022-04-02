package main 

import (
	"os"
	"fmt"
	"github.com/gocolly/colly"
	"strings"
	"time"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"sync"
	"io/ioutil"
	"encoding/json"
	"log"
	"strconv"
	"crypto/sha256"
	"encoding/hex"
)

// Variables
var cantMonos_ =0
var tamCola_ =0
var numNr_ =0
var urlInicial_ =""
var archivo_ =""

type responseMsg struct {
	indice int
	url string
	estado string
	palabras int
	enlaces int
	cola int
	sha string
}

type trabajito struct {
	Url string
	Busqueda string
	Referencias int
}

type cache struct{
	mu sync.RWMutex
	lista map[string]string
}

//Estructura que tendrá la información a escribir en el JSON
type Datos struct {
	Origen string 		`json:"origen"`
	Cont_palabras int	`json:"cont_palabras"`
	Cont_enlaces int	`json:"cont_enlaces"`
	Sha string			`json:"sha"`
	Url string			`json:"url"`
	Mono string			`json:"mono"`
}

//Se crea un slice de tipo Datos vacío que vaya guardando todo lo que se recolecte
var Slice_hechos = make([]Datos, 0)

//Cola de espera
var colita = &cache{lista: make(map[string]string)}

//Funciones para el manejo de la cola (se utilizó Mutex)
//len(colita.lista)<5
func agregar(k string, v string){
	if(len(colita.lista)<tamCola_){
		colita.mu.Lock()
		colita.lista[k]=v
		colita.mu.Unlock()
	}
}

func quitar(k string){
	colita.mu.Lock()
	delete(colita.lista,k)
	colita.mu.Unlock()
}

func leer() string{
	colita.mu.Lock()
	str:=""
	for k,v:= range colita.lista{
		str+=fmt.Sprintf("%s -> %s \n",k,v)
	}
	colita.mu.Unlock()
	return str
}

//Función para la interfaz gráfica
func listenForActivity(sub chan responseMsg) tea.Cmd {
	return func() tea.Msg {		

		//Definición de los canales
		jobs := make(chan trabajito, 100)
		results := make(chan trabajito, 100)

		//Definición de los 3 monos
	//	go mono(jobs,results, sub,0)
	//	go mono(jobs,results, sub,1)
	//	go mono(jobs,results, sub,2)

		for k:=0;k<cantMonos_;k++{
			go mono(jobs,results, sub,k)
		}
		//Se realiza la primera búsqueda y se define el Nr = 3
		jobs <- trabajito {"https://es.wikipedia.org/wiki/Chuck_Norris","GRUPO4_SOPES", numNr_}

		for r:= range results{
			x:= r
			//fmt.Println(x.Busqueda)

			//Timer para que se muestre en pantalla
			time.Sleep(time.Duration(1) * time.Second)
			quitar(x.Busqueda)
			jobs <- x
		}
		return nil
	}
}

func waitForActivity(sub chan responseMsg) tea.Cmd{
	return func () tea.Msg{
		return responseMsg(<-sub)
	}
}

//Definición de los monos (worker)
//se envía un canal de trabajo y uno de resultados
func mono(jobs <- chan trabajito, results chan <- trabajito, sub chan responseMsg, indice int){
	
	for j := range jobs {
		Url := j.Url
		Nr := j.Referencias

		conteo_palabras := 0
		var enlaces []string
		var nombres_enlaces[]string
		var sha string

		sub <- responseMsg {indice, Url, "trabajanding", 0,0,-1,sha}

		//Se crea el recolector
		collector := colly.NewCollector(colly.Async(false))

		//Función que se realiza cada vez que se realiza una petición
		collector.OnRequest(func(collector *colly.Request) { })

		//OnHTML ejecuta algo cada vez que se encuentre un query que coincida con el parámetro indicado
		collector.OnHTML("div#mw-content-text p", func(element *colly.HTMLElement) {
			conteo_palabras += len(strings.Split(element.Text," "))
			sub <- responseMsg {indice, Url, "Trabajanding...", conteo_palabras,len(enlaces),-1,sha}
			time.Sleep(500)
		})
		collector.OnHTML("div#mw-content-text p a", func(element *colly.HTMLElement) {
			enlaces= append(enlaces, element.Request.AbsoluteURL(element.Attr("href")))
			nombres_enlaces=append(enlaces, element.Text)
			sub <- responseMsg {indice, Url, "Trabajanding", conteo_palabras,len(enlaces),-1,sha}
		})

		collector.OnHTML("div#mw-content-text",func(e *colly.HTMLElement){
			result:= e.ChildText("p")
			sha=getSha(result)

			//fmt.Println(e.ChildText("p"), "sha:" , sha)
			//fmt.Println("sha: ",sha)
			sub <- responseMsg {indice, Url, "Trabajanding", conteo_palabras,len(enlaces),-1,sha}
		})

		//OnScraped se ejecuta al final luego de los OnHTML arreglo[4] = 4 arreglo[3]
		collector.OnScraped(func (element *colly.Response) {
			sub <- responseMsg {indice, Url, "Descansanding", conteo_palabras, len(enlaces),-1,sha}
			for i:=0; i< Nr; i++{
				if (len(enlaces)>1&&Nr<len(enlaces)){
					//&& len(enlaces)<len(enlaces[i])
					//fmt.Println("AQUIIMPRESION")
					aux:=enlaces[i] //ARREGLO[4] LEN=4 ARREGLO[3]
					//fmt.Println("Auxiliar: ", aux)
					nombre:=nombres_enlaces[i]
					//fmt.Println("Nombres enlaces: ", nombre)
					if (len(results)<10){
						agregar(nombre,aux)
						//fmt.Println("Despues agregar...")
						results <- trabajito {aux, nombre, Nr-1}
					}
				}
			}

			//Se crea una estructura para los datos del mono
			data := Datos {
				Origen: "a",
				Cont_palabras: conteo_palabras,
				Cont_enlaces: len(enlaces),
				Sha: sha,
				Url: Url,
				Mono: "c",
			}
			
			//Antes de terminar agregamos la info al slice
			Slice_hechos = append(Slice_hechos, data)

			//Tiempo proporcional de espera
			time.Sleep(time.Duration(conteo_palabras/500)*time.Second)
		})

		//URL que visita el recolector
		collector.Visit(Url)
	}
}

type model struct{
	sub chan responseMsg

	monos	[]string
	urls	[]string
	palabras	[]int 
	enlaces		[]int 
	estados		[]string 
	sha []string
	
	links string
	rola int
	spinner		spinner.Model
	quitting	bool
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		spinner.Tick,
		listenForActivity(m.sub),
		waitForActivity(m.sub),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd){
	switch msg.(type){
	case tea.KeyMsg:
		m.quitting=true
		return m,tea.Quit

	case responseMsg:
		//ftm.Println(msg)
		respuesta:=msg.(responseMsg)

		if (respuesta.cola == -1){
			m.urls[respuesta.indice]=respuesta.url
			m.palabras[respuesta.indice]=respuesta.palabras
			m.enlaces[respuesta.indice]=respuesta.enlaces
			m.estados[respuesta.indice]=respuesta.estado
			m.sha[respuesta.indice]=respuesta.sha
			m.links=leer()
		}
		return m, waitForActivity(m.sub) //wait for next event
		
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner,cmd=m.spinner.Update(msg)
		return m,cmd
		
	default:
		return m,nil
	}
}

//Función para mostrar los resultados
func (m model) View() string {
	var style =lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color("#7D56F4")).
		Height(1)

	s:=fmt.Sprintf(style.Render("----- Silencio monos trabajando -----"))
	s+= fmt.Sprintf("\n\n")
	for j:=0;j<cantMonos_;j++{
		//s+= fmt.Sprintf("%s %s url: %s \n palabras contadas: %d enlaces: %d \n\n", m.monos[i], m.estados[i], m.urls[i], m.palabras[i], m.enlaces[i])
		m.monos[j]="mono_"+strconv.Itoa(j);
	}

	for i:=0;i<cantMonos_;i++{
		s+= fmt.Sprintf("%s %s url: %s \n palabras contadas: %d enlaces: %d  sha: %s \n\n", m.monos[i], m.estados[i], m.urls[i], m.palabras[i], m.enlaces[i], m.sha[i])
	}

	s += fmt.Sprintf(style.Render("----- Cola de trabajo -----"))
	s += fmt.Sprintf("\n\n %s",m.links)
	s += fmt.Sprintf("\n\nPresione cualquier tecla para salir")
	
	//Aquí finaliza la ejecución del programa
	if m.quitting{
		//Se crea el json con el struct de datos
		writeJSON(Slice_hechos)

		s += "\n"		
	}

	return s
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
	
	_ = ioutil.WriteFile(archivo_, file, 0644)
}

func ejecucion() {
	p := tea.NewProgram(model {
		sub: make(chan responseMsg),
		//monos:	[]string{"Espino","Turk","Juanito"},
		monos:	make([]string,cantMonos_),
		//urls:	[]string{"","",""},
		urls:	make([]string,cantMonos_),
		//estados:	[]string{"Esperando","Esperando","Esperando"},
		estados:	make([]string,cantMonos_),
		//palabras:	[]int{0,0,0},
		palabras:	make([]int,cantMonos_),
		enlaces:	make([]int, cantMonos_),
		sha: make([]string,cantMonos_),

		//enlaces:	[]int{0,0,0},

		spinner: spinner.New(),
		//cola 0,
	})
	
	if p.Start() != nil {
		fmt.Println("Error al iniciar el programa")
		os.Exit(1)
	}

	
}

//función para obtener Sha256 
func getSha(cadena string) string {
	h := sha256.New()
	h.Write([]byte(cadena))
	sha_hash := hex.EncodeToString(h.Sum(nil))
	return sha_hash
}

func main(){

	var opcion1 int
    
	for ok := true; ok; ok = !(opcion1>2) {
		fmt.Println("SELECCIONE SU OPCION")
		fmt.Println("1. Ingreso de datos")
		fmt.Println("2. Ejecutar")
		fmt.Println("3. Salir")
		fmt.Println("INGRESE OPCION:")
		fmt.Scan(&opcion1)
		fmt.Println("")
	
		switch opcion1 {
			case 1:
				{
					fmt.Println("1. Cantidad de mono buscadores")
					fmt.Scan(&cantMonos_)
					fmt.Println("")

					fmt.Println("2. Tamaño de la cola")
					fmt.Scan(&tamCola_)
					fmt.Println("")

					fmt.Println("3. Numero Nr")
					fmt.Scan(&numNr_)
					fmt.Println("")

					fmt.Println("4. URL inicial")
					fmt.Scan(&urlInicial_)
					fmt.Println("")

					fmt.Println("3. Nombre del archivo")
					fmt.Scan(&archivo_)
					archivo_ += ".json"
					fmt.Println("")
				}
				
			case 2:
				{
					fmt.Println("Ejecutando...")
					ejecucion()					
				}
			
			default:
				fmt.Println("Salir")
				os.Exit(3)
		}
	}
}