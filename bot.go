package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
)

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

var ansi = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansi.ReplaceAllString(s, "")
}

func main() {
	conn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Fprintln(conn, "BotInversor")
	log.Println("Enviado: BotInversor")

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := scanner.Text()
		log.Printf("Mensagem recebida: %q", msg)

		cleanMsg := stripANSI(msg)
		log.Printf("Mensagem limpa: %q", cleanMsg)

		if strings.Contains(cleanMsg, "BotInversor") {
			continue
		}

		if strings.Contains(cleanMsg, "disse") {
			parts := strings.Split(cleanMsg, ":")
			if len(parts) > 1 {
				original := strings.TrimSpace(parts[1])
				log.Printf("Mensagem original: %q", original)

				inverted := reverse(original)
				log.Printf("Mensagem invertida: %q", inverted)

				fmt.Fprintln(conn, inverted)
				log.Printf("Enviado: %q", inverted)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println("Erro ao escanear:", err)
	}
}
