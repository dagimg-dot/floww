package config

import (
	"fmt"
	"strings"

	"github.com/dagimg-dot/floww/internal/diagnostic"
	"github.com/pelletier/go-toml/v2/unstable"
	"gopkg.in/yaml.v3"
)

// Positions maps schema paths (e.g. "workspaces[0].apps[1].exec", built with
// diagnostic.Path) to source positions. The path "" holds the document start.
type Positions struct {
	doc  diagnostic.Position
	path map[string]diagnostic.Position
}

// Position returns the source position for a schema path.
func (p *Positions) Position(path string) (diagnostic.Position, bool) {
	if p == nil {
		return diagnostic.Position{}, false
	}
	pos, ok := p.path[path]
	return pos, ok
}

// buildPositionsFromYAML builds a position index from a YAML (or JSON — YAML
// 1.2 is a superset) document node tree. Duplicate keys resolve last-wins to
// match decoder behavior; alias and merge-key values point at the anchor
// definition, which is where a fix would go.
func buildPositionsFromYAML(root *yaml.Node) *Positions {
	p := &Positions{
		doc:  diagnostic.Position{Line: 1, Column: 1},
		path: make(map[string]diagnostic.Position),
	}
	p.path[""] = p.doc
	if root == nil || len(root.Content) == 0 {
		return p
	}
	doc := root.Content[0]
	p.doc = diagnostic.Position{Line: doc.Line, Column: doc.Column}
	p.path[""] = p.doc
	walkYAMLMapping(p.path, doc, "", make(map[*yaml.Node]bool))
	return p
}

func walkYAMLMapping(m map[string]diagnostic.Position, n *yaml.Node, base string, active map[*yaml.Node]bool) {
	if n == nil || n.Kind != yaml.MappingNode || active[n] {
		return
	}
	active[n] = true
	defer delete(active, n)

	for i := 0; i+1 < len(n.Content); i += 2 {
		key, val := n.Content[i], n.Content[i+1]
		if key.Tag == "!!merge" {
			walkYAMLMapping(m, aliasTarget(val), base, active)
			continue
		}
		if key.Kind != yaml.ScalarNode {
			continue
		}
		path := diagnostic.Path(base, key.Value)
		switch val.Kind {
		case yaml.ScalarNode:
			m[path] = yamlPos(val)
		case yaml.AliasNode:
			m[path] = yamlPos(aliasTarget(val))
		case yaml.MappingNode:
			walkYAMLMapping(m, val, path, active)
		case yaml.SequenceNode:
			walkYAMLSequence(m, val, path, active)
		}
	}
}

func walkYAMLSequence(m map[string]diagnostic.Position, n *yaml.Node, base string, active map[*yaml.Node]bool) {
	for i, item := range n.Content {
		path := diagnostic.Path(base, i)
		m[path] = yamlPos(item)
		switch item.Kind {
		case yaml.MappingNode:
			walkYAMLMapping(m, item, path, active)
		case yaml.AliasNode:
			walkYAMLMapping(m, aliasTarget(item), path, active)
		}
	}
}

func aliasTarget(n *yaml.Node) *yaml.Node {
	if n == nil || n.Alias == nil {
		return nil
	}
	return n.Alias
}

func yamlPos(n *yaml.Node) diagnostic.Position {
	if n == nil {
		return diagnostic.Position{}
	}
	return diagnostic.Position{
		Line:   n.Line,
		Column: n.Column,
		Length: len([]rune(n.Value)),
	}
}

// buildPositionsFromTOML builds a position index from a TOML document using
// go-toml's unstable parser. Table headers are resolved against previously
// seen array tables (a dotted header like [[workspaces.apps]] refers to the
// last element of [[workspaces]]) so paths match the decoder's view.
func buildPositionsFromTOML(data []byte) *Positions {
	p := &Positions{
		doc:  diagnostic.Position{Line: 1, Column: 1},
		path: make(map[string]diagnostic.Position),
	}
	p.path[""] = p.doc

	parser := unstable.Parser{}
	parser.Reset(data)
	lastIdx := make(map[string]int)
	var context []string

	for parser.NextExpression() {
		expr := parser.Expression()
		switch expr.Kind {
		case unstable.Comment:
			continue
		case unstable.Table, unstable.ArrayTable:
			parts := tomlKeyParts(expr.Key())
			if len(parts) == 0 {
				continue
			}
			parent := []string{}
			for _, part := range parts[:len(parts)-1] {
				parent = append(parent, part)
				if idx, ok := lastIdx[tomlJoin(parent)]; ok {
					parent = append(parent, fmt.Sprintf("[%d]", idx))
				}
			}
			tablePath := make([]string, 0, len(parent)+1)
			tablePath = append(tablePath, parent...)
			tablePath = append(tablePath, parts[len(parts)-1])
			tableKey := tomlJoin(tablePath)
			if expr.Kind == unstable.ArrayTable {
				idx := 0
				if v, ok := lastIdx[tableKey]; ok {
					idx = v + 1
				}
				lastIdx[tableKey] = idx
				itemPath := append(append([]string{}, tablePath...), fmt.Sprintf("[%d]", idx))
				p.path[tomlJoin(itemPath)] = tomlPos(&parser, tomlFirstKey(expr.Key()))
				context = itemPath
			} else {
				context = tablePath
			}
		case unstable.KeyValue:
			parts := tomlKeyParts(expr.Key())
			if len(parts) == 0 {
				continue
			}
			val := expr.Value()
			full := append(append([]string{}, context...), parts...)
			valPos := tomlPos(&parser, val)
			if val.Kind == unstable.Array {
				// array node Raw positions are unreliable (parser quirk);
				// anchor at the first element instead
				if el := val.Child(); el != nil && el.Valid() {
					valPos = tomlPos(&parser, el)
				} else {
					continue
				}
			}
			p.path[tomlJoin(full)] = valPos
			if val.Kind == unstable.Array {
				idx := 0
				for it := val.Children(); it.Next(); idx++ {
					el := it.Node()
					elemPath := append(append([]string{}, full...), fmt.Sprintf("[%d]", idx))
					p.path[tomlJoin(elemPath)] = tomlPos(&parser, el)
					if el.Kind == unstable.InlineTable {
						for cit := el.Children(); cit.Next(); {
							kv := cit.Node()
							kparts := tomlKeyParts(kv.Key())
							if len(kparts) == 0 {
								continue
							}
							fullKV := append(append([]string{}, elemPath...), kparts...)
							p.path[tomlJoin(fullKV)] = tomlPos(&parser, kv.Value())
						}
					}
				}
			}
		}
	}
	return p
}

func tomlKeyParts(it unstable.Iterator) []string {
	var parts []string
	for it.Next() {
		parts = append(parts, string(it.Node().Data))
	}
	return parts
}

// tomlJoin joins path segments, attaching index segments ("[0]") to the
// preceding segment without a dot, matching diagnostic.Path's grammar
// ("workspaces[0].apps[1]").
func tomlJoin(segs []string) string {
	var sb strings.Builder
	for i, s := range segs {
		if i > 0 && !strings.HasPrefix(s, "[") {
			sb.WriteByte('.')
		}
		sb.WriteString(s)
	}
	return sb.String()
}

func tomlFirstKey(it unstable.Iterator) *unstable.Node {
	if it.Next() {
		return it.Node()
	}
	return nil
}

func tomlPos(parser *unstable.Parser, n *unstable.Node) diagnostic.Position {
	if n == nil {
		return diagnostic.Position{}
	}
	shape := parser.Shape(n.Raw)
	pos := diagnostic.Position{Line: shape.Start.Line, Column: shape.Start.Column}
	if shape.Start.Line == shape.End.Line {
		pos.Length = shape.End.Column - shape.Start.Column
	} else {
		pos.Length = 1
	}
	return pos
}
