package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Ports struct {
	Port int `json:"port"`
}

func InitFlags() (string, int, int, int, string) {
	addressFlag := flag.String("address", "scanme.nmap.org", "Address for scanning ports")
	firstPortFlag := flag.String("first_port", "1", "Scanning will start from this port")
	lastPortFlag := flag.String("last_port", "1024", "There will be scanning before this port")
	bufferChanFlag := flag.String("max_buffer", "100", "This number is responsible for the number of port scans at the moment, it makes no sense to put it above the last port")
	saveFormatFlag := flag.String("save_format", "", "Accept: json, txt")

	flag.Parse()

	address := *addressFlag
	firstPort, err := strconv.Atoi(*firstPortFlag)
	if err != nil {
		fmt.Println("Flag 'first_port' must be a number")
		os.Exit(1)
	}

	lastPort, err := strconv.Atoi(*lastPortFlag)
	if err != nil {
		fmt.Println("Flag 'last_port' must be a number")
		os.Exit(1)
	}

	bufferChan, err := strconv.Atoi(*bufferChanFlag)
	if err != nil {
		fmt.Println("Flag 'max_buffer' must be a number")
		os.Exit(1)
	}

	saveFormat := strings.ToLower(*saveFormatFlag)
	if saveFormat != "txt" && saveFormat != "json" && saveFormat != "" {
		fmt.Println("Flag 'save_format' the value must match 1 from the list:\n- json\n- txt")
		os.Exit(1)
	}

	return address, firstPort, lastPort, bufferChan, saveFormat
}

func SaveResult(result []int, format string) error {
	if _, err := os.Stat("result"); os.IsNotExist(err) {
		err := os.Mkdir("result", 0755)
		if err != nil {
			fmt.Println("Couldn't create a folder:", err)
			return err
		}
	}

	if format == "txt" {
		file, err := os.Create("result/result.txt")
		if err != nil {
			fmt.Println("Unable to create TXT file:", err)
			return err
		}

		defer file.Close()

		for _, number := range result {
			_, err := fmt.Fprintln(file, number)
			if err != nil {
				fmt.Println("Unable to write TXT file:", err)
				return err
			}
		}
	} else if format == "json" {
		file, err := os.Create("result/result.json")
		if err != nil {
			fmt.Println("Unable to create JSON file:", err)
			return err
		}

		defer file.Close()

		portSlice := make([]Ports, len(result))
		for index, port := range result {
			portSlice[index] = Ports{Port: port}
		}

		encoder := json.NewEncoder(file)
		err = encoder.Encode(portSlice)
		if err != nil {
			fmt.Println("Unable to write JSON file:", err)
		}
	}
	return nil
}

func worker(ports chan int, result chan int, adrs string) {
	for port := range ports {
		fmt.Printf("Test request for %d\n", port)
		address := fmt.Sprintf("%s:%d", adrs, port)
		conn, err := net.Dial("tcp", address)
		if err != nil {
			result <- 0
			continue
		}
		conn.Close()
		result <- port
	}
}

func main() {
	address, firstPort, lastPort, bufferChan, saveFormat := InitFlags()

	if firstPort > lastPort {
		fmt.Println("Error: The first port cannot exceed the last port.")
	} else if firstPort <= 0 || lastPort <= 0 {
		fmt.Println("Error: The port cannot be <= 0.")
	} else {
		fmt.Println("Press 'Enter' to start scan...")
		fmt.Scanf(" ")

		ports := make(chan int, bufferChan)
		results := make(chan int)
		var openPorts []int

		for i := 0; i < cap(ports); i++ {
			go worker(ports, results, address)
		}

		fmt.Println("Scanning...")
		startTime := time.Now()

		go func() {
			for numPort := firstPort; numPort <= lastPort; numPort++ {
				ports <- numPort
			}
		}()

		for i := firstPort; i <= lastPort; i++ {
			port := <-results
			if port != 0 {
				openPorts = append(openPorts, port)
			}
		}

		close(ports)
		close(results)

		elapsedTime := time.Since(startTime)
		sort.Ints(openPorts)

		fmt.Println("\nScanning time:", elapsedTime)
		if len(openPorts) != 0 {
			fmt.Printf("Found: %d element\nOpen ports:\n", len(openPorts))
			for _, port := range openPorts {
				fmt.Printf("%d open\n", port)
			}

			if saveFormat != "" {
				err := SaveResult(openPorts, saveFormat)
				if err != nil {
					os.Exit(1)
				}
			}
		} else {
			fmt.Println("No port is open")
		}
		fmt.Scanf(" ")
	}
}
