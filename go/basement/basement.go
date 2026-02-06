package basement

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
)

var (
	codeBlockRe   = regexp.MustCompile("`+")
	horizontalRe  = regexp.MustCompile("(?m)^[ ]{0,2}([ ]?[*_-][ ]?){3,}[ \t]*$")
	headerRe      = regexp.MustCompile("(?m)^(\\#{1,6})[ \\t]+(.+?)[ \\t]*\\#*([\r\n]+|$)")
	listRe        = regexp.MustCompile("(?m)^([ \\t]{1,})[*+-]([ \\t]{1,})")
	quoteRe       = regexp.MustCompile("(?m)^[ \\t]*>([ \\t]?)")
	colorRe       = regexp.MustCompile("(?s)(!?)#([a-zA-Z0-9]{3,8})\\((.+?)\\)([^)]|$)")

	// Precomputed regexes for boldUnderlineStrike
	styleRegexes []*regexp.Regexp
	styleReplacements []struct{ start, end string }
)

func init() {
	styles := []struct {
		char      string
		codeStart string
		codeEnd   string
	}{
		{"\\*", "\x1b[1m", "\x1b[22m"}, // Bold
		{"-", "\x1b[2m", "\x1b[22m"},   // Dim
		{"_", "\x1b[4m", "\x1b[24m"},   // Underline
		{":", "\x1b[5m", "\x1b[25m"},   // Blink
		{"!", "\x1b[7m", "\x1b[27m"},   // Reverse
		{"\\?", "\x1b[8m", "\x1b[28m"}, // Hidden
		{"~", "\x1b[9m", "\x1b[29m"},   // Strike
	}

	for _, style := range styles {
		c := style.char
		pattern := fmt.Sprintf(
			"(?s)(%s%s)(\\S|\\S.*?\\S)%s%s|(%s)(\\S|\\S.*?\\S)%s",
			c, c, c, c, c, c,
		)
		styleRegexes = append(styleRegexes, regexp.MustCompile(pattern))
		styleReplacements = append(styleReplacements, struct{ start, end string }{style.codeStart, style.codeEnd})
	}
}

// Parse takes a basement formatted string and returns the ANSI escaped string
func Parse(txt string) string {
	// Local map to ensure thread safety
	codeMap := make(map[string]string)

	// Preserve code blocks
	txt = processCodeBlocks(txt, codeMap)

	txt = horizontal(txt)
	txt = header(txt)
	txt = boldUnderlineStrike(txt)
	txt = list(txt)
	txt = quote(txt)
	txt = color(txt)

	// Restore code blocks
	for hash, content := range codeMap {
		txt = strings.ReplaceAll(txt, hash, content)
	}

	return txt
}

type replacement struct {
	start, end int
	text       string
}

func processCodeBlocks(txt string, codeMap map[string]string) string {
	indices := codeBlockRe.FindAllStringIndex(txt, -1)

	if len(indices) == 0 {
		return txt
	}

	var replacements []replacement
	used := make(map[int]bool)

	for i := 0; i < len(indices); i++ {
		if used[i] {
			continue
		}

		len1 := indices[i][1] - indices[i][0]

		// Find closer
		found := -1
		for j := i + 1; j < len(indices); j++ {
			if used[j] {
				continue
			}
			len2 := indices[j][1] - indices[j][0]
			if len1 == len2 {
				found = j
				break
			}
		}

		if found != -1 {
			start := indices[i][1] // end of opener
			end := indices[found][0] // start of closer

			content := txt[start:end]
			hash := md5Base64(content)
			codeMap[hash] = content

			replacements = append(replacements, replacement{
				start: start,
				end:   end,
				text:  hash,
			})

			// Mark used
			used[i] = true
			used[found] = true
			// Mark everything in between as used (skipped)
			for k := i + 1; k < found; k++ {
				used[k] = true
			}
		}
	}

	if len(replacements) == 0 {
		return txt
	}

	var sb strings.Builder
	last := 0
	for _, r := range replacements {
		sb.WriteString(txt[last:r.start])
		sb.WriteString(r.text)
		last = r.end
	}
	sb.WriteString(txt[last:])

	return sb.String()
}

func md5Base64(text string) string {
	hash := md5.Sum([]byte(text))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func horizontal(txt string) string {
	line := strings.Repeat("─", 72)
	return horizontalRe.ReplaceAllString(txt, "\x1b[1m"+line+"\x1b[22m")
}

func header(txt string) string {
	return headerRe.ReplaceAllStringFunc(txt, func(match string) string {
		parts := headerRe.FindStringSubmatch(match)
		hash := parts[1]
		content := parts[2]
		suffix := parts[3]

		if len(hash) == 1 {
			content = "\x1b[1m" + content + "\x1b[22m"
		}
		return "\x1b[7m " + content + " \x1b[27m" + suffix
	})
}

func boldUnderlineStrike(txt string) string {
	for i, re := range styleRegexes {
		style := styleReplacements[i]
		txt = re.ReplaceAllStringFunc(txt, func(m string) string {
			sub := re.FindStringSubmatch(m)
			if sub[1] != "" {
				return style.start + sub[2] + style.end
			}
			return style.start + sub[4] + style.end
		})
	}
	return txt
}

func list(txt string) string {
	return listRe.ReplaceAllString(txt, "$1•$2")
}

func quote(txt string) string {
	return quoteRe.ReplaceAllString(txt, "\x1b[7m$1\x1b[27m$1")
}

func color(txt string) string {
	return colorRe.ReplaceAllStringFunc(txt, func(match string) string {
		parts := colorRe.FindStringSubmatch(match)
		bg := parts[1]
		rgb := parts[2]
		content := parts[3]
		suffix := parts[4]

		return getColor(bg, rgb, content) + suffix
	})
}

func getColor(bg, rgb, txt string) string {
	out := GetColorCode(rgb)

	if out == "" {
		out = txt
	} else {
		out = out + txt + "\x1b[39m"
	}

	if bg == "" {
		return out
	}
	return "\x1b[7m" + out + "\x1b[27m"
}
