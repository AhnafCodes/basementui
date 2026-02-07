package tui

// Row creates a horizontal layout node
func Row(children ...interface{}) *LayoutNode {
	n := &LayoutNode{
		Direction: DirRow,
		Width:     Auto(),
		Height:    Auto(),
	}
	for _, child := range children {
		n.addChild(wrapChild(child))
	}
	return n
}

// Col creates a vertical layout node
func Col(children ...interface{}) *LayoutNode {
	n := &LayoutNode{
		Direction: DirColumn,
		Width:     Auto(),
		Height:    Auto(),
	}
	for _, child := range children {
		n.addChild(wrapChild(child))
	}
	return n
}

// Box wraps a child with optional border and padding
func Box(child interface{}, border bool, padding int) *LayoutNode {
	n := &LayoutNode{
		Direction: DirColumn,
		Width:     Auto(),
		Height:    Auto(),
		Padding:   padding,
		Border:    border,
	}
	n.addChild(wrapChild(child))
	return n
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

// addChild links a child node into this node's doubly linked child list. O(1).
func (n *LayoutNode) addChild(child *LayoutNode) {
	child.Parent = n
	child.Next = nil
	child.Prev = n.LastChild
	if n.LastChild != nil {
		n.LastChild.Next = child
	} else {
		n.FirstChild = child
	}
	n.LastChild = child
}

// wrapChild ensures a value is represented as a *LayoutNode.
// If v is already a *LayoutNode, it is returned directly.
// Otherwise, it is wrapped in a leaf LayoutNode with Content set.
func wrapChild(v interface{}) *LayoutNode {
	if node, ok := v.(*LayoutNode); ok {
		return node
	}
	return &LayoutNode{
		Width:   Auto(),
		Height:  Auto(),
		Content: v,
	}
}