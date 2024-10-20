package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Print("Digite seu apelido: ")
	reader := bufio.NewReader(os.Stdin)
	nickname, _ := reader.ReadString('\n')
	nickname = strings.TrimSpace(nickname)
	fmt.Fprintf(conn, nickname+"\n")

	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	for {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		if text == "\\exit" {
			fmt.Println("Saindo do chat...")
			break
		}

		fmt.Fprintln(conn, text)
	}

	conn.Close()
}
