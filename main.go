package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"myfirstlsp/analysis"
	"myfirstlsp/lsp"
	"myfirstlsp/rpc"
	"os"
	"strings"

	"golang.org/x/exp/maps"
)

func main() {
	err := os.MkdirAll("./.customLsp/.tempFiles/", os.ModePerm)
	if err != nil {
		panic(err)
	}
	logger := getLogger("./.customLsp/log.txt")
	logger.Println("I started")
	defer printError(logger)

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

func printError(logger *log.Logger) {
	r := recover()

	if r != nil {
		logger.Println(r)
		panic(r)
	}
}

func handleMessage(logger *log.Logger, writer io.Writer, state analysis.State, method string, contents []byte) {
	logger.Printf("recieved msg with method: %s", method)
	switch method {
	case "initialize":

		createCacheDirectory(logger)

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
		err := state.CacheDocument(request.Params.TextDocument.URI)
		if err != nil {
			logger.Printf("Error Caching: %s", err)
		}

		err = state.LintDocument(request.Params.TextDocument.URI, logger)
		if err != nil {
			logger.Printf("Error Linting: %s", err)
		}

		response := state.PublishDiagnostics(request.Params.TextDocument.URI, logger)
		writeResponse(writer, response)

	case "textDocument/didChange":
		var request lsp.DidChangeTextDocumentNotification
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Printf("Couldn't handle DidChange this %s", err)
		}
		logger.Printf("Changed: %s", request.Params.TextDocument.URI)

		for _, change := range request.Params.ContentChanges {
			state.UpdateDocument(request.Params.TextDocument.URI, change.Text)
			err := state.CacheDocument(request.Params.TextDocument.URI)

			if err != nil {
				logger.Printf("Error Caching: %s", err)
			}

			err = state.LintDocument(request.Params.TextDocument.URI, logger)
			if err != nil {
				logger.Printf("Error Linting: %s", err)
			}

			response := state.PublishDiagnostics(request.Params.TextDocument.URI, logger)
			logger.Println("Published Diagnostics")
			writeResponse(writer, response)

		}

	case "textDocument/semanticTokens/full":
		var request lsp.SemanticTokenRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			logger.Printf("textDocument/sematicTokens/full %s", err)
		}

		response := state.SemanticFormat(request.ID, request.Params.TextDocument.URI, logger)

		if response != nil {
			writeResponse(writer, response)
			logger.Println("Responded")
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

	case "shutdown":
		keys := maps.Keys(state.Documents)
		filePath := analysis.GetTempPath()

		for _, d := range keys {
			fileName := analysis.GetTempFileName(d)

			err := os.Remove(fmt.Sprintf("%s.temp_%s", filePath, fileName))

			if err != nil {
				logger.Printf("could not clean up file %s", fmt.Sprintf("%s.temp_%s", filePath, fileName))
			}
		}

		pathElements := strings.Split(filePath, "/")

		err := os.RemoveAll(strings.Join(pathElements[:len(pathElements)-2], "/"))

		if err != nil {
			logger.Printf("Could not delete file: %s", strings.Join(pathElements[:len(pathElements)-1], "/"))
		}

		logger.Printf("Deleted: %s", strings.Join(pathElements[:len(pathElements)-1], "/"))

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

func createCacheDirectory(logger *log.Logger) {
	err := os.MkdirAll("./.customLsp/.tempFiles/", os.ModePerm)

	logger.Println("Creating temp folder")

	if err != nil {
		logger.Println("Error Creating File")
	}

	f, err := os.Create("./.customLsp/.gitignore")

	logger.Println("Creating GitIgnore")

	if err != nil {
		logger.Println("Error Creating gitignore")
	}

	logger.Println("Writing to gitignore")

	l, err := f.WriteString("*")
	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}
	fmt.Println(l, "bytes written successfully")
	err = f.Close()

	if err != nil {
		logger.Println("Error Closing file")
	}
}
