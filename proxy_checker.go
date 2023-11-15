package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/inancgumus/screen"
)

type ResponseData struct {
	IP string `json:"ip"`
}

func isIPv4(ip string) bool {
	return net.ParseIP(ip).To4() != nil
}

var validProxies int32
var badProxies int32
var totalProxies int32

func printLogo() {
	fmt.Printf(
		`
                                               :     ..     :                  
                                        .      -.    --    .-      .           
                                         :.    .=.   ==   .=.    .:            
                                          --.   ==. .==. .==   .--             
                                   .:.     -=-  -==.:==:.==-  -=-     .:.      
                                     .--:   -==::===-==-===::==-   :--.        
                                       .-=-:.-================-.:-=-.          
                                .::..    :========================:    ..::.   
                                   .:-==---======================---==-:.      
                                       :-==========================-:          
                               .:::::::---========================---:::::::.  
                                  ...::--==========================--::...     
                                      .:-==========================-:.         
                                 .:-====--========================--====-:.    
                               ..       .-========================-.       ..  
                                      :===-:====================:-===:         
                                   .--:.  .-==-==============-==-.  .:--.      
                                 ..      :==-.:===-======-===:.-==:      ..    
                                        --:   ==- :==::==: -==   :--           
                                      .:.    -=:  :=-  -=:  :=-    .:.         
                                            .=.   :=.  .=:   .=.               
                                            -     --    --     -               
                                           .      -      -      .              
                                                  .      .                     
                                        Coded by recoo33 / savalierdev
										
		`)
}

func updateTitle() {
	for {
		screen.Clear()
		screen.MoveTopLeft()
		printLogo()
		fmt.Printf("             Valid Proxies: %d | Bad Proxies: %d | Remaining: %d\n", validProxies, badProxies, totalProxies-(validProxies+badProxies))
		time.Sleep(500 * time.Millisecond)
	}
}

func checkProxy(proxy string, wg *sync.WaitGroup, sem chan struct{}) {
	defer wg.Done()

	sem <- struct{}{}
	defer func() { <-sem }()

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(&url.URL{Scheme: "http", Host: proxy}),
		},
		Timeout: time.Second * 3,
	}

	response, err := client.Post("https://location-api.f-secure.com/v1/ip-country", "", nil)
	if err != nil || response.StatusCode != 200 {
		atomic.AddInt32(&badProxies, 1)
		return
	}
	defer response.Body.Close()

	var data ResponseData
	if err := json.NewDecoder(response.Body).Decode(&data); err != nil {
		atomic.AddInt32(&badProxies, 1)
		return
	}

	atomic.AddInt32(&validProxies, 1)

	if isIPv4(data.IP) {
		saveProxyToFile(proxy, "work_proxies_v4.txt")
	} else {
		saveProxyToFile(proxy, "work_proxies_v6.txt")
	}
}

func saveProxyToFile(proxy, fileName string) {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	file.WriteString(proxy + "\n")
}

func main() {
	var threadCount int

	fmt.Print("Thread: ")
	fmt.Scanln(&threadCount)
	printLogo()

	file, err := os.Open("proxy.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	sem := make(chan struct{}, threadCount)
	var wg sync.WaitGroup

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		proxy := scanner.Text()
		atomic.AddInt32(&totalProxies, 1)
		wg.Add(1)
		go checkProxy(proxy, &wg, sem)
	}

	// Update console
	go updateTitle()

	wg.Wait()
}
