package internal

import (
	"unicode"
	"unicode/utf8"

	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
)

const (
	MaxTag = 536870911 // 2^29 - 1

	SpecialReservedStart = 19000
	SpecialReservedEnd   = 19999

	// NB: It would be nice to use constants from generated code instead of hard-coding these here.
	// But code-gen does not emit these as constants anywhere. The only places they appear in generated
	// code are struct tags on fields of the generated descriptor protos.
	File_packageTag           = 2
	File_dependencyTag        = 3
	File_messagesTag          = 4
	File_enumsTag             = 5
	File_servicesTag          = 6
	File_extensionsTag        = 7
	File_optionsTag           = 8
	File_syntaxTag            = 12
	Message_nameTag           = 1
	Message_fieldsTag         = 2
	Message_nestedMessagesTag = 3
	Message_enumsTag          = 4
	Message_extensionRangeTag = 5
	Message_extensionsTag     = 6
	Message_optionsTag        = 7
	Message_oneOfsTag         = 8
	Message_reservedRangeTag  = 9
	Message_reservedNameTag   = 10
	ExtensionRange_startTag   = 1
	ExtensionRange_endTag     = 2
	ExtensionRange_optionsTag = 3
	ReservedRange_startTag    = 1
	ReservedRange_endTag      = 2
	Field_nameTag             = 1
	Field_extendeeTag         = 2
	Field_numberTag           = 3
	Field_labelTag            = 4
	Field_typeTag             = 5
	Field_defaultTag          = 7
	Field_optionsTag          = 8
	Field_jsonNameTag         = 10
	OneOf_nameTag             = 1
	OneOf_optionsTag          = 2
	Enum_nameTag              = 1
	Enum_valuesTag            = 2
	Enum_optionsTag           = 3
	EnumVal_nameTag           = 1
	EnumVal_numberTag         = 2
	EnumVal_optionsTag        = 3
	Service_nameTag           = 1
	Service_methodsTag        = 2
	Service_optionsTag        = 3
	Method_nameTag            = 1
	Method_inputTag           = 2
	Method_outputTag          = 3
	Method_optionsTag         = 4
	Method_inputStreamTag     = 5
	Method_outputStreamTag    = 6

	// All *Options messages use the same tag for the field
	// that stores uninterpreted options
	UninterpretedOptionsTag = 999

	Uninterpreted_nameTag      = 2
	Uninterpreted_identTag     = 3
	Uninterpreted_posIntTag    = 4
	Uninterpreted_negIntTag    = 5
	Uninterpreted_doubleTag    = 6
	Uninterpreted_stringTag    = 7
	Uninterpreted_aggregateTag = 8
	UninterpretedName_nameTag  = 1
)

func JsonName(name string) string {
	var js []rune
	nextUpper := false
	for i, r := range name {
		if r == '_' {
			nextUpper = true
			continue
		}
		if i == 0 {
			js = append(js, r)
		} else if nextUpper {
			nextUpper = false
			js = append(js, unicode.ToUpper(r))
		} else {
			js = append(js, r)
		}
	}
	return string(js)
}

func InitCap(name string) string {
	r, sz := utf8.DecodeRuneInString(name)
	return string(unicode.ToUpper(r)) + name[sz:]
}

type SourceInfoMap map[string]*dpb.SourceCodeInfo_Location

func (m SourceInfoMap) Get(path []int32) *dpb.SourceCodeInfo_Location {
	return m[asMapKey(path)]
}

func (m SourceInfoMap) Put(path []int32, loc *dpb.SourceCodeInfo_Location) {
	m[asMapKey(path)] = loc
}

func (m SourceInfoMap) PutIfAbsent(path []int32, loc *dpb.SourceCodeInfo_Location) bool {
	k := asMapKey(path)
	if _, ok := m[k]; ok {
		return false
	}
	m[k] = loc
	return true
}

func asMapKey(slice []int32) string {
	// NB: arrays should be usable as map keys, but this does not
	// work due to a bug: https://github.com/golang/go/issues/22605
	//rv := reflect.ValueOf(slice)
	//arrayType := reflect.ArrayOf(rv.Len(), rv.Type().Elem())
	//array := reflect.New(arrayType).Elem()
	//reflect.Copy(array, rv)
	//return array.Interface()

	b := make([]byte, len(slice)*4)
	for i, s := range slice {
		j := i * 4
		b[j] = byte(s)
		b[j+1] = byte(s >> 8)
		b[j+2] = byte(s >> 16)
		b[j+3] = byte(s >> 24)
	}
	return string(b)
}

func CreateSourceInfoMap(fd *dpb.FileDescriptorProto) SourceInfoMap {
	res := SourceInfoMap{}
	for _, l := range fd.GetSourceCodeInfo().GetLocation() {
		res.Put(l.Path, l)
	}
	return res
}

// CreatePrefixList returns a list of package prefixes to search when resolving
// a symbol name. If the given package is blank, it returns only the empty
// string. If the given package contains only one token, e.g. "foo", it returns
// that token and the empty string, e.g. ["foo", ""]. Otherwise, it returns
// successively shorter prefixes of the package and then the empty string. For
// example, for a package named "foo.bar.baz" it will return the following list:
//   ["foo.bar.baz", "foo.bar", "foo", ""]
func CreatePrefixList(pkg string) []string {
	if pkg == "" {
		return []string{""}
	}

	numDots := 0
	// one pass to pre-allocate the returned slice
	for i := 0; i < len(pkg); i++ {
		if pkg[i] == '.' {
			numDots++
		}
	}
	if numDots == 0 {
		return []string{pkg, ""}
	}

	prefixes := make([]string, numDots+2)
	// second pass to fill in returned slice
	for i := 0; i < len(pkg); i++ {
		if pkg[i] == '.' {
			prefixes[numDots] = pkg[:i]
			numDots--
		}
	}
	prefixes[0] = pkg

	return prefixes
}
