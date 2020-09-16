package prototype

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Get returns the name of the getter method for desc.
func Get(desc protoreflect.FieldDescriptor) string {
	camelCasedName := strcase.ToCamel(string(desc.Name()))
	if desc.IsMap() {
		return fmt.Sprintf("get%sMap", camelCasedName)
	}
	if desc.IsList() {
		// repeated scalars
		return fmt.Sprintf("get%sList", camelCasedName)
	}
	// scalars and wrappers
	return fmt.Sprintf("get%s", camelCasedName)
}

// Set returns the name of the setter method for desc.
//
// Panics, if desc is a map, since there are no setters for maps.
func Set(desc protoreflect.FieldDescriptor) string {
	camelCasedName := strcase.ToCamel(string(desc.Name()))
	if desc.IsMap() {
		panic("Setter called on a map field.")
	}
	if desc.IsList() {
		// repeated scalars and repeated wrappers
		return fmt.Sprintf("set%sList", camelCasedName)
	}
	// scalars and wrappers
	return fmt.Sprintf("set%s", camelCasedName)
}

// Clear returns the name of the clearer method for desc.
func Clear(desc protoreflect.FieldDescriptor) string {
	camelCasedName := strcase.ToCamel(string(desc.Name()))
	if desc.IsMap() {
		return fmt.Sprintf("clear%sMap", camelCasedName)
	}
	if desc.IsList() {
		return fmt.Sprintf("clear%sList", camelCasedName)
	}
	// scalars
	return fmt.Sprintf("clear%s", camelCasedName)
}

// Has returns the name of the has method for desc.
//
// Panics, if desc is not inside a oneof, since has method are only intended for
// fields that are inside a oneof.
func Has(desc protoreflect.FieldDescriptor) string {
	of := desc.ContainingOneof()
	if of == nil {
		panic("not a oneof")
	}
	camelCasedName := strcase.ToCamel(string(desc.Name()))
	return fmt.Sprintf("has%s", camelCasedName)
}

// Add returns the name of the adder method for desc.
//
// Panics, if desc is not a list, since adder methods are only intended for list
// fields
func Add(desc protoreflect.FieldDescriptor) string {
	if !desc.IsList() {
		panic("not a list")
	}
	// repeated scalars
	camelCasedName := strcase.ToCamel(string(desc.Name()))
	return fmt.Sprintf("addTo%s", camelCasedName)
}

// Type returns the TypeScript type that corresponds to desc.Kind().
//
// For Enums and Message kinds (including maps), it returns the Enum respec.
// Message name in context. For repeated fields (excluding maps), it returns the
// Javascript type of the elements.
//
// - If desc is of BoolKind, "boolean" is returned
// - If desc is of BytesKinde, "Uint8Array | string" is returned
// - If desc is of any of the numeric types, "number" is returned
// - If desc is of StringKind, "string" is returned
// - If desc is of EnumKind, the Enum name (in context) is returned
// - If desc is of MessageKind, the Message name (in context) is returned
func Type(desc protoreflect.FieldDescriptor) string {
	// Not for maps or repeated types
	switch desc.Kind() {
	case protoreflect.BoolKind:
		return "boolean"
	case protoreflect.BytesKind:
		return "Uint8Array | string"
	case protoreflect.DoubleKind,
		protoreflect.Fixed32Kind,
		protoreflect.Fixed64Kind,
		protoreflect.FloatKind,
		protoreflect.Int32Kind,
		protoreflect.Int64Kind,
		protoreflect.Sfixed32Kind,
		protoreflect.Sfixed64Kind,
		protoreflect.Sint32Kind,
		protoreflect.Sint64Kind,
		protoreflect.Uint32Kind,
		protoreflect.Uint64Kind:
		return "number"
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return nameInContext(desc.ParentFile(), desc.Message())
	case protoreflect.EnumKind:
		return enumNameInContext(desc.ParentFile(), desc.Enum())
	case protoreflect.StringKind:
		return "string"
	default:
		panic(fmt.Sprintf("unrecognized kind: %v", desc.Kind()))
	}
}

// Default returns the default value for with respect to desc.Kind().
//
// For repeated fields, it returns the default value of the element types. It
// panics if desc is of MessageKind.
//
// - If desc is of BoolKind, "false" is returned
// - If desc is of BytesKind, '""' is retured
// - If desc is of any of the numeric types, "0" is returned
// - If desc is of EnumKind, "0" is returned
// - If desc is of String, '""' is retured
func Default(desc protoreflect.FieldDescriptor) string {
	switch desc.Kind() {
	case protoreflect.BoolKind:
		return "false"
	case protoreflect.BytesKind:
		return "\"\""
	case protoreflect.DoubleKind,
		protoreflect.Fixed32Kind,
		protoreflect.Fixed64Kind,
		protoreflect.FloatKind,
		protoreflect.Int32Kind,
		protoreflect.Int64Kind,
		protoreflect.Sfixed32Kind,
		protoreflect.Sfixed64Kind,
		protoreflect.Sint32Kind,
		protoreflect.Sint64Kind,
		protoreflect.Uint32Kind,
		protoreflect.Uint64Kind:
		return "0"
	case protoreflect.EnumKind:
		return "0"
	case protoreflect.StringKind:
		return "\"\""
	default:
		panic(fmt.Sprintf("Default called on kind: %v", desc.Kind()))
	}
}

// NormaliseFieldName modifies the field name n to match the logic found in
// protobuf/compiler/js/js_generator.cc`. See: https://goo.gl/tX1dPQ.
func NormalizedFieldName(n string) string {
	switch n {
	case "abstract",
		"boolean",
		"break",
		"byte",
		"case",
		"catch",
		"char",
		"class",
		"const",
		"continue",
		"debugger",
		"default",
		"delete",
		"do",
		"double",
		"else",
		"enum",
		"export",
		"extends",
		"false",
		"final",
		"finally",
		"float",
		"for",
		"function",
		"goto",
		"if",
		"implements",
		"import",
		"in",
		"instanceof",
		"int",
		"interface",
		"long",
		"native",
		"new",
		"null",
		"package",
		"private",
		"protected",
		"public",
		"return",
		"short",
		"static",
		"super",
		"switch",
		"synchronized",
		"this",
		"throw",
		"throws",
		"transient",
		"try",
		"typeof",
		"var",
		"void",
		"volatile",
		"while",
		"with":
		return "pb_" + n
	default:
		return n
	}
}

// Ctor returns the name of the constructor for desc.,
//
// Some methods of the google-protobuf npm package require a ctor as argument.
// e.g.
//
// 	jspb.Message.getMapField(
// 		msg: Message,
// 		fieldNumber: number,
// 		noLazyCreate: boolean,
// 		valueCtor: typeof Message | undefined): Map<any, any>;)
//
// For repeated fields, the constructor for element types is returned. For
// messages and enums, the Enum respec. Message name (in context) is returned.
// For primitive types, "undefined" is returned.
//
// Panics, if desc is a map, since the for maps only the Ctor of the key or
// value type is ever needed.
func Ctor(desc protoreflect.FieldDescriptor) string {
	if desc.IsMap() {
		panic("Ctor called of a map field.")
	}
	if desc.Kind() == protoreflect.MessageKind {
		return nameInContext(desc.ParentFile(), desc.Message())
	}
	return "undefined"
}

// BinaryReaderFunc returns the name of the function that should be called on
// the jspb.BinaryReader class to read the field described by desc.
func BinaryReaderFunc(desc protoreflect.FieldDescriptor) string {
	packed := ""
	if desc.IsPacked() {
		packed = "Packed"
	}
	switch desc.Kind() {
	case protoreflect.BoolKind:
		return "read" + packed + "Bool"
	case protoreflect.BytesKind:
		// readBytes is ised for repeated and non-repeated bytes
		return "readBytes"
	case protoreflect.EnumKind:
		return "read" + packed + "Enum"
	case protoreflect.DoubleKind:
		return "read" + packed + "Double"
	case protoreflect.Int32Kind:
		return "read" + packed + "Int32"
	case protoreflect.Int64Kind:
		return "read" + packed + "Int64"
	case protoreflect.Uint32Kind:
		return "read" + packed + "Uint32"
	case protoreflect.Uint64Kind:
		return "read" + packed + "Uint64"
	case protoreflect.Sint32Kind:
		return "read" + packed + "Sint32"
	case protoreflect.Sint64Kind:
		return "read" + packed + "Sint64"
	case protoreflect.Fixed32Kind:
		return "read" + packed + "Fixed32"
	case protoreflect.Fixed64Kind:
		return "read" + packed + "Fixed64"
	case protoreflect.FloatKind:
		return "read" + packed + "Float"
	case protoreflect.StringKind:
		// readString is used for repeated and non-repeated strings
		return "readString"
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return "readMessage"
	default:
		panic(fmt.Sprintf("BinaryReaderFunc called with Kind: %v", desc.Kind()))
	}
}

// BinaryWriterFunc returns the name of the function that should be called on
// the BinaryWriter class to write the field described by desc.
func BinaryWriterFunc(desc protoreflect.FieldDescriptor) string {
	packed := ""
	if desc.IsPacked() {
		packed = "Packed"
	} else if desc.Cardinality() == protoreflect.Repeated {
		packed = "Repeated"
	}
	switch desc.Kind() {
	case protoreflect.BoolKind:
		return "write" + packed + "Bool"
	case protoreflect.BytesKind:
		// readBytes is ised for repeated and non-repeated bytes
		return "write" + packed + "Bytes"
	case protoreflect.EnumKind:
		return "write" + packed + "Enum"
	case protoreflect.DoubleKind:
		return "write" + packed + "Double"
	case protoreflect.Int32Kind:
		return "write" + packed + "Int32"
	case protoreflect.Int64Kind:
		return "write" + packed + "Int64"
	case protoreflect.Uint32Kind:
		return "write" + packed + "Uint32"
	case protoreflect.Uint64Kind:
		return "write" + packed + "Uint64"
	case protoreflect.Sint32Kind:
		return "write" + packed + "Sint32"
	case protoreflect.Sint64Kind:
		return "write" + packed + "Sint64"
	case protoreflect.Fixed32Kind:
		return "write" + packed + "Fixed32"
	case protoreflect.Fixed64Kind:
		return "write" + packed + "Fixed64"
	case protoreflect.FloatKind:
		return "write" + packed + "Float"
	case protoreflect.StringKind:
		// readString is used for repeated and non-repeated strings
		return "write" + packed + "String"
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return "write" + packed + "Message"
	default:
		panic(fmt.Sprintf("BinaryWriterFunc called with Kind: %v", desc.Kind()))
	}
}
