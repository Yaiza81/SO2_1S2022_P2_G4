package main 

import (
	"os"
	"fmt"
	"github.com/gocolly/gocolly"
	"strings"
	"time"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"sync"
)

type responseMsg struct{
	indice int
	url string
	estado string
	palabras int
	enlaces int
	cola int
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

var colita =&cache{lista: make(map[string]string)}

func agregar(k string, v string){
	if(len(colita.lista)<5){
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
		str+=fmt.Sprintf("%s → %s \n",k,v)
	}
	colita.mu.Unlock()
	return str
}

func listenForActivity(sub chan responseMsg) tea.Cmd(){
	return func() tea.Msg{
		jobs:=make(chan trabajito,100)
		results:=make(chan trabajito,100)

		go mono(jobs,results, sub,0)
		go mono(jobs,results, sub,1)
		go mono(jobs,results, sub,2)

		jobs ← trabajito {"https://es.wikipedia.org/wiki/Chuck_Norris","Chuck",3}

		for r:= range results{
			x:= r
			//fmt.Println(x.Busqueda)

			time.Sleep(time.Duration(1) * time.Second)
			quitar(x.Busqueda)
			jobs ← x
		}
		return nil
	}
}

func waitForActivity(sub chan responseMsg) tea.Cmd{
	return func () tea.Msg{
		return responseMsg(←sub)
	}
}

func mono(jobs ←chan trabajito, results chan← trabajito, sub chan responseMsg, indice int){
	
	for j:= range jobs{
		Url:= j.Url
		Nr.j.Referencias

		conteo_palabras:=0
		var enlaces []string
		var nombres_enlaces[]string

		sub ← responseMsg {indice, Url, "trabajanading", 0,0,-1}

		c:= colly.NewCollector(colly.Async(false))
		c.OnRequest(func(c *colly.Request) { })

		c.OnHTML("div#mw-content-text p", func(e *colly.HTMLElement) {
			conteo_palabras+=len(strings.Splite(e.Text," "))
			sub ← responseMsg {indice, Url, "trabajanading", conteo_palabras,len(enlaces),-1}
			time.Sleep(500)
		})

		c.OnHTML("div#mw-content-text p a", func(e *colly.HTMLElement) {
			enlaces= append(enlaces, e.Request.AbsoluteURL(e.Attr("href")))
			nombres_enlaces=append(enlaces, e.Text)
			sub ← responseMsg {indice, Url, "trabajanading", conteo_palabras,len(enlaces),-1}
		})

		c.OnScraped(func (e *colly.Response){
			sub ← responseMsg {indice, Url, "trabajanading", conteo_palabras,len(enlaces),-1}
			for i:=0; i< Nr; i++{
				if (len(enlaces)>1){
					aux:=enlaces[i]
					nombre:=nombres_enlaces[i]
					if (len(results)<10){
						agregar(nombre,aux)
						results ← trabajito {aux, nombre, Nr-1}
					}
				}
			}
			time.Sleep(time.Duration(conteo_palabras/500)*time.Second)
		})
		c.Visit(Url)
	}
}

type model struct{
	sub chan responseMsg

	monos	[]string
	urls	[]string
	palabras	[]string 
	enlaces		[]string 
	estados		[]string 
	
	links string
	rola int
	spinner		spinner.Model
	quitting	bool
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd){
	switch msg.(type){
	case tea.KeyMsg:
		m.quitting=true
		return m,tea.Quit

	case responseMsg:
		//ftm.Println(msg)
		respuesta:=msg.(responseMsg)

		if (respuesta.cola-1){
			m.urls[respuesta.indice]=respuesta.url
			m.palabras[respuesta.indice]=respuesta.palabras
			m.enlaces[respuesta.indice]=respuesta.enlaces
			m.estados[respuesta.indice]=respuesta.estado
			m.links=leer()
		}
		return m, waitForActivity(m.sub) //wait for next event
		
	case spinner.TickMsg:
		var cmd.tea.Cmd
		m.spinner,cmd=m..spinner.Update(msg)
		return m,cmd
		
	default:
		return m,nil
	}
}

func (m model) View() string{
	var style =lipgloss.NewStyle().
		Bold(true)
		Background(lipgloss.Color("#7D56F4")).
		Height(1)

	s:=fmt.Sprintf(style.Render("----- Silencio monos trabajando -----"))
	s+= fmt.Sprintf("\n\n")

	for i:=0;i<3;i++{
		s+= fmt.Sprintf("%s %s url: %s \n palabras contadas: %d enlaces: %d \n\n", m.monos[i], m.estados[i], m.urls[i], m.palabras[i], m.enlaces[i])
	}

	s += fmt.Sprintf(style.Render("----- Cola de trabajo -----"))
	s += fmt.Sprintf("\n\n %s",m.links)
	s += fmt.Sprintf("\n\nPresione cualquier tecla para salir")
	
	if m.quitting{
		s += "\n"
	}

	return s
}

func main(){
	p:=tea.NewProgram(model{
		sub: make(chan responseMsg),
		monos:	[]string{"Espino","Turk","Juanito"},
		urls:	[]string{"","",""},
		estados:	[]string{"Esperando","Esperando","Esperando"},
		palabras:	[]int{0,0,0},
		enlaces:	[]int{0,0,0},

		spinner: spinner.New(),
		//cola 0,
	})

	if p.Start() ≠ nil {
		fmt.Println("could not start the program")
		os.Exit(1)
	}
}