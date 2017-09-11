package main

import (
	"fmt"
	"strings"
	"flag"
	"log"
	"io/ioutil"
	"encoding/json"
	"net/http"
	"os"
)

var configFile = flag.String("c", "", "location of configuration json file")

const userAgent = "noip-update-madeye.com/%d %s"
const version = 2

type Config struct {
	Urls []string
	Hostnames []string
	NoipUsername string
	NoipPassword string
	NoipUrl string
	NoipAdminEmail string
}

func checkError(e error){
	if e != nil {
		log.Fatalf("Error: %s\n", e)
	}
}

func initLogging() *os.File {
	f, err := os.OpenFile("update-noip.log", os.O_APPEND | os.O_CREATE | os.O_RDWR, 0666)
	checkError(err)

	// assign it to the standard logger
	log.SetOutput(f)

	return f
}

func loadConfig(filename string) (Config, error){
	var config Config
	configData, err := ioutil.ReadFile(filename)

	if err != nil {
		return config, err
	}

	if err = json.Unmarshal(configData, &config); err != nil {
		return config, err
	}

	return config, nil
}

func getUserAgent(config Config) string {
	return fmt.Sprintf(userAgent, version, config.NoipAdminEmail)
}

func getCurrentIp(config Config) string {
	userAgent := getUserAgent(config)
	client := &http.Client{}

	for _, url := range config.Urls {
		log.Printf("Retrieving IP address from: %s", url)
		req, err := http.NewRequest("GET", url, nil)
		checkError(err)
		req.Header.Add("User-Agent", userAgent)

		resp, err := client.Do(req)
		if resp != nil {
			//log.Printf("Status: %d", resp.StatusCode)
			if resp.StatusCode == 200 {
				body, err := ioutil.ReadAll(resp.Body)
				resp.Body.Close()
				checkError(err)
				//log.Printf("Body: %s", string(body))
				ipAddress := strings.TrimSpace(string(body))
				log.Printf("IP address: %s", ipAddress)
				return ipAddress
			}
		} else {
			log.Printf("Nil response!")
		}
	}

	return ""
}

func updateNoip(config Config, ip string, hostname string) string {
	url := fmt.Sprintf(config.NoipUrl, ip, hostname)
	userAgent := getUserAgent(config)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	checkError(err)
	req.Header.Add("User-Agent", userAgent)
	req.SetBasicAuth(config.NoipUsername, config.NoipPassword)

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if resp != nil {
		//log.Printf("Status: %d", resp.StatusCode)
		if resp.StatusCode == 200 {
			body, err := ioutil.ReadAll(resp.Body)
			checkError(err)
			//log.Printf("Body: %s", string(body))
			return strings.TrimSpace(string(body))
		} else {
			log.Printf("Failed to update: %s", resp.Status)
		}
	} else {
		log.Printf("Nil response!")
	}

	return ""
}

func main(){
	flag.Parse()
	logFile := initLogging()
	defer logFile.Close()
	config, err := loadConfig(*configFile)
	checkError(err)

	ip := getCurrentIp(config)

	if len(ip) > 0 {
		for _, host := range config.Hostnames {
			response := updateNoip(config, ip, host)
			log.Printf("%s => %s", host, response)
		}
	} else {
		log.Fatal("Cannot determine IP address")
	}
}

