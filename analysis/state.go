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
			stringTokenList = append(stringTokenList, CreateStringTokens(cell, cellStartLineNo[i], "'", logger)...)
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
		logger.Printf("Going into encoding: %d %d", t.absLineNo, t.absStartIndex)
		prevLineNo, prevStartIndex = encodeToken(&t, prevLineNo, prevStartIndex)

		newList = append(newList, t)
		logger.Printf("After encoding: %d %d", *t.relativeLineNo, *t.relativeStartIndex)

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
