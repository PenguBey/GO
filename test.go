package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/goburrow/modbus"
	"github.com/tarm/serial"
)

type yanit struct {
	Veri []int16 `json:"data"`
}

func main() {

	seriPortKonfig := &serial.Config{
		Name:     "COM3",
		Baud:     19200,
		Parity:   serial.ParityNone,
		StopBits: serial.Stop1,
		Size:     8,
	}

	yonetici := modbus.NewRTUClientHandler(seriPortKonfig.Name)
	yonetici.BaudRate = seriPortKonfig.Baud
	yonetici.Parity = string(seriPortKonfig.Parity)
	yonetici.Timeout = 10 * time.Second

	istemci := modbus.NewClient(yonetici)

	ustID := byte(10)
	adresi := uint16(9)
	miktar := uint16(10)

	yonetici.SlaveId = ustID

	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {

		sonuçlar, err := istemci.ReadHoldingRegisters(adresi, miktar)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(sonuçlar) < int(miktar)*2 {
			http.Error(w, "No data returned", http.StatusInternalServerError)
			return
		}

		var veri []int16
		for i := 0; i < len(sonuçlar); i += 2 {
			değer := int16(sonuçlar[i])<<8 | int16(sonuçlar[i+1])
			veri = append(veri, değer)
		}

		yanit := yanit{Veri: []int16{}}
		if len(veri) > 0 {
			yanit.Veri = append(yanit.Veri, veri[0])
		}
		if len(veri) > 4 {
			yanit.Veri = append(yanit.Veri, veri[4])
		}

		jsonyanit, err := json.Marshal(yanit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonyanit)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
