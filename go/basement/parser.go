package basement

import (
	"regexp"
	"strings"
)

var (
	// Block Regexes
	headerBlockRe = regexp.MustCompile(`^(\#{1,6})[ \t]+(.+)`)
	hrBlockRe     = regexp.MustCompile(`^(\*{3,}|-{3,}|_{3,})$`)
	listBlockRe   = regexp.MustCompile(`^([ \t]*)([*+-]|\d+\.)[ \t]+(.+)`)
	quoteBlockRe  = regexp.MustCompile(`^>[ \t]*(.+)`)
	codeFenceRe   = regexp.MustCompile(`^` + "```" + `(.*)`) // Capture language

	// Inline Regexes
	// Added (~~.+?~~) for Strikethrough
	inlineTokenRe = regexp.MustCompile(`(%v)|(\*\*.+?\*\*)|(\*.+?\*)|(__.+?__)|(~~.+?~~)|(!?#[a-zA-Z0-9]{3,8}\(.+?\))`)
)

// ParseAST parses the input string into an AST
func ParseAST(input string) *Node {
	root := NewNode(NodeRoot)
	lines := strings.Split(input, "\n")

	var currentList *Node
	var inCodeBlock bool
	var codeBlockLang string
	var codeBlockContent strings.Builder

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// 1. Handle Code Blocks (Stateful)
		if matches := codeFenceRe.FindStringSubmatch(trimmed); matches != nil {
			if inCodeBlock {
				// End of code block
				node := NewNode(NodeCodeBlock)
				node.Content = codeBlockContent.String()
				node.Lang = codeBlockLang
				root.AddChild(node)
				codeBlockContent.Reset()
				inCodeBlock = false
				codeBlockLang = ""
			} else {
				// Start of code block
				inCodeBlock = true
				codeBlockLang = strings.TrimSpace(matches[1])
			}
			continue
		}
		if inCodeBlock {
			codeBlockContent.WriteString(line + "\n")
			continue
		}

		// 2. Handle Lists (Stateful grouping)
		if matches := listBlockRe.FindStringSubmatch(line); matches != nil {
			if currentList == nil {
				currentList = NewNode(NodeList)
				root.AddChild(currentList)
			}
			item := NewNode(NodeListItem)
			item.Children = parseInline(matches[3])
			currentList.AddChild(item)
			continue
		} else {
			if trimmed != "" {
				currentList = nil
			}
		}

		// 3. Handle Headers
		if matches := headerBlockRe.FindStringSubmatch(line); matches != nil {
			level := len(matches[1])
			content := matches[2]

			style := Style{Bold: true}
			if level == 1 {
				style.Reverse = true
			} else if level == 2 {
				style.Underline = true
			}

			node := NewNode(NodeHeader)
			node.Style = style
			node.Children = parseInline(content)
			root.AddChild(node)
			continue
		}

		// 4. Handle Horizontal Rules
		if hrBlockRe.MatchString(trimmed) {
			root.AddChild(NewNode(NodeHR))
			continue
		}

		// 5. Handle Blockquotes
		if matches := quoteBlockRe.FindStringSubmatch(line); matches != nil {
			node := NewNode(NodeQuote)
			node.Children = parseInline(matches[1])
			root.AddChild(node)
			continue
		}

		// 6. Default: Paragraph / Text Block
		if trimmed == "" {
			root.AddChild(NewNode(NodeText))
			continue
		}

		node := NewNode(NodeBlock)
		node.Children = parseInline(line)
		root.AddChild(node)
	}

	return root
}

// parseInline parses inline styles, colors, and holes
func parseInline(text string) []*Node {
	var nodes []*Node

	lastIndex := 0
	matches := inlineTokenRe.FindAllStringIndex(text, -1)

	for _, match := range matches {
		start, end := match[0], match[1]

		// Add preceding text
		if start > lastIndex {
			nodes = append(nodes, &Node{
				Type:    NodeText,
				Content: text[lastIndex:start],
			})
		}

		token := text[start:end]

		if token == "%v" {
			nodes = append(nodes, &Node{
				Type:   NodeHole,
				HoleID: -1,
			})
		} else if strings.HasPrefix(token, "**") {
			// Bold
			content := token[2 : len(token)-2]
			styleNode := NewNode(NodeStyle)
			styleNode.Style = Style{Bold: true}
			styleNode.Children = parseInline(content)
			nodes = append(nodes, styleNode)
		} else if strings.HasPrefix(token, "*") {
			// Italic
			content := token[1 : len(token)-1]
			styleNode := NewNode(NodeStyle)
			styleNode.Style = Style{Italic: true}
			styleNode.Children = parseInline(content)
			nodes = append(nodes, styleNode)
		} else if strings.HasPrefix(token, "__") {
			// Underline
			content := token[2 : len(token)-2]
			styleNode := NewNode(NodeStyle)
			styleNode.Style = Style{Underline: true}
			styleNode.Children = parseInline(content)
			nodes = append(nodes, styleNode)
		} else if strings.HasPrefix(token, "~~") {
			// Strikethrough
			content := token[2 : len(token)-2]
			styleNode := NewNode(NodeStyle)
			styleNode.Style = Style{Strike: true}
			styleNode.Children = parseInline(content)
			nodes = append(nodes, styleNode)
		} else if strings.Contains(token, "#") {
			// Color
			isBg := strings.HasPrefix(token, "!")
			startParen := strings.Index(token, "(")
			endParen := strings.LastIndex(token, ")")

			if startParen > -1 && endParen > startParen {
				colorName := token[1:startParen]
				if isBg {
					colorName = token[2:startParen]
				}
				content := token[startParen+1 : endParen]

				styleNode := NewNode(NodeStyle)
				ansiColor := GetColorCode(colorName)

				if isBg {
					styleNode.Style = Style{BgColor: ansiColor}
				} else {
					styleNode.Style = Style{Color: ansiColor}
				}

				styleNode.Children = parseInline(content)
				nodes = append(nodes, styleNode)
			} else {
				nodes = append(nodes, &Node{Type: NodeText, Content: token})
			}
		}

		lastIndex = end
	}

	if lastIndex < len(text) {
		nodes = append(nodes, &Node{
			Type:    NodeText,
			Content: text[lastIndex:],
		})
	}

	return nodes
}
