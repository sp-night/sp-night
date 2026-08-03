package render

import (
	"fmt"
	"strings"
	"text/template/parse"
)

// Linting the one rule the whole project rests on.
//
// A template must never reference a palette colour directly:
//
//	✗  palette = 4={{ .C.marginal }}
//	✓  palette = 4={{ .R.ansi.blue }}
//
// Why: moving syntax.keyword from marginal to temporal has to repaint Neovim,
// bat, fish and the website at once. A template with marginal written into it
// stays behind and nobody notices — which is exactly how a family of ports
// drifts apart over a couple of years.
//
// Until now this was prose in the style guide. Here it is a gate.
//
// The exception is the variable lists: waybar.css, gtk.css and hyprland.conf
// publish the raw palette as @define-color / $var because the end user will want
// to write @sp_sodio in their own stylesheet. Those templates declare
// raw_palette: true in their frontmatter, and even then everything the template
// itself styles must still use a role.
//
// The check walks the parse tree rather than the text. A textual search cannot
// tell a real reference from the counter-example in a comment — the scaffolded
// stub shows the wrong form on purpose, and a grep flags its own documentation.

// Finding is one lint result.
type Finding struct {
	Line    int
	Column  int
	Excerpt string
	Message string
}

func (f Finding) String() string {
	return fmt.Sprintf("%d:%d: %s\n      %s", f.Line, f.Column, f.Message, f.Excerpt)
}

const rawPaletteMessage = "references the raw palette (.C) instead of a role (.R). " +
	"Ask for a role, not a colour — or set raw_palette: true if this template " +
	"publishes the palette as variables for the user's own stylesheet."

// Lint checks a parsed template against the roles-only rule.
func (t *Template) Lint() []Finding {
	if t.Spec.RawPalette || t.tpl == nil || t.tpl.Tree == nil {
		return nil
	}

	var out []Finding
	walk(t.tpl.Tree.Root, func(n parse.Node) {
		if !referencesRawPalette(n) {
			return
		}
		line, col, excerpt := t.locate(int(n.Position()))
		out = append(out, Finding{Line: line, Column: col, Excerpt: excerpt, Message: rawPaletteMessage})
	})
	return out
}

// referencesRawPalette reports whether a node reaches for .C.
func referencesRawPalette(n parse.Node) bool {
	switch v := n.(type) {
	case *parse.FieldNode:
		// {{ .C.sodio }} and {{ index .C "sodio" }} both arrive here.
		return len(v.Ident) > 0 && v.Ident[0] == "C"
	case *parse.VariableNode:
		// {{ index $.C . }} — the first ident is the variable itself.
		return len(v.Ident) > 1 && v.Ident[1] == "C"
	default:
		return false
	}
}

// walk visits every node of a template's parse tree.
func walk(n parse.Node, visit func(parse.Node)) {
	if n == nil {
		return
	}
	visit(n)

	switch v := n.(type) {
	case *parse.ListNode:
		if v == nil {
			return
		}
		for _, c := range v.Nodes {
			walk(c, visit)
		}
	case *parse.ActionNode:
		walk(v.Pipe, visit)
	case *parse.PipeNode:
		if v == nil {
			return
		}
		for _, d := range v.Decl {
			walk(d, visit)
		}
		for _, c := range v.Cmds {
			walk(c, visit)
		}
	case *parse.CommandNode:
		for _, a := range v.Args {
			walk(a, visit)
		}
	case *parse.IfNode:
		walkBranch(&v.BranchNode, visit)
	case *parse.RangeNode:
		walkBranch(&v.BranchNode, visit)
	case *parse.WithNode:
		walkBranch(&v.BranchNode, visit)
	case *parse.TemplateNode:
		walk(v.Pipe, visit)
	case *parse.ChainNode:
		walk(v.Node, visit)
	}
}

func walkBranch(b *parse.BranchNode, visit func(parse.Node)) {
	walk(b.Pipe, visit)
	if b.List != nil {
		walk(b.List, visit)
	}
	if b.ElseList != nil {
		walk(b.ElseList, visit)
	}
}

// locate turns a byte offset in the body into a line, a column and the text of
// that line.
func (t *Template) locate(offset int) (line, col int, excerpt string) {
	if offset < 0 || offset > len(t.Body) {
		return 0, 0, ""
	}
	before := t.Body[:offset]
	line = strings.Count(before, "\n") + 1
	lineStart := strings.LastIndex(before, "\n") + 1
	col = offset - lineStart + 1

	rest := t.Body[lineStart:]
	if i := strings.Index(rest, "\n"); i >= 0 {
		rest = rest[:i]
	}
	return line, col, strings.TrimSpace(rest)
}
