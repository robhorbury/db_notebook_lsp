package lsp

type InitialiseRequest struct {
	Request
	Params InitialiseRequestParams `json:"params"`
}

type InitialiseRequestParams struct {
	ClientInfo *ClientInfo `json:"clientInfo"`
	// ..... More to add here!
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitialiseResponse struct {
	Response
	Result InitialiseResult `json:"result"`
}

type InitialiseResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   ServerInfo         `json:"serverInfo"`
}

type ServerCapabilities struct {
	TextDocumentSync       int                  `json:"textDocumentSync"`
	HoverProvider          bool                 `json:"hoverProvider"`
	SemanticTokensProvider SematicTokensOptions `json:"semanticTokensProvider"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func NewInitialiseResponse(id int) InitialiseResponse {
	return InitialiseResponse{
		Response: Response{
			RPC: "2.0",
			ID:  &id,
		},
		Result: InitialiseResult{
			Capabilities: ServerCapabilities{
				TextDocumentSync: 1,
				HoverProvider:    true,
				SemanticTokensProvider: SematicTokensOptions{
					Legend: SemanticTokensLegend{
						TokenTypes:     []string{"namespace", "property", "method", "string"},
						TokenModifiers: []string{},
					},
					Full: true,
				},
			},
			ServerInfo: ServerInfo{
				Name:    "myfirstlsp",
				Version: "0.0.1",
			},
		},
	}
}
