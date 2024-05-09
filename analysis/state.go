package analysis

import (
	"bytes"
	"fmt"
	"log"
	"myfirstlsp/lsp"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

func newDocument(contents string) *document {
	doc := document{contents: contents}
	return &doc
}

type document struct {
	contents string
}

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

func (s *State) LintDocument(uri string, logger *log.Logger) error {
	execPath, err := exec.LookPath("ruff")
	if err != nil {
		return err
	}

	filePath := GetTempPath()
	fileName := GetTempFileName(uri)

	linterRes, err := getLintedResults(execPath, fmt.Sprintf("%s.temp_%s", filePath, fileName))
	if err != nil {
		return fmt.Errorf("Error: %s: linter Result: %s", err, linterRes)
	}
	execPath, err = exec.LookPath("mypy")
	logger.Println(execPath)
	if err != nil {
		logger.Printf("Error in pyright: %s", err)
		return err
	}
	typeRes, err := getTypeResults(execPath, fmt.Sprintf("%s.temp_%s", filePath, fileName), logger)

	if err != nil {
		return fmt.Errorf("Error: %s: type Result: %s", err, linterRes)
	}

	lines := strings.Split(linterRes, "\n")
	lines = append(lines, strings.Split(typeRes, "\n")...)
	//lines := strings.Split(typeRes, "\n")

	var pythonLines []string
	for _, l := range lines {
		if strings.Contains(l, ".py") {
			pythonLines = append(pythonLines, l)
		}
	}

	s.LinterResults[uri] = strings.Join(pythonLines, "\n")

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

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func (s *State) PublishDiagnostics(uri string, logger *log.Logger) *lsp.PublishDiagnosticNotification {
	logger.Println(s.LinterResults[uri])
	lines := strings.Split(s.LinterResults[uri], "\n")
	var errorMsgs []errorMessage
	var diagnostics []lsp.Diagnostic

	isNotebook := strings.Contains(s.Documents[uri], "# Databricks notebook source")

	for _, line := range lines {
		if strings.Contains(line, ".py") {
			errorStr := strings.Split(line, ".py:")
			errorString := errorStr[len(errorStr)-1]
			lineNo, err := strconv.Atoi(strings.Split(errorString, ":")[0])
			panicOnErr(err)
			char, err := strconv.Atoi(strings.Split(errorString, ":")[1])
			panicOnErr(err)
			errorString = strings.Join(strings.Split(errorString, ":")[2:], " ")

			errorString = strings.Trim(errorString, " ")
			errorWords := strings.Split(errorString, " ")

			code := errorWords[0]
			desc := strings.Join(errorWords[1:], " ")

			severity := 3
			if strings.Contains(code, "E") || strings.Contains(code, "warning") {
				severity = 2
			} else if strings.Contains(code, "F") || strings.Contains(code, "error") {
				severity = 1
			}

			if !strings.Contains(desc, "[name-defined]") && (isNotebook && !strings.Contains(desc, "Undefined name `spark`") && !strings.Contains(desc, "Undefined name `dbutils`")) {
				errorMsgs = append(errorMsgs, errorMessage{
					line:     lineNo,
					char:     char,
					code:     code,
					desc:     desc,
					source:   "Ruff",
					severity: severity,
				})
			}
		}
	}

	for _, msg := range errorMsgs {
		diagnostics = append(diagnostics, lsp.Diagnostic{
			Range: lsp.Range{
				StartPosition: lsp.Position{
					Line:      msg.line - 1,
					Character: 0},
				EndPosition: lsp.Position{
					Line:      msg.line - 1,
					Character: 1001},
			},
			Severity: msg.severity,
			Code:     msg.code,
			Source:   msg.source,
			Message:  msg.desc,
		})

	}

	response := lsp.PublishDiagnosticNotification{
		Notification: lsp.Notification{
			RPC:    "2.0",
			Method: "textDocument/publishDiagnostics",
		},
		Params: lsp.PublishDiagnosticParams{
			URI:        uri,
			Diagnostic: diagnostics,
		},
	}

	return &response

}

func getPyRightResults(uri string, logger *log.Logger) {

}

type errorMessage struct {
	line     int
	char     int
	code     string
	desc     string
	source   string
	severity int
}

func (s *State) SemanticFormat(id int, uri string, logger *log.Logger) *lsp.SemanticTokenResponse {

	doc := s.Documents[uri]

	isDatabricksNotebook := strings.Contains(strings.ToLower(doc), "# databricks notebook source")
	containsSql := strings.Contains(strings.ToLower(doc), "# magic %sql")

	if isDatabricksNotebook && containsSql {

		var intList []int
		var tokenList []token

		sqlCells, cellStartLineNo := splitIntoSQLCells(doc)
		for i, cell := range sqlCells {

			allTokenList := findTokenInCell(cell, cellStartLineNo[i], logger)
			stringTokenList := CreateStringTokens(cell, cellStartLineNo[i], "\"", logger)
			tokenList = append(tokenList, mergeTokenLists(allTokenList, stringTokenList)...)
			//tokenList = append(tokenList, stringTokenList...)
		}

		tokenList = orderTokenList(tokenList)
		intList = append(intList, encodeTokenList(tokenList, logger)...)

		response := lsp.SemanticTokenResponse{
			Response: lsp.Response{
				RPC: "2.0",
				ID:  &id,
			},

			Result: lsp.SemanticTokenResult{
				Data: intListToUint(intList),
			},
		}

		return &response
	} else {
		logger.Println("Not a notebook and doesn't contain sql")
	}
	return nil

}

func mergeTokenLists(allTokenList, stringTokenList []token) []token {
	var finalList []token
	var overlap bool
	for _, t1 := range allTokenList {
		overlap = false
		for _, t2 := range stringTokenList {
			if t1.absStartIndex > t2.absStartIndex && t1.absStartIndex < t2.absStartIndex+t2.length && t1.absLineNo == t2.absLineNo {
				overlap = true
			}

		}
		if !overlap {
			finalList = append(finalList, t1)
		}
	}

	finalList = append(finalList, stringTokenList...)

	return finalList
}

func intListToUint(intList []int) []uint {
	var newList []uint

	for _, i := range intList {
		newList = append(newList, uint(i))
	}
	return newList

}

type token struct {
	tokenValue         string
	absLineNo          int
	absStartIndex      int
	length             int
	relativeLineNo     *int
	relativeStartIndex *int
	tokenType          *int
	tokenModifiers     *int // TODO: make this generic
}

func findTokenInCell(cell string, startLineNo int, logger *log.Logger) []token {
	var tokenList []token

	cells := splitCellIntoLines(cell)

	for i, line := range cells {
		tokenList = append(tokenList,
			findTokenInLine(line, i+startLineNo, logger)...,
		)
	}
	return tokenList
}

func orderTokenList(inputList []token) []token {
	sort.Slice(inputList, func(i, j int) bool {
		if inputList[i].absLineNo != inputList[j].absLineNo {

			return inputList[i].absLineNo < inputList[j].absLineNo
		} else {
			return inputList[i].absStartIndex < inputList[j].absStartIndex
		}
	})

	return inputList

}

func encodeTokenList(inputList []token, logger *log.Logger) []int {

	var newList []token
	var intEncoded []int

	prevLineNo := 0
	prevStartIndex := 0
	for _, t := range inputList {
		prevLineNo, prevStartIndex = encodeToken(&t, prevLineNo, prevStartIndex)

		newList = append(newList, t)

	}

	for _, t := range newList {

		intEncoded = append(intEncoded, []int{
			int((*t.relativeLineNo)),
			int(*t.relativeStartIndex),
			int(t.length),
			int(*t.tokenType),
			int(*t.tokenModifiers)}...)

	}

	return intEncoded

}

func encodeToken(t *token, prevLineNo int, prevStartIndex int) (int, int) {

	if prevLineNo == t.absLineNo {
		relLineNo := 0
		relStartIndex := t.absStartIndex - prevStartIndex

		t.relativeLineNo = &relLineNo
		t.relativeStartIndex = &(relStartIndex)
	} else {
		relLineNo := t.absLineNo - prevLineNo
		t.relativeLineNo = &relLineNo
		startIndex := t.absStartIndex
		t.relativeStartIndex = &(startIndex)
	}

	word := strings.ReplaceAll(t.tokenValue, "(", "")
	word = strings.ReplaceAll(word, ")", "")

	if inList(word, GetSqlTokens()) {
		sqlTokenType := 1
		t.tokenType = &sqlTokenType
	} else if inList(word, GetSqlFunctions()) {
		sqlTokenType := 2
		t.tokenType = &sqlTokenType
	}

	return t.absLineNo, t.absStartIndex

}

func findTokenInLine(line string, lineNo int, logger *log.Logger) []token {

	var tokenList []token

	words := SplitStringWithPosition(line)

	for word, positions := range words {
		if word != "spaces" && word != "#" && word != "magic" {

			defaultTokentype := 0
			defaultTokenModifiers := 0

			for _, position := range positions {

				tokenList = append(tokenList,
					token{
						tokenValue:     word,
						absLineNo:      lineNo - 1,
						absStartIndex:  position,
						length:         len(word),
						tokenType:      &defaultTokentype,
						tokenModifiers: &defaultTokenModifiers,
					})
			}

		}
	}
	return tokenList
}

func splitCellIntoLines(cell string) []string {
	lines := strings.Split(cell, "\n")
	return lines
}

func splitIntoSQLCells(doc string) ([]string, []int) {
	allCells := strings.Split(strings.ToLower(doc), "# command ----------")

	var sqlCells []string
	var numberSqlCellLines []int
	var numberCellLines []int

	for _, cell := range allCells {
		lines := splitCellIntoLines(cell)
		numberCellLines = append(numberCellLines, len(lines))
	}

	for i, cell := range allCells {

		if strings.Contains(cell, "# magic %sql") {
			sqlCells = append(sqlCells, cell)
			numberSqlCellLines = append(numberSqlCellLines, sum(numberCellLines[:i]))

		}
	}
	return sqlCells, numberSqlCellLines
}

func sum(nums []int) int {
	total := 0

	for _, num := range nums {
		total += num
	}

	return total
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

func getTypeResults(execPath string, id string, logger *log.Logger) (string, error) {
	logger.Println("Getting Type Res::")

	logger.Println(id)

	command := exec.Command(execPath, id, "--show-column-number")
	// set var to get the output
	var out bytes.Buffer

	// set the output to our variable
	command.Stdout = &out
	err := command.Run()

	if out.String() == "" {
		command.Stderr = &out
		err = command.Run()
	}
	logger.Println(out.String())
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

func inList(s string, l []string) bool {
	for _, e := range l {
		if s == e {
			return true
		}
	}
	return false
}

func inListInt(s int, l []int) bool {
	for _, e := range l {
		if s == e {
			return true
		}
	}
	return false
}
