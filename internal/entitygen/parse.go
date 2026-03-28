package entitygen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// Field describes one persisted column.
type Field struct {
	GoName    string
	Column    string
	UpdateKey string // map key for UpdateDoc (bson > json > column)
	SQLType   string
	Kind      TypeKind
	Nullable  bool
}

// TypeKind is a coarse type for SQL and scan codegen.
type TypeKind int

const (
	KindString TypeKind = iota
	KindInt64
	KindBool
)

// ParsedEntity is the metadata extracted from a Go struct.
type ParsedEntity struct {
	TypeName string
	Package  string
	Fields   []Field
}

// ParseFile parses srcPath and returns metadata for struct typeName.
func ParseFile(srcPath, typeName string) (*ParsedEntity, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, srcPath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	var out *ParsedEntity
	ast.Inspect(f, func(n ast.Node) bool {
		if out != nil {
			return false
		}
		ts, ok := n.(*ast.TypeSpec)
		if !ok || ts.Name == nil || ts.Name.Name != typeName {
			return true
		}
		st, ok := ts.Type.(*ast.StructType)
		if !ok {
			return true
		}
		pkg := f.Name.Name
		fields, err2 := parseStructFields(st)
		if err2 != nil {
			err = err2
			return false
		}
		out = &ParsedEntity{TypeName: typeName, Package: pkg, Fields: fields}
		return false
	})
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, fmt.Errorf("entitygen: type %q not found in %s", typeName, srcPath)
	}
	return out, nil
}

func parseStructFields(st *ast.StructType) ([]Field, error) {
	var fields []Field
	if st.Fields == nil {
		return fields, nil
	}
	for _, fld := range st.Fields.List {
		if len(fld.Names) == 0 {
			continue // embedded
		}
		for _, id := range fld.Names {
			if !id.IsExported() {
				continue
			}
			tag := ""
			if fld.Tag != nil {
				tag = strings.Trim(fld.Tag.Value, "`")
			}
			col := columnFromTag(tag)
			if col == "" {
				continue
			}
			kind, nullable, err := goTypeKind(fld.Type)
			if err != nil {
				return nil, fmt.Errorf("field %s: %w", id.Name, err)
			}
			sqlt := sqlType(kind, nullable)
			fields = append(fields, Field{
				GoName:    id.Name,
				Column:    col,
				UpdateKey: updateDocKey(tag, col),
				SQLType:   sqlt,
				Kind:      kind,
				Nullable:  nullable,
			})
		}
	}
	return fields, nil
}

func columnFromTag(tag string) string {
	if v := structTagValue(tag, "db"); v != "" {
		return v
	}
	if v := structTagValue(tag, "bson"); v != "" {
		if v == "-" {
			return ""
		}
		return v
	}
	return ""
}

func updateDocKey(tag, column string) string {
	if v := structTagValue(tag, "bson"); v != "" && v != "-" {
		return v
	}
	if v := firstJSONField(structTagValue(tag, "json")); v != "" && v != "-" {
		return v
	}
	return column
}

func firstJSONField(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	i := strings.IndexByte(s, ',')
	if i < 0 {
		return s
	}
	return strings.TrimSpace(s[:i])
}

func structTagValue(tag, key string) string {
	t := reflectTag(tag)
	return t.Get(key)
}

// reflectTag parses `...` tag string without importing reflect on hot path in cmd — tiny parser.
func reflectTag(tag string) tagLookup {
	return tagLookup{raw: tag}
}

type tagLookup struct {
	raw string
}

func (t tagLookup) Get(key string) string {
	for t.raw != "" {
		i := 0
		for i < len(t.raw) && t.raw[i] == ' ' {
			i++
		}
		t.raw = t.raw[i:]
		if t.raw == "" {
			break
		}
		i = 0
		for i < len(t.raw) && t.raw[i] > ' ' && t.raw[i] != ':' && t.raw[i] != '"' {
			i++
		}
		if i == 0 || i >= len(t.raw) || t.raw[i] != ':' {
			break
		}
		name := t.raw[:i]
		t.raw = t.raw[i+1:]
		for len(t.raw) > 0 && t.raw[0] == ' ' {
			t.raw = t.raw[1:]
		}
		if len(t.raw) == 0 || t.raw[0] != '"' {
			break
		}
		t.raw = t.raw[1:]
		i = 0
		for i < len(t.raw) && t.raw[i] != '"' {
			if t.raw[i] == '\\' && i+1 < len(t.raw) {
				i++
			}
			i++
		}
		val := t.raw[:i]
		if i < len(t.raw) {
			t.raw = t.raw[i+1:]
		} else {
			t.raw = ""
		}
		if name == key {
			return val
		}
	}
	return ""
}

func goTypeKind(expr ast.Expr) (TypeKind, bool, error) {
	nullable := false
	e := expr
	if star, ok := expr.(*ast.StarExpr); ok {
		nullable = true
		e = star.X
	}
	id, ok := e.(*ast.Ident)
	if !ok {
		return 0, false, fmt.Errorf("unsupported type %s", typeString(expr))
	}
	switch id.Name {
	case "string":
		return KindString, nullable, nil
	case "int64":
		return KindInt64, nullable, nil
	case "bool":
		return KindBool, nullable, nil
	default:
		return 0, false, fmt.Errorf("unsupported type %q (use string, *string, int64, *int64, bool, *bool)", id.Name)
	}
}

func typeString(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.StarExpr:
		return "*" + typeString(x.X)
	default:
		return fmt.Sprintf("%T", e)
	}
}

func sqlType(k TypeKind, nullable bool) string {
	switch k {
	case KindString:
		if nullable {
			return "TEXT"
		}
		return "TEXT NOT NULL"
	case KindInt64:
		if nullable {
			return "BIGINT"
		}
		return "BIGINT NOT NULL"
	case KindBool:
		if nullable {
			return "BOOLEAN"
		}
		return "BOOLEAN NOT NULL"
	default:
		return "TEXT"
	}
}
