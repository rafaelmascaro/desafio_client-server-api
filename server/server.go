package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type (
	Cotacao struct {
		USDBRL ItemCotacao `json:"USDBRL"`
	}

	ItemCotacao struct {
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

	RetornoCotacao struct {
		Bid string `json:"bid"`
	}
)

var db *sql.DB

func main() {
	var err error
	db, err = dbConfig()
	if err != nil {
		panic(err)
	}
	defer db.Close()
	http.HandleFunc("/cotacaco", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cotacao, err := consultaCotacao(ctx)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("erro ao buscar cotação: %v\n", err)
		return
	}

	err = salvaCotacao(db, &cotacao.USDBRL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("erro ao persistir dados: %v\n", err)
		return
	}

	ret := RetornoCotacao{Bid: cotacao.USDBRL.Bid}
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("erro ao serializar cotação: %v\n", err)
		return
	}
}

func consultaCotacao(ctx context.Context) (*Cotacao, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var cotacao Cotacao
	err = json.Unmarshal(data, &cotacao)
	if err != nil {
		return nil, err
	}

	return &cotacao, nil
}

func dbConfig() (*sql.DB, error) {
	db, err := sql.Open("sqlite", ".cotacao.db")
	if err != nil {
		return nil, err
	}

	createTable := `create table if not exists cotacoes (
		"code" text,
		"codein" text,
		"name" text,
		"high" text,
		"low" text,
		"varBid" text,
		"pctChange" text,
		"bid" text,
		"ask" text,
		"timestamp" text,
		"create_date" text
	)`

	stmt, err := db.Prepare(createTable)

	if err != nil {
		return nil, err
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func salvaCotacao(db *sql.DB, item *ItemCotacao) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	insertSQL := "insert into cotacoes(code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, create_date) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	stmt, err := db.PrepareContext(ctx, insertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, item.Code, item.Codein, item.Name, item.High, item.Low, item.VarBid, item.PctChange, item.Bid, item.Ask, item.Timestamp, item.CreateDate)
	return err
}
