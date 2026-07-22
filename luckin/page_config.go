package luckin

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/anaskhan96/soup"
	"github.com/dop251/goja/ast"
	"github.com/dop251/goja/parser"
)

const baseUrl = "https://in.luckincoffee.com"

type PageConfig struct {
	CFCenter struct {
		EncryptKey string `json:"encryptKey"`
	} `json:"cfcenter"`
}

func getEncryptKey() (string, error) {
	resp, err := soup.Get(baseUrl)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	doc := soup.HTMLParse(resp)
	links := doc.FindAll("script")[3]

	js := links.Text()
	program, err := parser.ParseFile(nil, "", js, 0)
	if err != nil {
		return "", fmt.Errorf("failed to parse JavaScript: %v", err)
	}
	var pageConfigRaw string
	for _, stmt := range program.Body {
		// find variable for window.pageConfig
		if decl, ok := stmt.(*ast.ExpressionStatement); ok {
			if assign, ok := decl.Expression.(*ast.AssignExpression); ok {
				if ident, ok := assign.Left.(*ast.DotExpression); ok {
					if ident.Left.(*ast.Identifier).Name == "window" && ident.Identifier.Name == "pageConfig" {
						// log.Printf("Found window.pageConfig assignment: %v", assign.Right)
						if strLit, ok := assign.Right.(*ast.StringLiteral); ok {
							pageConfigRaw = string(strLit.Value)
							break
						}
					}
				}
			}
		}
	}
	if pageConfigRaw == "" {
		return "", fmt.Errorf("failed to find window.pageConfig assignment")
	}

	var pageConfig PageConfig
	pageConfigDecoded, err := base64.StdEncoding.DecodeString(pageConfigRaw)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 pageConfig: %v", err)
	}
	err = json.Unmarshal(pageConfigDecoded, &pageConfig)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal pageConfig: %v", err)
	}
	// Do something with the parsed program, e.g., analyze it or extract information
	return pageConfig.CFCenter.EncryptKey, nil
}
