package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Payment struct {
	Description string    `json:"description"`
	USD         int       `json:"usd"`
	FullName    string    `json:"full_name"`
	Address     string    `json:"address"`
	Time        time.Time `json:"time"`
}

type httpResponse struct {
	Money          int       `json:"money"`
	PaymentHistory []Payment `json:"history"`
}

var mtx = sync.Mutex{}
var money = 1_000_000
var paymentHistory = make([]Payment, 0)

func (p Payment) Println() {
	fmt.Println("Description:", p.Description)
	fmt.Println("USD:", p.USD)
	fmt.Println("FullName:", p.FullName)
	fmt.Println("Address:", p.Address)
}

func payHandler(w http.ResponseWriter, r *http.Request) {
	var payment Payment
	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payment.Time = time.Now()

	payment.Println()

	mtx.Lock()
	if money-payment.USD >= 0 {
		money -= payment.USD
	} else {
		_, err := w.Write([]byte("Недостаточно средств"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println(err)
			return
		}
	}
	paymentHistory = append(paymentHistory, payment)

	httpResponse := httpResponse{
		Money:          money,
		PaymentHistory: paymentHistory,
	}

	SliceOfByte, err := json.Marshal(httpResponse)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	if _, err := w.Write(SliceOfByte); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	mtx.Unlock()

}

func main() {
	http.HandleFunc("/pay", payHandler)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
