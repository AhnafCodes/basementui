package tui

// Direction defines the layout direction
type Direction int

const (
	DirRow Direction = iota
	DirColumn
)

// SizeType defines how a node is sized
type SizeType int

const (
	SizeAuto SizeType = iota // Sized by content
	SizeFixed                // Fixed number of cells
	SizeFlex                 // Proportional to remaining space
)

// Size represents a dimension constraint
type Size struct {
	Type  SizeType
	Value int // For Fixed (cells) or Flex (weight)
}

// Fixed creates a fixed size constraint
func Fixed(n int) Size {
	return Size{Type: SizeFixed, Value: n}
}

// Flex creates a flexible size constraint
func Flex(n int) Size {
	return Size{Type: SizeFlex, Value: n}
}

// Auto creates an auto size constraint
func Auto() Size {
	return Size{Type: SizeAuto}
}

// LayoutNode represents a node in the layout tree.
// Uses a doubly linked list structure (inspired by LinkeDOM) instead of
// child slices for O(1) insertions and zero slice allocations.
type LayoutNode struct {
	Direction Direction
	Width     Size
	Height    Size
	Padding   int
	Border    bool
	Content   interface{} // For leaf nodes: string, Renderable, or Signal

	// Linked list pointers
	Parent     *LayoutNode
	FirstChild *LayoutNode
	LastChild  *LayoutNode
	Prev       *LayoutNode
	Next       *LayoutNode

	// Calculated during Measure pass
	computedX, computedY int
	computedW, computedH int
}