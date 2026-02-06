package tui

// Row creates a horizontal layout node
func Row(children ...interface{}) *LayoutNode {
	return &LayoutNode{
		Direction: DirRow,
		Width:     Auto(), // Default to Auto? Or Flex? Usually Rows fill width.
		Height:    Auto(),
		Children:  children,
	}
}

// Col creates a vertical layout node
func Col(children ...interface{}) *LayoutNode {
	return &LayoutNode{
		Direction: DirColumn,
		Width:     Auto(),
		Height:    Auto(),
		Children:  children,
	}
}

// Box wraps a child with optional border and padding
func Box(child interface{}, border bool, padding int) *LayoutNode {
	return &LayoutNode{
		Direction: DirColumn, // Default to Col for single child
		Width:     Auto(),
		Height:    Auto(),
		Padding:   padding,
		Border:    border,
		Children:  []interface{}{child},
	}
}

// WithSize sets the size constraints for a node
func (n *LayoutNode) WithSize(w, h Size) *LayoutNode {
	n.Width = w
	n.Height = h
	return n
}

// WithWidth sets the width constraint
func (n *LayoutNode) WithWidth(w Size) *LayoutNode {
	n.Width = w
	return n
}

// WithHeight sets the height constraint
func (n *LayoutNode) WithHeight(h Size) *LayoutNode {
	n.Height = h
	return n
}
