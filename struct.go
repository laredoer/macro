package macro

import "strings"

type Field struct {
	FieldName string
	FieldType string
	FieldTag  string
}

type Struct struct {
	Name        string
	Fields      []Field
	Annotations []string
}

func (s *Struct) GetRecv() string {
	return strings.ToLower(s.Name)
}

func (s *Struct) GetRecvType() string {
	return "*" + s.Name
}

func (s *Struct) GetContent() string {
	var buf strings.Builder
	buf.WriteString(s.Name)
	buf.WriteByte('(')
	for _, field := range s.Fields {
		buf.WriteString(field.FieldName)
		buf.WriteByte('=')
		buf.WriteString(field.FieldType)
		buf.WriteByte(',')
	}
	buf.WriteByte(')')
	return buf.String()
}
