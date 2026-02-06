package main

import (
	"basement/signals"
	"basement/tui"
)

func main() {
	// Example 11: Markdown Rendering
	// Demonstrates rendering of a complex Markdown document.
	// Note: BasementUI supports a subset of Markdown syntax.

	markdown := `
# BasementUI Markdown Demo

---
__Advertisement :)__

- __[pica](https://nodeca.github.io/pica/demo/)__ - high quality and fast image
  resize in browser.
- __[babelfish](https://github.com/nodeca/babelfish/)__ - developer friendly
  i18n with plurals support and easy syntax.

You will like those projects!

---

# h1 Heading 8-)
## h2 Heading
### h3 Heading
#### h4 Heading
##### h5 Heading
###### h6 Heading


## Horizontal Rules

___

---

***


## Typographic replacements

Enable typographer option to see result.

(c) (C) (r) (R) (tm) (TM) (p) (P) +-

test.. test... test..... test?..... test!....

!!!!!! ???? ,,  -- ---

"Smartypants, double quotes" and 'single quotes'


## Emphasis

**This is bold text**

__This is bold text__ (Rendered as Underline in BasementUI)

*This is italic text* (Not supported yet)

_This is italic text_ (Not supported yet)

~~Strikethrough~~ (Not supported yet)


## Blockquotes


> Blockquotes can also be nested...
>> ...by using additional greater-than signs right next to each other...
> > > ...or with spaces between arrows.


## Lists

Unordered

+ Create a list by starting a line with ` + "`+`" + `, ` + "`-`" + `, or ` + "`*`" + `
+ Sub-lists are made by indenting 2 spaces:
  - Marker character change forces new list start:
    * Ac tristique libero volutpat at
    + Facilisis in pretium nisl aliquet
    - Nulla volutpat aliquam velit
+ Very easy!

Ordered

1. Lorem ipsum dolor sit amet
2. Consectetur adipiscing elit
3. Integer molestie lorem at massa


1. You can use sequential numbers...
1. ...or keep all the numbers as ` + "`1.`" + `

Start numbering with offset:

57. foo
1. bar


## Code

Inline ` + "`code`" + `

Indented code

    // Some comments
    line 1 of code
    line 2 of code
    line 3 of code


Block code "fences"

` + "```" + `
Sample text here...
` + "```" + `

Syntax highlighting

` + "``` js" + `
var foo = function (bar) {
  return bar++;
};

console.log(foo(5));
` + "```" + `

## Tables

| Option | Description |
| ------ | ----------- |
| data   | path to data files to supply the data that will be passed into templates. |
| engine | engine to be used for processing templates. Handlebars is the default. |
| ext    | extension to be used for dest files. |

Right aligned columns

| Option | Description |
| ------:| -----------:|
| data   | path to data files to supply the data that will be passed into templates. |
| engine | engine to be used for processing templates. Handlebars is the default. |
| ext    | extension to be used for dest files. |


## Links

[link text](http://dev.nodeca.com)

[link with title](http://nodeca.github.io/pica/demo/ "title text!")

Autoconverted link https://github.com/nodeca/pica (enable linkify to see)


## Images

![Minion](https://octodex.github.com/images/minion.png)
![Stormtroopocat](https://octodex.github.com/images/stormtroopocat.jpg "The Stormtroopocat")

Like links, Images also have a footnote style syntax

![Alt text][id]

With a reference later in the document defining the URL location:

[id]: https://octodex.github.com/images/dojocat.jpg  "The Dojocat"


## Plugins

The killer feature of ` + "`markdown-it`" + ` is very effective support of
[syntax plugins](https://www.npmjs.org/browse/keyword/markdown-it-plugin).


### [Emojies](https://github.com/markdown-it/markdown-it-emoji)

> Classic markup: :wink: :cry: :laughing: :yum:
>
> Shortcuts (emoticons): :-) :-( 8-) ;)

see [how to change output](https://github.com/markdown-it/markdown-it-emoji#change-output) with twemoji.


### [Subscript](https://github.com/markdown-it/markdown-it-sub) / [Superscript](https://github.com/markdown-it/markdown-it-sup)

- 19^th^
- H~2~O


### [\<ins>](https://github.com/markdown-it/markdown-it-ins)

++Inserted text++


### [\<mark>](https://github.com/markdown-it/markdown-it-mark)

==Marked text==


### [Footnotes](https://github.com/markdown-it/markdown-it-footnote)

Footnote 1 link[^first].

Footnote 2 link[^second].

Inline footnote^[Text of inline footnote] definition.

Duplicated footnote reference[^second].

[^first]: Footnote **can have markup**

    and multiple paragraphs.

[^second]: Footnote text.


### [Definition lists](https://github.com/markdown-it/markdown-it-deflist)

Term 1

:   Definition 1
with lazy continuation.

Term 2 with *inline markup*

:   Definition 2

        { some code, part of Definition 2 }

    Third paragraph of definition 2.

_Compact style:_

Term 1
  ~ Definition 1

Term 2
  ~ Definition 2a
  ~ Definition 2b


### [Abbreviations](https://github.com/markdown-it/markdown-it-abbr)

This is HTML abbreviation example.

It converts "HTML", but keep intact partial entries like "xxxHTMLyyy" and so on.

*[HTML]: Hyper Text Markup Language

### [Custom containers](https://github.com/markdown-it/markdown-it-container)

::: warning
*here be dragons*
:::

(Press 'q' or Ctrl+C to exit. Use Up/Down to scroll.)
`

	// Reactive scroll state
	scrollY := signals.New(0)

	app := func() tui.Renderable {
		// We can't easily pass scrollY to Template directly as a prop for the whole view
		// because Template parses the string.
		// But Render() now uses screen.ScrollY.
		// So we just need to update screen.ScrollY when the signal changes.
		// Wait, Render() reads screen.ScrollY directly.
		// But screen.ScrollY is not a signal.
		// We need to bind the signal to the screen property or update it manually.

		return tui.Template(markdown)
	}

	screen := tui.NewScreen()
	defer screen.Close()

	// Create an effect to sync signal to screen
	signals.CreateEffect(func() {
		screen.ScrollY = scrollY.Get()
		// Note: This effect is necessary to update the screen state reactively.
		// The render effect (created by tui.Render) depends on signals accessed within it.
		// Since wrappedApp accesses scrollY.Get(), the render effect will re-run when scrollY changes.
		// However, we also need to ensure screen.ScrollY is updated *before* the render logic runs.
		// Because effects run in creation order, and this effect is created before tui.Render,
		// this runs first, updates the screen state, and then the render effect runs with the new state.
	})

	// To make it reactive, we need to access scrollY inside the Render loop.
	// We can wrap the app function.

	wrappedApp := func() tui.Renderable {
		scrollY.Get() // Register dependency so render effect re-runs on scroll
		return app()
	}

	tui.Render(screen, wrappedApp)

	// Handle Input
	quit := make(chan bool)
	screen.OnKey(func(ev tui.KeyEvent) {
		if ev.Rune == 'q' || (ev.Key == tui.KeyChar && ev.Mod == tui.ModCtrl && ev.Rune == 'c') {
			quit <- true
		}

		if ev.Key == tui.KeyArrowDown {
			scrollY.Set(scrollY.Get() + 1)
		} else if ev.Key == tui.KeyArrowUp {
			val := scrollY.Get()
			if val > 0 {
				scrollY.Set(val - 1)
			}
		}
	})
	<-quit
}
