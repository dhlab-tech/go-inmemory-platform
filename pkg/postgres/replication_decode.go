package postgres

import (
	"fmt"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5/pgtype"
)

// DecodeText assigns a logical replication text/binary cell into *dest (empty string if NULL).
func DecodeText(m *pgtype.Map, oid uint32, col pglogrepl.TupleDataColumn, dest *string) error {
	if col.DataType == pglogrepl.TupleDataTypeNull {
		*dest = ""
		return nil
	}
	format := int16(0)
	if col.DataType == pglogrepl.TupleDataTypeBinary {
		format = 1
	}
	return m.Scan(oid, format, col.Data, dest)
}

// DecodeTextPtr assigns into *string or nil if NULL.
func DecodeTextPtr(m *pgtype.Map, oid uint32, col pglogrepl.TupleDataColumn, dest **string) error {
	if col.DataType == pglogrepl.TupleDataTypeNull {
		*dest = nil
		return nil
	}
	format := int16(0)
	if col.DataType == pglogrepl.TupleDataTypeBinary {
		format = 1
	}
	var s string
	if err := m.Scan(oid, format, col.Data, &s); err != nil {
		return err
	}
	*dest = &s
	return nil
}

// DecodeInt8 decodes INT8 (bigint) OID 20 into *dest.
func DecodeInt8(m *pgtype.Map, oid uint32, col pglogrepl.TupleDataColumn, dest *int64) error {
	if col.DataType == pglogrepl.TupleDataTypeNull {
		*dest = 0
		return nil
	}
	format := int16(0)
	if col.DataType == pglogrepl.TupleDataTypeBinary {
		format = 1
	}
	return m.Scan(oid, format, col.Data, dest)
}

// DecodeInt8Ptr decodes nullable bigint.
func DecodeInt8Ptr(m *pgtype.Map, oid uint32, col pglogrepl.TupleDataColumn, dest **int64) error {
	if col.DataType == pglogrepl.TupleDataTypeNull {
		*dest = nil
		return nil
	}
	format := int16(0)
	if col.DataType == pglogrepl.TupleDataTypeBinary {
		format = 1
	}
	var v int64
	if err := m.Scan(oid, format, col.Data, &v); err != nil {
		return err
	}
	*dest = &v
	return nil
}

// DecodeBool decodes bool into *dest.
func DecodeBool(m *pgtype.Map, oid uint32, col pglogrepl.TupleDataColumn, dest *bool) error {
	if col.DataType == pglogrepl.TupleDataTypeNull {
		*dest = false
		return nil
	}
	format := int16(0)
	if col.DataType == pglogrepl.TupleDataTypeBinary {
		format = 1
	}
	return m.Scan(oid, format, col.Data, dest)
}

// DecodeBoolPtr decodes nullable bool.
func DecodeBoolPtr(m *pgtype.Map, oid uint32, col pglogrepl.TupleDataColumn, dest **bool) error {
	if col.DataType == pglogrepl.TupleDataTypeNull {
		*dest = nil
		return nil
	}
	format := int16(0)
	if col.DataType == pglogrepl.TupleDataTypeBinary {
		format = 1
	}
	var v bool
	if err := m.Scan(oid, format, col.Data, &v); err != nil {
		return err
	}
	*dest = &v
	return nil
}

// ColumnIndexByName returns the index of name in rel.Columns, or -1.
func ColumnIndexByName(rel *pglogrepl.RelationMessageV2, name string) int {
	for i, c := range rel.Columns {
		if c.Name == name {
			return i
		}
	}
	return -1
}

// TupleColumn returns the tuple cell at idx or nil if out of range / missing.
func TupleColumn(tuple *pglogrepl.TupleData, idx int) *pglogrepl.TupleDataColumn {
	if tuple == nil || idx < 0 || idx >= len(tuple.Columns) {
		return nil
	}
	return tuple.Columns[idx]
}

// TupleColumnOID returns the PostgreSQL type OID for rel.Columns[idx].
func TupleColumnOID(rel *pglogrepl.RelationMessageV2, idx int) (uint32, error) {
	if rel == nil || idx < 0 || idx >= len(rel.Columns) {
		return 0, fmt.Errorf("postgres: replication: invalid column index %d", idx)
	}
	return rel.Columns[idx].DataType, nil
}
