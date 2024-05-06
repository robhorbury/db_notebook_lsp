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
	Documents     map[string]string
	LinterResults map[string]string
}

func NewState() State {
	return State{Documents: map[string]string{},
		LinterResults: map[string]string{}}
}

func (s *State) OpenDocument(uri, text string) {
	s.Documents[uri] = text
}

func (s *State) UpdateDocument(uri, text string) {
	s.Documents[uri] = text
}

func (s *State) CacheDocument(uri string) error {

	filePath := GetTempPath()
	fileName := GetTempFileName(uri)

	doc := newDocument(s.Documents[uri])

	err := os.WriteFile(fmt.Sprintf("%s.temp_%s", filePath, fileName), []byte(doc.contents), 0644)

	return err
}

func (s *State) LintDocument(uri string) error {
	execPath, err := exec.LookPath("ruff")
	if err != nil {
		return err
	}

	filePath := GetTempPath()
	fileName := GetTempFileName(uri)

	linterRes, err := getLintedResults(execPath, fmt.Sprintf("%s.temp_%s", filePath, fileName))

	if err != nil {
		return fmt.Errorf("Error: %s: Linter Result: %s", err, linterRes)
	}
	s.LinterResults[uri] = linterRes
	return nil
}

func (s *State) Hover(id int, uri string, position lsp.Position, logger *log.Logger) *lsp.HoverResponse {

	res := s.LinterResults[uri]

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

	if fmt.Sprintf("%s", err) == "exit status 2" {
		return out.String(), err

	} else {
		return out.String(), nil
	}

}

func parseLinterMessages(messages string, lineNo int, logger *log.Logger) *string {

	switch messages {
	case "All checks passed successfully":
		return nil
	}

	lines := strings.Split(messages, "\n")

	for _, l := range lines {
		logger.Printf("    %s \n", l)
	}

	lines = remove(lines, ".py")
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
