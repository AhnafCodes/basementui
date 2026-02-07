package tui

import (
	"basement/basement"
	"basement/signals"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Renderable represents a parsed template ready to be rendered
type Renderable struct {
	Root *basement.Node
	Args []interface{}
}

// Template parses the template and binds arguments
func Template(template string, args ...interface{}) Renderable {
	root := basement.ParseAST(template)

	// Assign HoleIDs
	holeCount := 0
	assignHoles(root, &holeCount)

	return Renderable{
		Root: root,
		Args: args,
	}
}

func assignHoles(n *basement.Node, count *int) {
	if n.Type == basement.NodeHole {
		n.HoleID = *count
		*count++
	}
	for _, child := range n.Children {
		assignHoles(child, count)
	}
}

// Render mounts the renderable to the screen
func Render(screen *Screen, fn func() Renderable) {
	// Create an effect for the rendering
	signals.CreateEffect(func() {
		// Execute the view function inside the effect.
		r := fn()

		// Use Frame to lock once for the entire render cycle
		screen.Frame(func() {
			// Render the tree to the Back buffer
			// Note: renderNode will access signal values via GetValue(),
			// which registers this effect as a subscriber.
			// Pass ScrollY as negative offset to y
			renderNode(screen, r.Root, r.Args, 0, -screen.ScrollY)
		})
	})
}

// renderNode draws the node to the screen. Returns the new X, Y position.
func renderNode(s *Screen, n *basement.Node, args []interface{}, x, y int) (int, int) {
	// Early exit if node is completely below the viewport
	// Note: This assumes nodes render downwards.
	// If y >= Height, we can skip.
	// But we need to be careful about returning the correct Y advancement for layout.
	// Since we don't have a layout engine here (except for LayoutNode),
	// skipping might be tricky if we need to calculate height.
	// However, for simple flow, if we are off screen, we can just return y + estimated height?
	// Or just let it run but DrawText will clip.
	// The optimization requested is "Early exit when y >= Height".

	if y >= s.Back.Height {
		// We still need to traverse to find the correct next Y?
		// If we return y, the parent loop continues.
		// If we return y + height, the parent loop continues.
		// If we stop traversing, we might miss side effects? (Unlikely in render)
		// But we might miss calculating the total height if we needed it.
		// Since we don't use the returned Y for anything other than placing the next sibling,
		// and the next sibling will also be off-screen, it's safe to return y.
		// Wait, if we return y, the next sibling draws at y.
		// If we return y+1, next sibling draws at y+1.
		// It doesn't matter much if they are all off screen.
		return x, y
	}

	switch n.Type {
	case basement.NodeRoot:
		curY := y
		for _, child := range n.Children {
			_, newY := renderNode(s, child, args, x, curY)
			curY = newY // Don't add extra line here, blocks handle it
		}
		return x, curY

	case basement.NodeBlock, basement.NodeHeader:
		// Apply block style
		curX := x
		for _, child := range n.Children {
			// Inherit style from block
			mergedStyle := mergeStyles(n.Style, child.Style)

			// Shallow copy to avoid mutating AST
			tempChild := *child
			tempChild.Style = mergedStyle

			newX, _ := renderNode(s, &tempChild, args, curX, y)
			curX = newX
		}
		return x, y + 1 // Blocks imply new line

	case basement.NodeHR:
		// Draw a horizontal line
		if y >= 0 && y < s.Back.Height {
			for i := 0; i < s.Back.Width; i++ {
				s.Back.Set(i, y, '─', basement.Style{Dim: true})
			}
		}
		return x, y + 1

	case basement.NodeQuote:
		// Draw quote bar
		if y >= 0 && y < s.Back.Height {
			s.Back.Set(x, y, '│', basement.Style{Dim: true})
		}
		curX := x + 2 // Indent
		for _, child := range n.Children {
			newX, _ := renderNode(s, child, args, curX, y)
			curX = newX
		}
		return x, y + 1

	case basement.NodeList:
		curY := y
		for _, child := range n.Children {
			_, newY := renderNode(s, child, args, x, curY)
			curY = newY
		}
		return x, curY

	case basement.NodeListItem:
		// Draw bullet
		if y >= 0 && y < s.Back.Height {
			s.Back.Set(x, y, '•', basement.Style{})
		}
		curX := x + 2
		for _, child := range n.Children {
			newX, _ := renderNode(s, child, args, curX, y)
			curX = newX
		}
		return x, y + 1

	case basement.NodeCodeBlock:
		// Use Highlighter
		spans := Highlight(n.Content, n.Lang)

		curY := y
		curX := x

		for _, span := range spans {
			// Handle newlines in span text
			parts := strings.Split(span.Text, "\n")
			for i, part := range parts {
				if i > 0 {
					curY++
					curX = x
				}
				if part == "" { continue }

				if curY >= 0 && curY < s.Back.Height {
					// Use unlocked version since we are inside Frame()
					s.drawTextUnlocked(curX, curY, part, span.Style)
				}
				curX += utf8.RuneCountInString(part)
			}
		}
		return x, curY + 1

	case basement.NodeText:
		// Handle empty text nodes as spacers if content is empty but it's a block context?
		// If content is empty string, DrawText does nothing.
		if n.Content == "" {
			return x, y + 1 // Treat as newline
		}
		if y >= 0 && y < s.Back.Height {
			// Use unlocked version since we are inside Frame()
			s.drawTextUnlocked(x, y, n.Content, n.Style)
		}
		return x + utf8.RuneCountInString(n.Content), y

	case basement.NodeStyle:
		curX := x
		for _, child := range n.Children {
			mergedStyle := mergeStyles(n.Style, child.Style)

			tempChild := *child // Shallow copy
			tempChild.Style = mergedStyle

			newX, _ := renderNode(s, &tempChild, args, curX, y)
			curX = newX
		}
		return curX, y

	case basement.NodeHole:
		if n.HoleID < len(args) {
			val := args[n.HoleID]

			// Resolve signal if present
			if getter, ok := val.(signals.Getter); ok {
				val = getter.GetValue()
			}

			// Check if it's a LayoutNode
			if layoutNode, ok := val.(*LayoutNode); ok {
				constraintW := s.Back.Width - x
				constraintH := s.Back.Height - y
				_, h := layoutNode.Measure(constraintW, constraintH)
				layoutNode.Draw(s, x, y)
				return x, y + h
			}

			str := fmt.Sprintf("%v", val)

			if containsMarkup(str) {
				dynamicRoot := basement.ParseAST(str)
				curX := x
				for _, child := range dynamicRoot.Children {
					if child.Type == basement.NodeBlock {
						for _, inlineChild := range child.Children {
							mergedStyle := mergeStyles(n.Style, inlineChild.Style)
							tempChild := *inlineChild
							tempChild.Style = mergedStyle
							newX, _ := renderNode(s, &tempChild, nil, curX, y)
							curX = newX
						}
					}
				}
				return curX, y
			} else {
				if y >= 0 && y < s.Back.Height {
					// Use unlocked version since we are inside Frame()
					s.drawTextUnlocked(x, y, str, n.Style)
				}
				return x + utf8.RuneCountInString(str), y
			}
		}
	}
	return x, y
}

func containsMarkup(s string) bool {
	for _, char := range []string{"**", "__", "#", "!"} {
		if strings.Contains(s, char) {
			return true
		}
	}
	return false
}

func mergeStyles(parent, child basement.Style) basement.Style {
	color := child.Color
	if color == "" {
		color = parent.Color
	}

	bgColor := child.BgColor
	if bgColor == "" {
		bgColor = parent.BgColor
	}

	return basement.Style{
		Bold:      parent.Bold || child.Bold,
		Dim:       parent.Dim || child.Dim,
		Italic:    parent.Italic || child.Italic,
		Underline: parent.Underline || child.Underline,
		Strike:    parent.Strike || child.Strike,
		Reverse:   parent.Reverse || child.Reverse,
		Blink:     parent.Blink || child.Blink,
		Color:     color,
		BgColor:   bgColor,
	}
}
