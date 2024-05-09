package lsp

type SemanticTokensLegend struct {
	TokenTypes     []string `json:"tokenTypes"`
	TokenModifiers []string `json:"tokenModifiers"`
}

type SematicTokensOptions struct {
	Legend SemanticTokensLegend `json:"legend"`
	Full   bool                 `json:"full"`
}

type SemanticTokenRequest struct {
	Request
	Params TextDocumentSemanticTokenParams
}

type SemanticTokenResponse struct {
	Response
	Result SemanticTokenResult `json:"result"`
}

type SemanticTokenResult struct {
	Data []uint `json:"data"`
}
