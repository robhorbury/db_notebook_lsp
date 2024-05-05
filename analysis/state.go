package analysis

import (
	"bytes"
	"fmt"
	"log"
	"myfirstlsp/lsp"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type State struct {
	Documents map[string]string
}

func NewState() State {
	return State{Documents: map[string]string{}}
}

func (s *State) OpenDocument(uri, text string) {
	s.Documents[uri] = text
}

func (s *State) UpdateDocument(uri, text string) {
	s.Documents[uri] = text
}

func (s *State) Hover(id int, uri string, position lsp.Position, logger *log.Logger) *lsp.HoverResponse {

	doc := newDocument(s.Documents[uri])
	fileName := strings.ReplaceAll(uri, ":", "_")
	fileName = strings.ReplaceAll(fileName, "/", "_")
	fileName = strings.ReplaceAll(fileName, "\\", "_")

	err := os.WriteFile(fmt.Sprintf("/Users/roberthorbury/Documents/myfirstlsp/.temp_%s", fileName), []byte(doc.contents), 0644)

	path, err := exec.LookPath("ruff")
	if err != nil {
		logger.Println("installing ruff is not  in your future")
	}

	logger.Println(path)

	if err != nil {
		logger.Println(err)
	}

	res, err := getLintedResults(path, fmt.Sprintf("/Users/roberthorbury/Documents/myfirstlsp/.temp_%s", fileName))

	if err != nil {
		logger.Println(err)
	}

	lineNoMessage := parseLinterMessages(res, position.Line+1, logger)

	if lineNoMessage != nil {
		value := *lineNoMessage
		logger.Printf("Errors Message: %s", value)

		response := lsp.HoverResponse{
			Response: lsp.Response{
				RPC: "2.0",
				ID:  &id,
			},
			Result: lsp.HoverResult{
				Contents: value,
			},
		}
		return &response
	} else {
		logger.Printf("No Linter message")
	}
	return nil
}

func newDocument(contents string) *document {
	doc := document{contents: contents}
	return &doc
}

type document struct {
	contents string
}

func (d *document) getLines() []string {
	return (strings.Split(d.contents, "\n"))
}

func (d *document) getLine(id int) string {
	return d.getLines()[id]
}

func (d *document) getHoverString(id int) string {
	line := d.getLine(id)
	if isLineAComment(line) {
		return "It is a comment"
	}
	return "Not a comment"
}

func isLineAComment(line string) bool {
	return strings.Contains(line, "#")
}

func getLintedResults(execPath string, id string) (string, error) {

	command := exec.Command(execPath, "check", id)

	// set var to get the output
	var out bytes.Buffer

	// set the output to our variable
	command.Stdout = &out
	err := command.Run()

	if out.String() == "" {
		command.Stderr = &out
		err = command.Run()
	}

	return out.String(), err

}

func parseLinterMessages(messages string, lineNo int, logger *log.Logger) *string {

	switch messages {
	case "All checks passed successfully":
		return nil
	}

	lines := strings.Split(messages, "\n")
	lines = remove(lines, ".py")

	for _, l := range lines {
		logger.Printf("    %s", l)
	}

	for _, line := range lines {

		errorMessages := strings.Split(line, ".py:")
		errorMessage := errorMessages[len(errorMessages)-1]
		errorLineNumber, err := strconv.Atoi(strings.Split(errorMessage, ":")[0])

		if err != nil {
			logger.Printf("Errorr: %s", err)
			panic(err)
		}

		if errorLineNumber == lineNo {
			styleViolations := strings.Split(errorMessage, ":")[2:]
			styleViolationsString := strings.Join(styleViolations, " ")

			return &styleViolationsString
		}

	}
	return nil
}

func remove(s []string, match string) []string {
	number_of_replaces := 0
	for i, e := range s {
		if !strings.Contains(e, match) {
			number_of_replaces++
			s[i] = s[len(s)-number_of_replaces]
		}
	}

	if number_of_replaces > 0 {
		return s[:len(s)-number_of_replaces]
	}
	return s
}
