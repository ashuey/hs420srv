package main

import (
	"github.com/bmizerany/pat"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

var switchCommands = [2][4][]byte{
	{
		{0x00, 0xff, 0xd5, 0x7b},
		{0x01, 0xfe, 0xd5, 0x7b},
		{0x02, 0xfd, 0xd5, 0x7b},
		{0x03, 0xfc, 0xd5, 0x7b},
	},
	{
		{0x04, 0xfb, 0xd5, 0x7b},
		{0x05, 0xfa, 0xd5, 0x7b},
		{0x06, 0xf9, 0xd5, 0x7b},
		{0x07, 0xf8, 0xd5, 0x7b},
	},
}

var powerCommand = []byte{0x10, 0xef, 0xd5, 0x7b}

type HandlerContext struct {
	port *os.File
}

func (ctx *HandlerContext) switchInput(w http.ResponseWriter, r *http.Request) {
	tv, err := strconv.Atoi(r.URL.Query().Get(":tv"))
	tv -= 1

	if err != nil {
		log.Print(err)
		http.Error(w, "Invalid output identifier", http.StatusBadRequest)
		return
	}

	input, err := strconv.Atoi(r.URL.Query().Get(":input"))
	input -= 1

	if err != nil {
		log.Print(err)
		http.Error(w, "Invalid source identifier", http.StatusBadRequest)
		return
	}

	if tv < 0 || tv >= len(switchCommands) {
		log.Print("TV choice out of range")
		http.Error(w, "Invalid output identifier", http.StatusBadRequest)
		return
	}

	inputChoices := switchCommands[tv]

	if input < 0 || input >= len(inputChoices) {
		log.Print("Input choice out of range")
		http.Error(w, "Invalid source identifier", http.StatusBadRequest)
		return
	}

	serialBytes := inputChoices[input]

	_, err = ctx.port.Write(serialBytes)

	if err != nil {
		log.Print(err)
		http.Error(w, "Error communicating with switch", http.StatusInternalServerError)
		return
	}

	_, _ = io.WriteString(w, "OK")
}

func (ctx *HandlerContext) powerCycle(w http.ResponseWriter, _ *http.Request) {
	_, err := ctx.port.Write(powerCommand)

	if err != nil {
		log.Print(err)
		http.Error(w, "Error communicating with switch", http.StatusInternalServerError)
		return
	}

	_, _ = io.WriteString(w, "OK")
}

func version(w http.ResponseWriter, _ *http.Request) {
	_, _ = io.WriteString(w, "1.0.0")
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s port_name", os.Args[0])
	}

	serialName := os.Args[1]

	s, err := os.OpenFile(serialName, os.O_WRONLY, 0660)

	if err != nil {
		log.Fatal("Error connecting to serial port")
	}

	ctx := &HandlerContext{s}

	router := pat.New()

	router.Post("/switch/:tv/input/:input", http.HandlerFunc(ctx.switchInput))
	router.Post("/power", http.HandlerFunc(ctx.powerCycle))
	router.Get("/version", http.HandlerFunc(version))

	http.Handle("/", router)

	log.Fatal(http.ListenAndServe(":10330", nil))
}
