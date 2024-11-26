package main

import (
	"context"
	"encoding/json"
	"errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"time"
)

type Cotacao struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

type CotacaoFormatada struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

func main() {
	createTable()
	http.HandleFunc("/cotacao", CotacaoHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Printf("Erro ao gerar servidor: %v", err)
	}
}

func CotacaoHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	cotacao, err := BuscaCotacao()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(cotacao.USDBRL.Bid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func BuscaCotacao() (*Cotacao, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}
	select {
	case <-ctx.Done():
		log.Print("Erro ao fazer requisição: Timeout")
		return nil, ctx.Err()
	default:
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Print("Erro ao ler requisição")
			return nil, err
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var c Cotacao
		err = json.Unmarshal(body, &c)
		if err != nil {
			return nil, err
		}
		insertDB(&c)
		return &c, nil
	}
}

func insertDB(cotacao *Cotacao) {
	dsn := "gorm.db"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Print("Erro ao conectar com banco de dados")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	insert := newInsert(cotacao)

	result := db.WithContext(ctx).Create(&insert)
	if result.Error != nil {
		if errors.Is(result.Error, ctx.Err()) {
			log.Print("Erro ao inserir no banco de dados: Timeout")
			return
		} else {
			log.Print("Erro ao inserir no banco de dados")
			return
		}
	}
}

func createTable() {
	dsn := "gorm.db"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Print("Erro ao conectar com banco de dados")
		return
	}

	err = db.AutoMigrate(&CotacaoFormatada{})
	if err != nil {
		log.Print("Erro ao criar tabela")
		return
	}
}

func newInsert(cotacao *Cotacao) *CotacaoFormatada {
	var c CotacaoFormatada
	c.Code = cotacao.USDBRL.Code
	c.Codein = cotacao.USDBRL.Codein
	c.Name = cotacao.USDBRL.Name
	c.High = cotacao.USDBRL.High
	c.Low = cotacao.USDBRL.Low
	c.VarBid = cotacao.USDBRL.VarBid
	c.PctChange = cotacao.USDBRL.PctChange
	c.Bid = cotacao.USDBRL.Bid
	c.Ask = cotacao.USDBRL.Ask
	c.Timestamp = cotacao.USDBRL.Timestamp
	c.CreateDate = cotacao.USDBRL.CreateDate
	return &c
}
