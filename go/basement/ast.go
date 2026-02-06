package basement

// NodeType identifies the type of a node in the AST
type NodeType int

const (
	NodeRoot NodeType = iota
	NodeText
	NodeStyle
	NodeHole      // Represents a %v placeholder
	NodeBlock     // Generic block element (paragraph)
	NodeHeader    // Header element (#)
	NodeList      // List container
	NodeListItem  // List item
	NodeCodeBlock // Code block (```)
	NodeHR        // Horizontal Rule (---)
	NodeQuote     // Blockquote (>)
)

// Node represents a node in the AST
type Node struct {
	Type     NodeType
	Content  string      // For text nodes or code blocks
	Lang     string      // For code blocks (language identifier)
	Style    Style       // For styled nodes
	Children []*Node     // For nested nodes
	HoleID   int         // Index of the argument for this hole (0-based)
}

// NewNode creates a new node
func NewNode(typ NodeType) *Node {
	return &Node{
		Type:     typ,
		Children: make([]*Node, 0),
	}
}

// AddChild adds a child node
func (n *Node) AddChild(child *Node) {
	n.Children = append(n.Children, child)
}
