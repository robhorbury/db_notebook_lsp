package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"myfirstlsp/analysis"
	"myfirstlsp/lsp"
	"myfirstlsp/rpc"
	"os"
)

func main() {

	logger := getLogger("/Users/roberthorbury/Documents/myfirstlsp/log.txt")
	logger.Println("I started")

	installRuff(logger)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(rpc.Split)

	state := analysis.NewState()
	writer := os.Stdout

	for scanner.Scan() {
		msg := scanner.Bytes()
		method, contents, err := rpc.DecodeMessage(msg)
		if err != nil {
			logger.Printf("got an error: %s", err)
			continue
		}

		handleMessage(logger, writer, state, method, contents)
	}
}

func handleMessage(logger *log.Logger, writer io.Writer, state analysis.State, method string, contents []byte) {
	logger.Printf("recieved msg with method: %s", method)
	switch method {
	case "initialize":
		var request lsp.InitialiseRequest

		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Printf("Couldn't parse this %s", err)
		}

		logger.Printf("Connected to %s %s",
			request.Params.ClientInfo.Name,
			request.Params.ClientInfo.Version)

		//Reply:
		msg := lsp.NewInitialiseResponse(request.ID)
		writeResponse(writer, msg)

		logger.Printf("sent reply")

	case "textDocument/didOpen":
		var request lsp.DidOpenTextDocumentNotification
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Printf("Couldn't parse DidOpen this %s", err)
		}
		logger.Printf("Opened: %s", request.Params.TextDocument.URI)
		state.OpenDocument(request.Params.TextDocument.URI, request.Params.TextDocument.Text)

	case "textDocument/didChange":
		var request lsp.DidChangeTextDocumentNotification
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Printf("Couldn't handle DidChange this %s", err)
		}
		logger.Printf("Changed: %s", request.Params.TextDocument.URI)

		for _, change := range request.Params.ContentChanges {
			state.UpdateDocument(request.Params.TextDocument.URI, change.Text)
		}

	case "textDocument/hover":
		var request lsp.HoverRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Printf("textDocument/hover %s", err)
		}
		// Create a response
		response := state.Hover(request.ID, request.Params.TextDocument.URI, request.Params.Position, logger)
		//Write it back:

		if response != nil {
			writeResponse(writer, response)
		}

	}
}

func getLogger(filename string) *log.Logger {
	logfile, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		panic("bad file handed to logger")
	}

	return log.New(logfile, "[myfirstlsp]", log.Ldate|log.Ltime|log.Lshortfile)
}

func writeResponse(writer io.Writer, msg any) {
	reply := rpc.EncodeMessage(msg)
	writer.Write([]byte(reply))
}

func installRuff(logger *log.Logger) {
	files, err := os.ReadDir(".")

	if err != nil {
		logger.Println("errro in reading dir")
	}

	logger.Println(files)
}
