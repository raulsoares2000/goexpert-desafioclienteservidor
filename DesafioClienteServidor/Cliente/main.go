package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	cliente()
}

func cliente() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Printf("Erro ao criar requisição: %v", err)
		return
	}

	res, err := client.Do(req)
	if err != nil {
		if errors.Is(err, ctx.Err()) {
			log.Print("Erro ao fazer requisição: Timeout")
			return
		} else {
			log.Print("Erro ao ler requisição")
			return
		}
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Erro ao criar body: %v", err)
		return
	}

	var cotacaoBid string
	err = json.Unmarshal(body, &cotacaoBid)
	if err != nil {
		log.Printf("Erro ao ler bid: %v", err)
		return
	}

	file, err := os.Create("cotacao.txt")
	if err != nil {
		log.Printf("Erro ao criar arquivo %v", err)
	}
	defer file.Close()
	_, err = file.Write([]byte(cotacaoBid))
	if err != nil {
		log.Printf("Erro ao escrever no arquivo %v", err)
	}
}
