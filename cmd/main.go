package main

import (
	"fmt"
	"github.com/mkaminski/goaim/oscar"
	"io"
	"log"
	"net"
	"net/http"
	"os"
)

func main() {

	go listenBOS()
	go listenStats()
	go listenAlert()
	go listenOdir()

	//todo implement CHATNAV and ALERT

	// Listen on TCP port 5190
	listener, err := net.Listen("tcp", ":5190")
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	fmt.Println("Server is listening on port 5190")

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		// Handle connection in a separate goroutine
		go handleAuthConnection(conn)
	}
}

func webServer(ch chan bool) {
	http.HandleFunc("/send-im", func(w http.ResponseWriter, r *http.Request) {
		ch <- true
	})

	if err := http.ListenAndServe(":3333", nil); err != nil {
		panic(err.Error())
	}
}

func listenBOS() {
	// Listen on TCP port 5190
	listener, err := net.Listen("tcp", ":5191")
	if err != nil {
		log.Fatal(err)
	}

	ch := make(chan bool, 1)
	go webServer(ch)

	defer listener.Close()

	fmt.Println("Server is listening on port 5191")

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		seq := uint32(100)
		go handleBOSConnection(conn, &seq)
		go sendIM(conn, ch, &seq)
	}
}

func sendIM(conn net.Conn, ch chan bool, u *uint32) {
	for range ch {
		fmt.Println("sending im...")
	}
}

func listenStats() {
	// Listen on TCP port 5190
	listener, err := net.Listen("tcp", ":5192")
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	fmt.Println("Server is listening on port 5192")

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		fmt.Println("got a connection on listenStats")
		seq := uint32(100)
		if err := oscar.ReadBos(conn, &seq); err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		}
	}
}

func listenAlert() {
	// Listen on TCP port 5190
	listener, err := net.Listen("tcp", ":5193")
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	fmt.Println("Server is listening on port 5193")

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		fmt.Println("got a connection on listenAlert")
		seq := uint32(100)
		if err := oscar.ReadBos(conn, &seq); err != nil && err != io.EOF {
			if err == io.EOF {
				break
			} else {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		}
	}
}

func listenOdir() {
	// Listen on TCP port 5190
	listener, err := net.Listen("tcp", ":5194")
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	fmt.Println("Server is listening on port 5194")

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		fmt.Println("got a connection on listenOdir")
		seq := uint32(100)
		if err := oscar.ReadBos(conn, &seq); err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		}
	}
}

func handleAuthConnection(conn net.Conn) {
	defer conn.Close()
	seq := uint32(100)
	err := oscar.SendAndReceiveSignonFrame(conn, &seq)
	if err != nil {
		log.Println(err)
		return
	}

	err = oscar.ReceiveAndSendAuthChallenge(conn, conn, &seq)
	if err != nil {
		log.Println(err)
		return
	}

	err = oscar.ReceiveAndSendBUCPLoginRequest(conn, conn, &seq)
	if err != nil {
		log.Println(err)
		return
	}
}

func handleBOSConnection(conn net.Conn, seq *uint32) {
	//defer conn.Close()
	fmt.Println("SendAndReceiveSignonFrame...")
	if err := oscar.SendAndReceiveSignonFrame(conn, seq); err != nil {
		log.Println(err)
		return
	}

	fmt.Println("writeOServiceHostOnline...")
	if err := oscar.WriteOServiceHostOnline(conn, seq); err != nil {
		if err == io.EOF {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}

	if err := oscar.ReadBos(conn, seq); err != nil && err != io.EOF {
		if err != io.EOF {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}
}
