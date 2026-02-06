package tui

import (
	"basement/basement"
	"basement/signals"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Measure calculates the dimensions of the layout tree.
// It populates the computed fields in LayoutNode.
func (n *LayoutNode) Measure(constraintW, constraintH int) (int, int) {
	// 1. Determine available space for content (Box Model: Border-Box)
	// Padding and Border consume space from the constraint.

	horizontalDeduction := n.Padding * 2
	verticalDeduction := n.Padding * 2
	if n.Border {
		horizontalDeduction += 2
		verticalDeduction += 2
	}

	contentConstraintW := constraintW - horizontalDeduction
	contentConstraintH := constraintH - verticalDeduction

	if contentConstraintW < 0 { contentConstraintW = 0 }
	if contentConstraintH < 0 { contentConstraintH = 0 }

	// 2. Measure Children based on Direction

	var totalFixed int
	var totalAuto int
	var totalFlexWeight int

	// Initialize childGeoms storage
	n.childGeoms = make([]struct{ w, h int }, len(n.Children))

	// First pass: Measure Fixed and Auto children to determine remaining space for Flex
	for i, child := range n.Children {
		// Resolve signal if present
		val := resolveValue(child)

		// Determine size constraints for this child
		if node, ok := val.(*LayoutNode); ok {
			// It's a nested layout node
			if n.Direction == DirRow {
				if node.Width.Type == SizeFixed {
					w, h := node.Measure(node.Width.Value, contentConstraintH)
					n.childGeoms[i] = struct{ w, h int }{w, h}
					totalFixed += w
				} else if node.Width.Type == SizeAuto {
					w, h := node.Measure(contentConstraintW, contentConstraintH)
					n.childGeoms[i] = struct{ w, h int }{w, h}
					totalAuto += w
				} else { // Flex
					totalFlexWeight += node.Width.Value
				}
			} else { // Column
				if node.Height.Type == SizeFixed {
					w, h := node.Measure(contentConstraintW, node.Height.Value)
					n.childGeoms[i] = struct{ w, h int }{w, h}
					totalFixed += h
				} else if node.Height.Type == SizeAuto {
					w, h := node.Measure(contentConstraintW, contentConstraintH)
					n.childGeoms[i] = struct{ w, h int }{w, h}
					totalAuto += h
				} else { // Flex
					totalFlexWeight += node.Height.Value
				}
			}
		} else {
			// It's content (string, Renderable, etc.)
			w, h := measureContent(val, contentConstraintW, contentConstraintH)
			n.childGeoms[i] = struct{ w, h int }{w, h}

			if n.Direction == DirRow {
				totalAuto += w
			} else {
				totalAuto += h
			}
		}
	}

	// 3. Calculate Flex Space
	var availableSpace int
	if n.Direction == DirRow {
		availableSpace = contentConstraintW - totalFixed - totalAuto
	} else {
		availableSpace = contentConstraintH - totalFixed - totalAuto
	}
	if availableSpace < 0 { availableSpace = 0 }

	// 4. Second pass: Measure Flex children
	var maxCross int // Max height in Row, Max width in Col

	for i, child := range n.Children {
		val := resolveValue(child)

		if node, ok := val.(*LayoutNode); ok {
			isFlex := (n.Direction == DirRow && node.Width.Type == SizeFlex) ||
			          (n.Direction == DirColumn && node.Height.Type == SizeFlex)

			if isFlex {
				weight := 0
				if n.Direction == DirRow { weight = node.Width.Value } else { weight = node.Height.Value }

				share := 0
				if totalFlexWeight > 0 {
					share = (availableSpace * weight) / totalFlexWeight
				}

				var w, h int
				if n.Direction == DirRow {
					w, h = node.Measure(share, contentConstraintH)
				} else {
					w, h = node.Measure(contentConstraintW, share)
				}
				n.childGeoms[i] = struct{ w, h int }{w, h}
			}
		}

		// Update max cross dimension
		if n.Direction == DirRow {
			if n.childGeoms[i].h > maxCross { maxCross = n.childGeoms[i].h }
		} else {
			if n.childGeoms[i].w > maxCross { maxCross = n.childGeoms[i].w }
		}
	}

	// 5. Set Computed Dimensions
	finalW := constraintW
	finalH := constraintH

	if n.Width.Type == SizeAuto {
		if n.Direction == DirRow {
			contentW := 0
			for _, s := range n.childGeoms { contentW += s.w }
			finalW = contentW + horizontalDeduction
		} else {
			finalW = maxCross + horizontalDeduction
		}
	}

	if n.Height.Type == SizeAuto {
		if n.Direction == DirRow {
			finalH = maxCross + verticalDeduction
		} else {
			contentH := 0
			for _, s := range n.childGeoms { contentH += s.h }
			finalH = contentH + verticalDeduction
		}
	}

	n.computedW = finalW
	n.computedH = finalH

	return finalW, finalH
}

// Draw renders the layout tree to the screen
func (n *LayoutNode) Draw(screen *Screen, x, y int) {
	n.computedX = x
	n.computedY = y

	// Draw Border
	if n.Border {
		drawBorder(screen, x, y, n.computedW, n.computedH)
	}

	// Content area start
	contentX := x + n.Padding
	contentY := y + n.Padding
	if n.Border {
		contentX++
		contentY++
	}

	// Draw Children
	curX, curY := contentX, contentY

	for i, child := range n.Children {
		val := resolveValue(child)
		size := n.childGeoms[i]

		if node, ok := val.(*LayoutNode); ok {
			node.Draw(screen, curX, curY)
		} else {
			// Render content
			drawContent(screen, val, curX, curY, size.w, size.h)
		}

		// Advance cursor
		if n.Direction == DirRow {
			curX += size.w
		} else {
			curY += size.h
		}
	}
}

func resolveValue(v interface{}) interface{} {
	if s, ok := v.(signals.Getter); ok {
		return s.GetValue()
	}
	return v
}

func measureContent(v interface{}, maxW, maxH int) (int, int) {
	s := fmt.Sprintf("%v", v)

	// Handle newlines for correct measurement
	lines := strings.Split(s, "\n")

	maxLineLen := 0
	for _, line := range lines {
		l := utf8.RuneCountInString(line)
		if l > maxLineLen {
			maxLineLen = l
		}
	}

	w := maxLineLen
	h := len(lines)

	if w > maxW { w = maxW }
	if h > maxH { h = maxH }

	return w, h
}

func drawContent(screen *Screen, v interface{}, x, y, w, h int) {
	s := fmt.Sprintf("%v", v)

	// Handle newlines
	lines := strings.Split(s, "\n")

	for i, line := range lines {
		if i >= h {
			break
		}

		// Truncate line if too long
		if utf8.RuneCountInString(line) > w {
			runes := []rune(line)
			line = string(runes[:w])
		}

		// Draw directly to buffer to avoid mutex issues if called from Render loop
		// But Screen.DrawText uses mutex.
		// Since we are single-threaded in Render effect, mutex is fine but redundant.
		// However, to be consistent with drawBorder (which uses Set directly),
		// we should probably use Set directly here too.

		col := x
		for _, r := range line {
			screen.Back.Set(col, y+i, r, basement.Style{})
			col++
		}
	}
}

func drawBorder(screen *Screen, x, y, w, h int) {
	// Unicode box drawing
	// ┌─┐
	// │ │
	// └─┘

	style := basement.Style{} // Default style

	// Corners
	screen.Back.Set(x, y, '┌', style)
	screen.Back.Set(x+w-1, y, '┐', style)
	screen.Back.Set(x, y+h-1, '└', style)
	screen.Back.Set(x+w-1, y+h-1, '┘', style)

	// Top/Bottom
	for i := 1; i < w-1; i++ {
		screen.Back.Set(x+i, y, '─', style)
		screen.Back.Set(x+i, y+h-1, '─', style)
	}

	// Left/Right
	for i := 1; i < h-1; i++ {
		screen.Back.Set(x, y+i, '│', style)
		screen.Back.Set(x+w-1, y+i, '│', style)
	}
}
