package dynamic

// Marshalling and unmarshalling of dynamic messages to/from proto's standard text format

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"text/scanner"
	"unicode"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"

	"github.com/jhump/protoreflect/desc"
)

func (m *Message) MarshalText() ([]byte, error) {
	var b indentBuffer
	b.indentCount = -1 // no indentation
	if err := m.marshalText(&b); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (m *Message) MarshalTextIndent() ([]byte, error) {
	var b indentBuffer
	b.indent = "  " // TODO: option for indent?
	if err := m.marshalText(&b); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (m *Message) marshalText(b *indentBuffer) error {
	first := true
	// first the known fields
	for _, tag := range m.knownFieldTags() {
		itag := int32(tag)
		v := m.values[itag]
		fd := m.FindFieldDescriptor(itag)
		if fd.IsMap() {
			md := fd.GetMessageType()
			kfd := md.FindFieldByNumber(1)
			vfd := md.FindFieldByNumber(2)
			mp := v.(map[interface{}]interface{})
			keys := make([]interface{}, 0, len(mp))
			for k := range mp {
				keys = append(keys, k)
			}
			sort.Sort(sortable(keys))
			for _, mk := range keys {
				mv := mp[mk]
				err := b.maybeNext(&first)
				if err != nil {
					return err
				}
				err = marshalKnownFieldMapEntryText(b, fd, kfd, mk, vfd, mv)
				if err != nil {
					return err
				}
			}
		} else if fd.IsRepeated() {
			sl := v.([]interface{})
			for _, slv := range sl {
				err := b.maybeNext(&first)
				if err != nil {
					return err
				}
				err = marshalKnownFieldText(b, fd, slv)
				if err != nil {
					return err
				}
			}
		} else {
			err := b.maybeNext(&first)
			if err != nil {
				return err
			}
			err = marshalKnownFieldText(b, fd, v)
			if err != nil {
				return err
			}
		}
	}
	// then the unknown fields
	for _, tag := range m.unknownFieldTags() {
		itag := int32(tag)
		ufs := m.unknownFields[itag]
		for _, uf := range ufs {
			err := b.maybeNext(&first)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(b, "%d", tag)
			if err != nil {
				return err
			}
			if uf.Encoding == proto.WireStartGroup {
				err = b.WriteByte('{')
				if err != nil {
					return err
				}
				err = b.start()
				if err != nil {
					return err
				}
				in := newCodedBuffer(uf.Contents)
				err = marshalUnknownGroupText(b, in, true)
				if err != nil {
					return err
				}
				err = b.end()
				if err != nil {
					return err
				}
				err = b.WriteByte('}')
				if err != nil {
					return err
				}
			} else {
				err = b.sep()
				if err != nil {
					return err
				}
				if uf.Encoding == proto.WireBytes {
					err = writeString(b, string(uf.Contents))
					if err != nil {
						return err
					}
				} else {
					_, err = b.WriteString(strconv.FormatUint(uf.Value, 10))
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func marshalKnownFieldMapEntryText(b *indentBuffer, fd *desc.FieldDescriptor, kfd *desc.FieldDescriptor, mk interface{}, vfd *desc.FieldDescriptor, mv interface{}) error {
	var name string
	if fd.IsExtension() {
		name = fmt.Sprintf("[%s]", fd.GetFullyQualifiedName())
	} else {
		name = fd.GetName()
	}
	_, err := b.WriteString(name)
	if err != nil {
		return err
	}
	err = b.sep()
	if err != nil {
		return err
	}

	err = b.WriteByte('<')
	if err != nil {
		return err
	}
	err = b.start()
	if err != nil {
		return err
	}

	err = marshalKnownFieldText(b, kfd, mk)
	if err != nil {
		return err
	}
	err = b.next()
	if err != nil {
		return err
	}
	err = marshalKnownFieldText(b, vfd, mv)
	if err != nil {
		return err
	}

	err = b.end()
	if err != nil {
		return err
	}
	return b.WriteByte('>')
}

func marshalKnownFieldText(b *indentBuffer, fd *desc.FieldDescriptor, v interface{}) error {
	group := fd.GetType() == descriptor.FieldDescriptorProto_TYPE_GROUP
	if group {
		var name string
		if fd.IsExtension() {
			name = fmt.Sprintf("[%s]", fd.GetMessageType().GetFullyQualifiedName())
		} else {
			name = fd.GetMessageType().GetName()
		}
		_, err := b.WriteString(name)
		if err != nil {
			return err
		}
	} else {
		var name string
		if fd.IsExtension() {
			name = fmt.Sprintf("[%s]", fd.GetFullyQualifiedName())
		} else {
			name = fd.GetName()
		}
		_, err := b.WriteString(name)
		if err != nil {
			return err
		}
		err = b.sep()
		if err != nil {
			return err
		}
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int32, reflect.Int64:
		ed := fd.GetEnumType()
		if ed != nil {
			n := int32(rv.Int())
			vd := ed.FindValueByNumber(n)
			if vd == nil {
				_, err := b.WriteString(strconv.FormatInt(rv.Int(), 10))
				return err
			} else {
				_, err := b.WriteString(vd.GetName())
				return err
			}
		} else {
			_, err := b.WriteString(strconv.FormatInt(rv.Int(), 10))
			return err
		}
	case reflect.Uint32, reflect.Uint64:
		_, err := b.WriteString(strconv.FormatUint(rv.Uint(), 10))
		return err
	case reflect.Float32, reflect.Float64:
		f := rv.Float()
		var str string
		if math.IsNaN(f) {
			str = "nan"
		} else if math.IsInf(f, 1) {
			str = "inf"
		} else if math.IsInf(f, -1) {
			str = "-inf"
		} else {
			var bits int
			if rv.Kind() == reflect.Float32 {
				bits = 32
			} else {
				bits = 64
			}
			str = strconv.FormatFloat(rv.Float(), 'g', -1, bits)
		}
		_, err := b.WriteString(str)
		return err
	case reflect.Bool:
		_, err := b.WriteString(strconv.FormatBool(rv.Bool()))
		return err
	case reflect.Slice:
		return writeString(b, string(rv.Bytes()))
	case reflect.String:
		return writeString(b, rv.String())
	default:
		var err error
		if group {
			err = b.WriteByte('{')
		} else {
			err = b.WriteByte('<')
		}
		if err != nil {
			return err
		}
		err = b.start()
		if err != nil {
			return err
		}
		// must be a message
		if dm, ok := v.(*Message); ok {
			err = dm.marshalText(b)
			if err != nil {
				return err
			}
		} else {
			err = proto.CompactText(b, v.(proto.Message))
			if err != nil {
				return err
			}
		}
		err = b.end()
		if err != nil {
			return err
		}
		if group {
			return b.WriteByte('}')
		} else {
			return b.WriteByte('>')
		}
	}
}

// writeString writes a string in the protocol buffer text format.
// It is similar to strconv.Quote except we don't use Go escape sequences,
// we treat the string as a byte sequence, and we use octal escapes.
// These differences are to maintain interoperability with the other
// languages' implementations of the text format.
func writeString(b *indentBuffer, s string) error {
	// use WriteByte here to get any needed indent
	if err := b.WriteByte('"'); err != nil {
		return err
	}
	// Loop over the bytes, not the runes.
	for i := 0; i < len(s); i++ {
		var err error
		// Divergence from C++: we don't escape apostrophes.
		// There's no need to escape them, and the C++ parser
		// copes with a naked apostrophe.
		switch c := s[i]; c {
		case '\n':
			_, err = b.WriteString("\\n")
		case '\r':
			_, err = b.WriteString("\\r")
		case '\t':
			_, err = b.WriteString("\\t")
		case '"':
			_, err = b.WriteString("\\")
		case '\\':
			_, err = b.WriteString("\\\\")
		default:
			if c >= 0x20 && c < 0x7f {
				err = b.WriteByte(c)
			} else {
				_, err = fmt.Fprintf(b, "\\%03o", c)
			}
		}
		if err != nil {
			return err
		}
	}
	return b.WriteByte('"')
}

func marshalUnknownGroupText(b *indentBuffer, in *codedBuffer, topLevel bool) error {
	first := true
	for {
		if in.eof() {
			if topLevel {
				return nil
			}
			// this is a nested message: we are expecting an end-group tag, not EOF!
			return io.ErrUnexpectedEOF
		}
		tag, wireType, err := in.decodeTagAndWireType()
		if err != nil {
			return err
		}
		if wireType == proto.WireEndGroup {
			return nil
		}
		err = b.maybeNext(&first)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(b, "%d", tag)
		if err != nil {
			return err
		}
		if wireType == proto.WireStartGroup {
			err = b.WriteByte('{')
			if err != nil {
				return err
			}
			err = b.start()
			if err != nil {
				return err
			}
			err = marshalUnknownGroupText(b, in, false)
			if err != nil {
				return err
			}
			err = b.end()
			if err != nil {
				return err
			}
			err = b.WriteByte('}')
			if err != nil {
				return err
			}
			continue
		} else {
			err = b.sep()
			if err != nil {
				return err
			}
			if wireType == proto.WireBytes {
				contents, err := in.decodeRawBytes(false)
				if err != nil {
					return err
				}
				err = writeString(b, string(contents))
				if err != nil {
					return err
				}
			} else {
				var v uint64
				switch wireType {
				case proto.WireVarint:
					v, err = in.decodeVarint()
				case proto.WireFixed32:
					v, err = in.decodeFixed32()
				case proto.WireFixed64:
					v, err = in.decodeFixed64()
				default:
					return proto.ErrInternalBadWireType
				}
				if err != nil {
					return err
				}
				_, err = b.WriteString(strconv.FormatUint(v, 10))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (m *Message) UnmarshalText(text []byte) error {
	m.Reset()
	if err := m.UnmarshalMergeText(text); err != nil {
		return err
	}
	return m.Validate()
}

func (m *Message) UnmarshalMergeText(text []byte) error {
	return m.unmarshalText(newReader(text), tokenEOF)
}

func (m *Message) unmarshalText(tr *txtReader, end tokenType) error {
	for {
		tok := tr.next()
		if tok.tokTyp == end {
			return nil
		}
		if tok.tokTyp == tokenEOF {
			return io.ErrUnexpectedEOF
		}

		var fd *desc.FieldDescriptor
		if tok.tokTyp == tokenInt {
			// tag number (indicates unknown field)
			tag, err := strconv.ParseInt(tok.val.(string), 10, 32)
			if err != nil {
				return err
			}
			itag := int32(tag)
			fd = m.FindFieldDescriptor(itag)
			if fd == nil {
				// can't parse the value w/out field descriptor, so skip it
				tok = tr.next()
				if tok.tokTyp == tokenEOF {
					return io.ErrUnexpectedEOF
				} else if tok.tokTyp == tokenOpenBrace {
					if err := skipMessageText(tr, true); err != nil {
						return err
					}
				} else if tok.tokTyp == tokenColon {
					if err := skipFieldValueText(tr); err != nil {
						return err
					}
				} else {
					return textError(tok, "Expecting a colon ':' or brace '{'; instead got %q", tok.txt)
				}
				tok = tr.peek()
				if tok.tokTyp.IsSep() {
					tr.next() // consume separator
				}
				continue
			}
		} else {
			fieldName, err := unmarshalFieldNameText(tr, tok)
			if err != nil {
				return err
			}
			fd = m.FindFieldDescriptorByName(fieldName)
			if fd == nil {
				// See if it's a group name
				for _, field := range m.md.GetFields() {
					if field.GetType() == descriptor.FieldDescriptorProto_TYPE_GROUP && field.GetMessageType().GetName() == fieldName {
						fd = field
						break
					}
				}
				if fd == nil {
					// TODO: add a flag to just ignore unrecognized field names
					return textError(tok, "%q is not a recognized field name of %q", fieldName, m.md.GetFullyQualifiedName())
				}
			}
		}
		tok = tr.next()
		if tok.tokTyp == tokenEOF {
			return io.ErrUnexpectedEOF
		}
		if (fd.GetType() == descriptor.FieldDescriptorProto_TYPE_GROUP ||
			fd.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE) &&
			tok.tokTyp.EndToken() != tokenError {

			// TODO: use mf.NewMessage and, if not a dynamic message, use proto.UnmarshalText to unmarshal it
			g := m.mf.NewDynamicMessage(fd.GetMessageType())
			if err := g.unmarshalText(tr, tok.tokTyp.EndToken()); err != nil {
				return err
			}
			if fd.IsRepeated() {
				if err := m.TryAddRepeatedField(fd, g); err != nil {
					return err
				}
			} else {
				if err := m.TrySetField(fd, g); err != nil {
					return err
				}
			}
		} else {
			if tok.tokTyp != tokenColon {
				return textError(tok, "Expecting a colon ':'; instead got %q", tok.txt)
			}
			if err := m.unmarshalFieldValueText(fd, tr); err != nil {
				return err
			}
		}
		tok = tr.peek()
		if tok.tokTyp.IsSep() {
			tr.next() // consume separator
		}
	}
	return nil
}

func textError(tok *token, format string, args ...interface{}) error {
	var msg string
	if tok.tokTyp == tokenError {
		msg = tok.val.(error).Error()
	} else {
		msg = fmt.Sprintf(format, args...)
	}
	return fmt.Errorf("Line %d, col %d: %s", tok.pos.Line, tok.pos.Column, msg)
}

type setFunction func(*Message, *desc.FieldDescriptor, interface{}) error

func (m *Message) unmarshalFieldValueText(fd *desc.FieldDescriptor, tr *txtReader) error {
	var set setFunction
	if fd.IsRepeated() {
		set = (*Message).addRepeatedField
	} else {
		set = mergeField
	}
	tok := tr.peek()
	if tok.tokTyp == tokenOpenBracket {
		tr.next() // consume tok
		for {
			if err := m.unmarshalFieldElementText(fd, tr, set); err != nil {
				return err
			}
			tok = tr.peek()
			if tok.tokTyp == tokenCloseBracket {
				tr.next() // consume tok
				return nil
			} else if tok.tokTyp.IsSep() {
				tr.next() // consume separator
			}
		}
	}
	return m.unmarshalFieldElementText(fd, tr, set)
}

func (m *Message) unmarshalFieldElementText(fd *desc.FieldDescriptor, tr *txtReader, set setFunction) error {
	tok := tr.next()
	if tok.tokTyp == tokenEOF {
		return io.ErrUnexpectedEOF
	}

	var expected string
	switch fd.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		if tok.tokTyp == tokenIdent {
			if tok.val.(string) == "true" {
				return set(m, fd, true)
			} else if tok.val.(string) == "false" {
				return set(m, fd, false)
			}
		}
		expected = "boolean value"
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		if tok.tokTyp == tokenString {
			return set(m, fd, []byte(tok.val.(string)))
		}
		expected = "bytes string value"
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		if tok.tokTyp == tokenString {
			return set(m, fd, tok.val)
		}
		expected = "string value"
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		if tok.tokTyp == tokenFloat {
			return set(m, fd, float32(tok.val.(float64)))
		} else if tok.tokTyp == tokenInt {
			if f, err := strconv.ParseFloat(tok.val.(string), 32); err != nil {
				return err
			} else {
				return set(m, fd, float32(f))
			}
		}
		expected = "float value"
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		if tok.tokTyp == tokenFloat {
			return set(m, fd, tok.val)
		} else if tok.tokTyp == tokenInt {
			if f, err := strconv.ParseFloat(tok.val.(string), 64); err != nil {
				return err
			} else {
				return set(m, fd, f)
			}
		}
		expected = "float value"
	case descriptor.FieldDescriptorProto_TYPE_INT32,
		descriptor.FieldDescriptorProto_TYPE_SINT32,
		descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		if tok.tokTyp == tokenInt {
			if i, err := strconv.ParseInt(tok.val.(string), 10, 32); err != nil {
				return err
			} else {
				return set(m, fd, int32(i))
			}
		}
		expected = "int value"
	case descriptor.FieldDescriptorProto_TYPE_INT64,
		descriptor.FieldDescriptorProto_TYPE_SINT64,
		descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		if tok.tokTyp == tokenInt {
			if i, err := strconv.ParseInt(tok.val.(string), 10, 64); err != nil {
				return err
			} else {
				return set(m, fd, i)
			}
		}
		expected = "int value"
	case descriptor.FieldDescriptorProto_TYPE_UINT32,
		descriptor.FieldDescriptorProto_TYPE_FIXED32:
		if tok.tokTyp == tokenInt {
			if i, err := strconv.ParseUint(tok.val.(string), 10, 32); err != nil {
				return err
			} else {
				return set(m, fd, uint32(i))
			}
		}
		expected = "unsigned int value"
	case descriptor.FieldDescriptorProto_TYPE_UINT64,
		descriptor.FieldDescriptorProto_TYPE_FIXED64:
		if tok.tokTyp == tokenInt {
			if i, err := strconv.ParseUint(tok.val.(string), 10, 64); err != nil {
				return err
			} else {
				return set(m, fd, i)
			}
		}
		expected = "unsigned int value"
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		if tok.tokTyp == tokenIdent {
			// TODO: add a flag to just ignore unrecognized enum value names?
			vd := fd.GetEnumType().FindValueByName(tok.val.(string))
			if vd != nil {
				return set(m, fd, vd.GetNumber())
			}
		} else if tok.tokTyp == tokenInt {
			if i, err := strconv.ParseInt(tok.val.(string), 10, 32); err != nil {
				return err
			} else {
				return set(m, fd, int32(i))
			}
		}
		expected = fmt.Sprintf("enum %s value", fd.GetEnumType().GetFullyQualifiedName())
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE,
		descriptor.FieldDescriptorProto_TYPE_GROUP:

		endTok := tok.tokTyp.EndToken()
		if endTok != tokenError {
			dm := newMessageWithMessageFactory(fd.GetMessageType(), m.mf)
			if err := dm.unmarshalText(tr, endTok); err != nil {
				return err
			}
			// TODO: ideally we would use mf.NewMessage and, if not a dynamic message, use
			// jsonpb to unmarshal it. But the text parser isn't particularly amenable to that
			// so we instead convert a dynamic message to a generated one if the known-type
			// registry knows about the generated type...
			var ktr *KnownTypeRegistry
			if m.mf != nil {
				ktr = m.mf.ktr
			}
			pm := ktr.CreateIfKnown(fd.GetMessageType().GetFullyQualifiedName())
			if pm != nil {
				if err := dm.ConvertTo(pm); err != nil {
					return set(m, fd, pm)
				}
			}
			return set(m, fd, dm)
		}
		expected = fmt.Sprintf("message %s value", fd.GetMessageType().GetFullyQualifiedName())
	default:
		return fmt.Errorf("field %q of message %q has unrecognized type: %v", fd.GetFullyQualifiedName(), m.md.GetFullyQualifiedName(), fd.GetType())
	}

	// if we get here, token was wrong type; create error message
	var article string
	if strings.Index("aieou", expected[0:1]) != -1 {
		article = "an"
	} else {
		article = "a"
	}
	return textError(tok, "Expecting %s %s; got %q", article, expected, tok.txt)
}

func unmarshalFieldNameText(tr *txtReader, tok *token) (string, error) {
	if tok.tokTyp == tokenOpenBracket || tok.tokTyp == tokenOpenParen {
		// extension name
		var closeType tokenType
		var closeChar string
		if tok.tokTyp == tokenOpenBracket {
			closeType = tokenCloseBracket
			closeChar = "close bracket ']'"
		} else {
			closeType = tokenCloseParen
			closeChar = "close paren ')'"
		}
		// must be followed by an identifier
		tok = tr.next()
		if tok.tokTyp == tokenEOF {
			return "", io.ErrUnexpectedEOF
		} else if tok.tokTyp != tokenIdent {
			return "", textError(tok, "Expecting an identifier; instead got %q", tok.txt)
		}
		ident := tok.val.(string)
		// and then close bracket/paren
		tok = tr.next()
		if tok.tokTyp == tokenEOF {
			return "", io.ErrUnexpectedEOF
		} else if tok.tokTyp != closeType {
			return "", textError(tok, "Expecting a %s; instead got %q", closeChar, tok.txt)
		}
		return "[" + ident + "]", nil
	} else if tok.tokTyp == tokenIdent {
		// normal field name
		return tok.val.(string), nil
	} else {
		return "", textError(tok, "Expecting an identifier or tag number; instead got %q", tok.txt)
	}
}

func skipFieldNameText(tr *txtReader) error {
	tok := tr.next()
	if tok.tokTyp == tokenEOF {
		return io.ErrUnexpectedEOF
	} else if tok.tokTyp == tokenInt || tok.tokTyp == tokenIdent {
		return nil
	} else {
		_, err := unmarshalFieldNameText(tr, tok)
		return err
	}
}

func skipFieldValueText(tr *txtReader) error {
	tok := tr.peek()
	if tok.tokTyp == tokenOpenBracket {
		tr.next() // consume tok
		for {
			if err := skipFieldElementText(tr); err != nil {
				return err
			}
			tok = tr.peek()
			if tok.tokTyp == tokenCloseBracket {
				tr.next() // consume tok
				return nil
			} else if tok.tokTyp.IsSep() {
				tr.next() // consume separator
			}

		}
	}
	return skipFieldElementText(tr)
}

func skipFieldElementText(tr *txtReader) error {
	tok := tr.next()
	switch tok.tokTyp {
	case tokenEOF:
		return io.ErrUnexpectedEOF
	case tokenInt, tokenFloat, tokenString, tokenIdent:
		return nil
	case tokenOpenAngle:
		return skipMessageText(tr, false)
	default:
		return textError(tok, "Expecting an angle bracket '<' or a value; instead got %q", tok.txt)
	}
}

func skipMessageText(tr *txtReader, isGroup bool) error {
	for {
		tok := tr.peek()
		if tok.tokTyp == tokenEOF {
			return io.ErrUnexpectedEOF
		} else if isGroup && tok.tokTyp == tokenCloseBrace {
			return nil
		} else if !isGroup && tok.tokTyp == tokenCloseAngle {
			return nil
		}

		// field name or tag
		if err := skipFieldNameText(tr); err != nil {
			return err
		}

		// field value
		tok = tr.next()
		if tok.tokTyp == tokenEOF {
			return io.ErrUnexpectedEOF
		} else if tok.tokTyp == tokenOpenBrace {
			if err := skipMessageText(tr, true); err != nil {
				return err
			}
		} else if tok.tokTyp == tokenColon {
			if err := skipFieldValueText(tr); err != nil {
				return err
			}
		} else {
			return textError(tok, "Expecting a colon ':' or brace '{'; instead got %q", tok.txt)
		}

		tok = tr.peek()
		if tok.tokTyp.IsSep() {
			tr.next() // consume separator
		}
	}
}

type tokenType int

const (
	tokenError tokenType = iota
	tokenEOF
	tokenIdent
	tokenString
	tokenInt
	tokenFloat
	tokenColon
	tokenComma
	tokenSemiColon
	tokenOpenBrace
	tokenCloseBrace
	tokenOpenBracket
	tokenCloseBracket
	tokenOpenAngle
	tokenCloseAngle
	tokenOpenParen
	tokenCloseParen
)

func (t tokenType) IsSep() bool {
	return t == tokenComma || t == tokenSemiColon
}

func (t tokenType) EndToken() tokenType {
	switch t {
	case tokenOpenAngle:
		return tokenCloseAngle
	case tokenOpenBrace:
		return tokenCloseBrace
	default:
		return tokenError
	}
}

type token struct {
	tokTyp tokenType
	val    interface{}
	txt    string
	pos    scanner.Position
}

type txtReader struct {
	scanner    scanner.Scanner
	peeked     token
	havePeeked bool
	err        error
}

func newReader(text []byte) *txtReader {
	sc := scanner.Scanner{}
	sc.Init(bytes.NewReader(text))
	sc.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanFloats | scanner.ScanChars |
		scanner.ScanStrings | scanner.ScanComments | scanner.SkipComments
	// identifiers are same restrictions as Go identifiers, except we also allow dots since
	// we accept fully-qualified names
	sc.IsIdentRune = func(ch rune, i int) bool {
		return ch == '_' || unicode.IsLetter(ch) ||
			(i > 0 && unicode.IsDigit(ch)) ||
			(i > 0 && ch == '.')
	}
	// ignore errors; we handle them if/when we see malformed tokens
	sc.Error = func(s *scanner.Scanner, msg string) {}
	return &txtReader{scanner: sc}
}

func (p *txtReader) peek() *token {
	if p.havePeeked {
		return &p.peeked
	}
	t := p.scanner.Scan()
	if t == scanner.EOF {
		p.peeked.tokTyp = tokenEOF
		p.peeked.val = nil
		p.peeked.txt = ""
		p.peeked.pos = p.scanner.Position
	} else if err := p.processToken(t, p.scanner.TokenText(), p.scanner.Position); err != nil {
		p.peeked.tokTyp = tokenError
		p.peeked.val = err
	}
	p.havePeeked = true
	return &p.peeked
}

func (p *txtReader) processToken(t rune, text string, pos scanner.Position) error {
	p.peeked.pos = pos
	p.peeked.txt = text
	switch t {
	case scanner.Ident:
		p.peeked.tokTyp = tokenIdent
		p.peeked.val = text
	case scanner.Int:
		p.peeked.tokTyp = tokenInt
		p.peeked.val = text // can't parse the number because we don't know if it's signed or unsigned
	case scanner.Float:
		p.peeked.tokTyp = tokenFloat
		var err error
		if p.peeked.val, err = strconv.ParseFloat(text, 64); err != nil {
			return err
		}
	case scanner.Char, scanner.String:
		p.peeked.tokTyp = tokenString
		var err error
		if p.peeked.val, err = strconv.Unquote(text); err != nil {
			return err
		}
	case '-': // unary minus, for negative ints and floats
		ch := p.scanner.Peek()
		if ch >= '0' && ch <= '9' {
			t := p.scanner.Scan()
			if t == scanner.EOF {
				return io.ErrUnexpectedEOF
			} else if t == scanner.Float {
				p.peeked.tokTyp = tokenFloat
				text += p.scanner.TokenText()
				p.peeked.txt = text
				var err error
				if p.peeked.val, err = strconv.ParseFloat(text, 64); err != nil {
					p.peeked.pos = p.scanner.Position
					return err
				}
			} else if t == scanner.Int {
				p.peeked.tokTyp = tokenInt
				text += p.scanner.TokenText()
				p.peeked.txt = text
				p.peeked.val = text // can't parse the number because we don't know if it's signed or unsigned
			} else {
				p.peeked.pos = p.scanner.Position
				return fmt.Errorf("Expecting an int or float but got %q", p.scanner.TokenText())
			}
		}
	case ':':
		p.peeked.tokTyp = tokenColon
		p.peeked.val = ':'
	case ',':
		p.peeked.tokTyp = tokenComma
		p.peeked.val = ','
	case ';':
		p.peeked.tokTyp = tokenSemiColon
		p.peeked.val = ';'
	case '{':
		p.peeked.tokTyp = tokenOpenBrace
		p.peeked.val = '{'
	case '}':
		p.peeked.tokTyp = tokenCloseBrace
		p.peeked.val = '}'
	case '<':
		p.peeked.tokTyp = tokenOpenAngle
		p.peeked.val = '<'
	case '>':
		p.peeked.tokTyp = tokenCloseAngle
		p.peeked.val = '>'
	case '[':
		p.peeked.tokTyp = tokenOpenBracket
		p.peeked.val = '['
	case ']':
		p.peeked.tokTyp = tokenCloseBracket
		p.peeked.val = ']'
	case '(':
		p.peeked.tokTyp = tokenOpenParen
		p.peeked.val = '('
	case ')':
		p.peeked.tokTyp = tokenCloseParen
		p.peeked.val = ')'
	default:
		return fmt.Errorf("Invalid character: %c", t)
	}
	return nil
}

func (p *txtReader) next() *token {
	t := p.peek()
	if t.tokTyp != tokenEOF && t.tokTyp != tokenError {
		p.havePeeked = false
	}
	return t
}
