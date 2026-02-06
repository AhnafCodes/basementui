package basement

import (
	"testing"
)

func TestParseAST(t *testing.T) {
	input := "# Hello **World** %v"
	root := ParseAST(input)

	if root.Type != NodeRoot {
		t.Errorf("Expected root node")
	}

	if len(root.Children) != 1 {
		t.Errorf("Expected 1 block (line), got %d", len(root.Children))
	}

	block := root.Children[0]
	if block.Type != NodeHeader {
		t.Errorf("Expected header node, got %d", block.Type)
	}
	if !block.Style.Reverse {
		t.Errorf("Expected header to be reversed (level 1)")
	}

	// Children of the block: "Hello ", "**World**", " ", "%v"
	// Actually, "Hello " is text, "**World**" is style, " " is text, "%v" is hole
	// Let's check the structure

	// 1. Text "Hello " (Note: parseBlock trims space after #, so it might be just "Hello")
	// Wait, parseBlock: content := strings.TrimSpace(line[level:]) -> "Hello **World** %v"
	// Then parseInline parses that.

	// Expected:
	// 1. Text "Hello "
	// 2. Style (Bold) -> Text "World"
	// 3. Text " "
	// 4. Hole

	children := block.Children
	if len(children) != 4 {
		t.Errorf("Expected 4 inline nodes, got %d", len(children))
	}

	if children[0].Type != NodeText || children[0].Content != "Hello " {
		t.Errorf("Node 1 mismatch: %+v", children[0])
	}

	if children[1].Type != NodeStyle || !children[1].Style.Bold {
		t.Errorf("Node 2 mismatch: %+v", children[1])
	}

	if children[3].Type != NodeHole {
		t.Errorf("Node 4 mismatch: %+v", children[3])
	}
}
