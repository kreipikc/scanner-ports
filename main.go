package main

import (
	"bufio"
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

func InitFlags() (string, int, int, int, string, string) {
	addressFlag := flag.String("address", "", "Address for scanning ports")
	firstPortFlag := flag.String("first_port", "1", "Scanning will start from this port")
	lastPortFlag := flag.String("last_port", "1024", "There will be scanning before this port")
	bufferChanFlag := flag.String("max_buffer", "100", "This number is responsible for the number of port scans at the moment, it makes no sense to put it above the last port")
	saveFormatFlag := flag.String("save_format", "", "Accept: json, txt")
	documentFlag := flag.String("document", "", "Path for document with addresses for scanning")

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

	document := *documentFlag
	if _, err := os.Stat(document); os.IsNotExist(err) && document != "" {
		fmt.Println("No document found")
		os.Exit(1)
	}

	return address, firstPort, lastPort, bufferChan, saveFormat, document
}

func SaveResult(result []int, format string, address string) error {
	if _, err := os.Stat("result"); os.IsNotExist(err) {
		err := os.Mkdir("result", 0755)
		if err != nil {
			fmt.Println("Couldn't create a folder:", err)
			return err
		}
	}

	if format == "txt" {
		file, err := os.Create(fmt.Sprintf("result/OpenPorts_%s.txt", address))
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
		file, err := os.Create(fmt.Sprintf("result/OpenPorts_%s.json", address))
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

func ReadAddressDocument(doc string) []string {
	file, _ := os.Open(doc)
	defer file.Close()

	scanner := bufio.NewScanner(file)

	result := []string{}

	for scanner.Scan() {
		result = append(result, strings.TrimSpace(scanner.Text()))
	}
	return result
}

func worker(ports chan int, result chan int, adrs string) {
	for port := range ports {
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

func starter(address string, firstPort int, lastPort int, bufferChan int, saveFormat string) {
	ports := make(chan int, bufferChan)
	results := make(chan int)
	var openPorts []int

	for i := 0; i < cap(ports); i++ {
		go worker(ports, results, address)
	}

	fmt.Printf("\nScanning %s...", address)
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

	fmt.Println("\nResults for " + address)
	fmt.Println("Scanning time:", elapsedTime)
	if len(openPorts) != 0 {
		fmt.Printf("Found: %d element\nOpen ports:\n", len(openPorts))
		for _, port := range openPorts {
			fmt.Printf("%d open\n", port)
		}

		if saveFormat != "" {
			err := SaveResult(openPorts, saveFormat, address)
			if err != nil {
				os.Exit(1)
			}
		}
	} else {
		fmt.Println("No port is open")
	}
}

func main() {
	address, firstPort, lastPort, bufferChan, saveFormat, document := InitFlags()

	if firstPort > lastPort {
		fmt.Println("Error: The first port cannot exceed the last port.")
	} else if firstPort <= 0 || lastPort <= 0 {
		fmt.Println("Error: The port cannot be <= 0.")
	} else {
		if address != "" && document == "" {
			fmt.Println("Press 'Enter' to start scan...")
			fmt.Scan(" ")

			starter(address, firstPort, lastPort, bufferChan, saveFormat)
		} else if address == "" && document != "" {
			fmt.Println("Press 'Enter' to start scan...")
			fmt.Scan(" ")

			addressList := ReadAddressDocument(document)

			if len(addressList) != 0 {
				for _, adrs := range addressList {
					starter(adrs, firstPort, lastPort, bufferChan, saveFormat)
				}
			} else {
				fmt.Println("File empty")
				os.Exit(1)
			}
		} else {
			fmt.Println("Use only address or document")
			os.Exit(1)
		}
	}
}
