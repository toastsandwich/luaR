package luar

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

type Decoder struct {
	program *Program
}

func Unmarshal(data []byte, v interface{}) error {
	return NewDecoder(strings.NewReader(string(data))).Decode(v)
}

func NewDecoder(r io.Reader) *Decoder {
	data, _ := io.ReadAll(r)
	parser := NewParser(string(data))
	program, _ := parser.Parse()
	return &Decoder{program: program}
}

func (d *Decoder) Decode(v interface{}) error {
	return d.decode(v)
}

func (d *Decoder) decode(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("luar: expected pointer, got %v", rv.Kind())
	}

	rv = rv.Elem()

	assignments := d.getTopLevelAssignments()

	for _, assign := range assignments {
		if len(assign.Names) != 1 || len(assign.Values) != 1 {
			continue
		}

		name := assign.Names[0].Name
		value := assign.Values[0]

		fieldName := d.findFieldByTag(rv, name)
		if fieldName == "" {
			fieldName = name
		}

		field := rv.FieldByName(fieldName)
		if !field.IsValid() {
			continue
		}

		val, err := d.evalExpression(value)
		if err != nil {
			continue
		}

		if err := d.setValue(field, val); err != nil {
			continue
		}
	}

	return nil
}

func (d *Decoder) getTopLevelAssignments() []*AssignmentStatement {
	var assignments []*AssignmentStatement
	for _, stmt := range d.program.Statements {
		if assign, ok := stmt.(*AssignmentStatement); ok {
			assignments = append(assignments, assign)
		}
	}
	return assignments
}

func (d *Decoder) findFieldByTag(rv reflect.Value, luaName string) string {
	t := rv.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("lua")
		if tag == "" {
			tag = strings.ToLower(field.Name)
		}
		if tag == luaName {
			return field.Name
		}
	}
	return ""
}

func (d *Decoder) setValue(field reflect.Value, val interface{}) error {
	if !field.CanSet() {
		return fmt.Errorf("luar: cannot set unexported field")
	}

	switch field.Kind() {
	case reflect.String:
		if str, ok := val.(string); ok {
			field.SetString(str)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if n, ok := toInt64(val); ok {
			field.SetInt(n)
		}
	case reflect.Float32, reflect.Float64:
		field.SetFloat(toFloat64(val))
	case reflect.Bool:
		if b, ok := val.(bool); ok {
			field.SetBool(b)
		}
	case reflect.Slice:
		if slice, ok := val.([]interface{}); ok {
			sliceType := field.Type()
			elemType := sliceType.Elem()
			newSlice := reflect.MakeSlice(sliceType, len(slice), len(slice))
			for i, item := range slice {
				elem := reflect.New(elemType).Elem()
				if err := d.setValue(elem, item); err != nil {
					continue
				}
				newSlice.Index(i).Set(elem)
			}
			field.Set(newSlice)
		}
	case reflect.Map:
		if m, ok := val.(map[string]interface{}); ok {
			mapType := field.Type()
			mapVal := reflect.MakeMap(mapType)
			for k, v := range m {
				key := reflect.ValueOf(k)
				elem := reflect.New(mapType.Elem()).Elem()
				d.setValue(elem, v)
				mapVal.SetMapIndex(key, elem)
			}
			field.Set(mapVal)
		}
	case reflect.Struct:
		if m, ok := val.(map[string]interface{}); ok {
			for k, v := range m {
				fieldName := d.findFieldByTag(field, k)
				if fieldName == "" {
					fieldName = k
				}
				f := field.FieldByName(fieldName)
				if f.IsValid() {
					d.setValue(f, v)
				}
			}
		}
	}

	return nil
}

func (d *Decoder) evalExpression(expr Expression) (interface{}, error) {
	switch e := expr.(type) {
	case *Identifier:
		val := d.findVariable(e.Name)
		return val, nil
	case *NumberLiteral:
		if e.IsInt {
			return e.IntValue, nil
		}
		return e.Value, nil
	case *StringLiteral:
		return e.Value, nil
	case *BooleanLiteral:
		return e.Value, nil
	case *NilLiteral:
		return nil, nil
	case *TableLiteral:
		return d.evalTableLiteral(e)
	case *BinaryExpression:
		return d.evalBinaryExpression(e)
	default:
		return nil, nil
	}
}

func (d *Decoder) findVariable(name string) interface{} {
	assignments := d.getTopLevelAssignments()
	for _, assign := range assignments {
		if len(assign.Names) == 1 && assign.Names[0].Name == name && len(assign.Values) == 1 {
			return d.evalExpressionValue(assign.Values[0])
		}
	}
	return nil
}

func (d *Decoder) evalExpressionValue(expr Expression) interface{} {
	val, _ := d.evalExpression(expr)
	return val
}

func (d *Decoder) evalTableLiteral(t *TableLiteral) (interface{}, error) {
	result := make(map[string]interface{})

	for _, field := range t.Fields {
		var key string

		if ident, ok := field.Key.(*Identifier); ok {
			key = ident.Name
		} else if str, ok := field.Key.(*StringLiteral); ok {
			key = str.Value
		} else if num, ok := field.Key.(*NumberLiteral); ok {
			key = strconv.FormatFloat(num.Value, 'f', -1, 64)
		} else if idx, ok := field.Key.(*TableIndex); ok {
			if ident, ok := idx.Key.(*Identifier); ok {
				key = ident.Name
			} else if str, ok := idx.Key.(*StringLiteral); ok {
				key = str.Value
			}
		}

		value := d.evalExpressionValue(field.Value)

		if key != "" {
			result[key] = value
		} else {
			result[strconv.Itoa(len(result))] = value
		}
	}

	return result, nil
}

func (d *Decoder) evalBinaryExpression(e *BinaryExpression) (interface{}, error) {
	left := d.evalExpressionValue(e.Left)
	right := d.evalExpressionValue(e.Right)

	switch e.Operator {
	case PLUS:
		if isNumber(left) && isNumber(right) {
			return toFloat64(left) + toFloat64(right), nil
		}
		if isString(left) && isString(right) {
			return toString(left) + toString(right), nil
		}
	case MINUS:
		if isNumber(left) && isNumber(right) {
			return toFloat64(left) - toFloat64(right), nil
		}
	case STAR:
		if isNumber(left) && isNumber(right) {
			return toFloat64(left) * toFloat64(right), nil
		}
	case SLASH:
		if isNumber(left) && isNumber(right) {
			return toFloat64(left) / toFloat64(right), nil
		}
	case EQ:
		return reflect.DeepEqual(left, right), nil
	case NE:
		return !reflect.DeepEqual(left, right), nil
	case LT:
		if isNumber(left) && isNumber(right) {
			return toFloat64(left) < toFloat64(right), nil
		}
	case LE:
		if isNumber(left) && isNumber(right) {
			return toFloat64(left) <= toFloat64(right), nil
		}
	case GT:
		if isNumber(left) && isNumber(right) {
			return toFloat64(left) > toFloat64(right), nil
		}
	case GE:
		if isNumber(left) && isNumber(right) {
			return toFloat64(left) >= toFloat64(right), nil
		}
	}

	return nil, nil
}

func isNumber(v interface{}) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, float32, float64:
		return true
	}
	return false
}

func isString(v interface{}) bool {
	_, ok := v.(string)
	return ok
}

func toFloat64(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	}
	return 0
}

func toInt64(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case int64:
		return n, true
	case int:
		return int64(n), true
	case float64:
		return int64(n), true
	}
	return 0, false
}

func toString(v interface{}) string {
	s, _ := v.(string)
	return s
}

type Encoder struct {
	w           io.Writer
	indent      string
	indentLevel int
}

func Marshal(v interface{}) ([]byte, error) {
	var buf strings.Builder
	encoder := NewEncoder(&buf)
	err := encoder.Encode(v)
	if err != nil {
		return nil, err
	}
	return []byte(buf.String()), nil
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w, indent: "    "}
}

func (e *Encoder) Encode(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Struct {
		return e.encodeStructAsAssignments(rv)
	}

	return e.encodeValue(rv, false)
}

func (e *Encoder) encodeStructAsAssignments(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("lua")
		if tag == "" {
			tag = strings.ToLower(field.Name)
		}

		fieldVal := v.FieldByName(field.Name)
		if !fieldVal.IsValid() {
			continue
		}

		e.writeString(tag)
		e.writeString(" = ")
		e.encodeValue(fieldVal, true)
		e.writeString("\n")
	}

	return nil
}

func (e *Encoder) encodeValue(v reflect.Value, isTableValue bool) error {
	if !v.IsValid() {
		e.writeString("nil")
		return nil
	}

	switch v.Kind() {
	case reflect.String:
		e.writeString(fmt.Sprintf("%q", v.String()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		e.writeString(fmt.Sprintf("%d", v.Int()))
	case reflect.Float32, reflect.Float64:
		e.writeString(fmt.Sprintf("%g", v.Float()))
	case reflect.Bool:
		if v.Bool() {
			e.writeString("true")
		} else {
			e.writeString("false")
		}
	case reflect.Slice:
		if v.IsNil() {
			e.writeString("nil")
			return nil
		}
		e.writeString("{")
		e.indentLevel++
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				e.writeString(", ")
			}
			e.encodeValue(v.Index(i), false)
		}
		e.indentLevel--
		e.writeString("}")
	case reflect.Map:
		e.writeString("{")
		e.indentLevel++
		keys := v.MapKeys()
		first := true
		for _, key := range keys {
			if !first {
				e.writeString(", ")
			}
			first = false
			e.writeString(e.getLuaTag(key))
			e.writeString(" = ")
			e.encodeValue(v.MapIndex(key), true)
		}
		e.indentLevel--
		e.writeString("}")
	case reflect.Struct:
		e.encodeStruct(v)
	case reflect.Ptr:
		if v.IsNil() {
			e.writeString("nil")
			return nil
		}
		return e.encodeValue(v.Elem(), isTableValue)
	default:
		e.writeString("nil")
	}
	return nil
}

func (e *Encoder) encodeStruct(v reflect.Value) error {
	t := v.Type()
	e.writeString("{")
	e.indentLevel++

	fields := []string{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("lua")
		if tag == "" {
			tag = strings.ToLower(field.Name)
		}

		fieldVal := v.FieldByName(field.Name)
		if !fieldVal.IsValid() {
			continue
		}

		if fields = append(fields, tag); len(fields) > 1 {
			e.writeString(", ")
		}

		e.writeString(tag)
		e.writeString(" = ")
		e.encodeValue(fieldVal, true)
	}

	e.indentLevel--
	e.writeString("}")
	return nil
}

func (e *Encoder) getLuaTag(v reflect.Value) string {
	if v.Kind() != reflect.Struct {
		return ""
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		return field.Tag.Get("lua")
	}
	return ""
}

func (e *Encoder) writeString(s string) {
	e.w.Write([]byte(s))
}
