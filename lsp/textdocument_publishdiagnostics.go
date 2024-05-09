package lsp

type PublishDiagnosticParams struct {
	URI        string       `json:"uri"`
	Diagnostic []Diagnostic `json:"diagnostics"`
}

type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"`
	Code     string `json:"code"`
	Source   string `json:"source"`
	Message  string `json:"message"`
}

type Range struct {
	StartPosition Position `json:"start"`
	EndPosition   Position `json:"end"`
}

type PublishDiagnosticNotification struct {
	Notification
	Params PublishDiagnosticParams `json:"params"`
}
