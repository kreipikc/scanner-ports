package main

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

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
	viper.SetConfigFile("./envs/.env")
	viper.ReadInConfig()

	address := viper.Get("ADDRESS").(string)
	firstPort, _ := strconv.Atoi(viper.Get("FIRST_PORT").(string))
	lastPort, _ := strconv.Atoi(viper.Get("LAST_PORT").(string))
	buffer_chan, _ := strconv.Atoi(viper.Get("MAX_BUFFER_CHAN").(string))

	if firstPort > lastPort {
		fmt.Println("Error: The first port cannot exceed the last port.")
	} else if firstPort <= 0 || lastPort <= 0 {
		fmt.Println("Error: The port cannot be <= 0.")
	} else {
		fmt.Println("Press 'Enter' to start scan...")
		fmt.Scanf(" ")

		ports := make(chan int, buffer_chan)
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
		fmt.Println("Open ports: ")
		for _, port := range openPorts {
			fmt.Printf("%d open\n", port)
		}
		fmt.Scanf(" ")
	}
}
