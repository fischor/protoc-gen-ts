package main

import (
	"fmt"
	"strings"

	"github.com/fischor/protoc-gen-ts/internal/prototype"
	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func generateFile(gen *protogen.Plugin, file *protogen.File, params parameter) *protogen.GeneratedFile {
	if params.WellKnownPath != "" {
		prototype.WellKnownPath = params.WellKnownPath
	}

	g := gen.NewGeneratedFile(prototype.Path(file.Desc), file.GoImportPath)
	p := newPrinter(g)

	p.P("// Code generated by protoc-gen-ts. DO NOT EDIT.")
	p.P("// versions:")
	p.P("// 	protoc-gen-go ", "v0.0.1-devel")
	suffix := ""
	if *gen.Request.CompilerVersion.Suffix != "" {
		suffix = "-" + *gen.Request.CompilerVersion.Suffix
	}
	p.P("// 	protoc ",
		*gen.Request.CompilerVersion.Major, ".",
		*gen.Request.CompilerVersion.Minor, ".",
		*gen.Request.CompilerVersion.Patch,
		suffix)
	p.P("// source: ", file.Desc.Path())
	p.P()

	genImports(gen, file, p)
	p.P()

	for _, enum := range file.Enums {
		genEnum(gen, file, p, enum)
		p.P()
	}

	for _, msg := range file.Messages {
		genMessage(gen, file, p, msg)
		p.P()
	}

	for _, svc := range file.Services {
		genService(gen, file, p, svc)
		p.P()
	}

	for _, ext := range file.Extensions {
		genExtension(gen, file, p, ext)
		p.P()
	}

	return g
}

func genImports(gen *protogen.Plugin, file *protogen.File, g *Printer) {
	if len(file.Messages) > 0 || len(file.Extensions) > 0 {
		g.P("import jspb from \"google-protobuf\";")

	}
	if len(file.Services) > 0 {
		g.P("import grpcweb from \"grpc-web\";")
	}
	// TODO: rename inmport for file to Imports make imp.Alias and imp.Path
	// fields instead of methods.
	imps := prototype.Imports(file.Desc)
	for _, imp := range imps {
		g.P("import * as ", imp.Alias, " from \"", imp.Path, "\";")
	}
}

// extend google.protobuf.MessageOptions {
//   string resource_name = 8872138;
//   string resource_type = 8872139;
// }
//
// export class ExtensionFieldInfo<T> {
//   fieldIndex: number;
//   fieldName: number;
//   ctor: typeof Message;
//   toObjectFn: Message.StaticToObject;
//   isRepeated: number;
//   constructor(
//     fieldIndex: number,
//     fieldName: {[key: string]: number},
//     ctor: typeof Message,
//     toObjectFn: Message.StaticToObject,
//     isRepeated: number);
//   isMessageType(): boolean;
// }
//
//
//
//
// /**
//  * A tuple of {field number, class constructor} for the extension
//  * field named `http`.
//  * @type {!jspb.ExtensionFieldInfo<!proto.google.api.HttpRule>}
//  */
// proto.google.api.http = new jspb.ExtensionFieldInfo(
//     72295728,
//     {http: 0},
//     google_api_http_pb.HttpRule,
//      /** @type {?function((boolean|undefined),!jspb.Message=): !Object} */ (
//          google_api_http_pb.HttpRule.toObject),
//     0);
//
// google_protobuf_descriptor_pb.MethodOptions.extensionsBinary[72295728] = new jspb.ExtensionFieldBinaryInfo(
//     proto.google.api.http,
//     jspb.BinaryReader.prototype.readMessage,
//     jspb.BinaryWriter.prototype.writeMessage,
//     google_api_http_pb.HttpRule.serializeBinaryToWriter,
//     google_api_http_pb.HttpRule.deserializeBinaryFromReader,
//     false);
// // This registers the extension field with the extended class, so that
// // toObject() will function correctly.
// google_protobuf_descriptor_pb.MethodOptions.extensions[72295728] = proto.google.api.http;
func genExtension(gen *protogen.Plugin, file *protogen.File, p *Printer, extension *protogen.Extension) {
	// Note that map fields are not allowed to be extensions.

	// Generate the jspb.ExtensionsFieldInfo for the extensions.
	//
	// 	export class ExtensionFieldInfo<T> {
	// 	  constructor(
	// 	    fieldIndex: number,
	// 	    fieldName: {[key: string]: number},
	// 	    ctor: typeof Message,
	// 	    toObjectFn: Message.StaticToObject,
	// 	    isRepeated: number);
	// 	}
	extensionFieldInfo := fmt.Sprint("ExtensionFieldInfo_", extension.Extendee.GoIdent.GoName, "_", extension.GoName)
	p.P("export const ", extensionFieldInfo, " = new jspb.ExtensionFieldInfo(")
	p.Indent()
	p.P(extension.Desc.Number(), ",")
	p.P("{", extension.Desc.JSONName(), ": 0},")
	if extension.Desc.Kind() == protoreflect.MessageKind {
		// For messages, there the ctor is the message itself, the
		// toObjectFn is it toObject funtion.
		p.P(prototype.Type(extension.Desc), ",")
		p.P("// @ts-ignore")
		p.P(prototype.Type(extension.Desc), ".toObject,")
	} else {
		// For scalar, there is no ctor and no toObjectFn, use null
		// instead.
		p.P("null,")
		p.P("null,")
	}
	if extension.Desc.Cardinality() == protoreflect.Repeated {
		// For repeated fields, the isRepeated parameter is 1.
		p.P("1")
	} else {
		// For non-repeated fields, the isRepeated parameter is 0.
		p.P("0")
	}
	p.Outdent()
	p.P(");")
	p.P()

	// Add the extension to the status [File|Message|Method]Option map.
	//
	//
	// export class ExtensionFieldBinaryInfo<T> {
	//   constructor(
	//     fieldInfo: ExtensionFieldInfo<T>,
	//     binaryReaderFn: BinaryRead,
	//     binaryWriterFn: BinaryWrite,
	//     opt_binaryMessageSerializeFn: (msg: Message, writer: BinaryWriter) => void,
	//     opt_binaryMessageDeserializeFn: (msg: Message, reader: BinaryReader) => Message,
	//     opt_isPacked: boolean);
	// }
	optionName := prototype.NameInContext(extension.Desc.ParentFile(), extension.Desc.ContainingMessage())
	p.P(optionName, ".extensionsBinary[", extension.Desc.Number(), "] = new jspb.ExtensionFieldBinaryInfo(")
	p.Indent()
	p.P(extensionFieldInfo, ",")
	p.P("jspb.BinaryReader.prototype.", prototype.BinaryReaderFunc(extension.Desc), ",")
	p.P("jspb.BinaryWriter.prototype.", prototype.BinaryWriterFunc(extension.Desc), ",")
	// opt_binaryMessageSerializeFn and opt_binaryMessageDeserializeFn
	if extension.Desc.Kind() == protoreflect.MessageKind {
		p.P("// @ts-ignore")
		p.P(prototype.Type(extension.Desc), ".serializeBinaryToWriter,")
		p.P(prototype.Type(extension.Desc), ".deserializeBinaryFromReader,")
	} else {
		p.P("undefined,")
		p.P("undefined,")
	}
	// opt_isPacked
	if extension.Desc.IsPacked() {
		p.P("true);")
	} else {
		p.P("false);")
	}
	p.Outdent()
	p.P()

	// Add the extension to the status [File|Message|Method]Option map.
	p.P(optionName, ".extensions[", extension.Desc.Number(), "] =", extensionFieldInfo, ";")
}

func genEnum(gen *protogen.Plugin, file *protogen.File, p *Printer, enum *protogen.Enum) {
	p.P("export enum ", enum.Desc.Name(), " {")
	p.Indent()
	for _, val := range enum.Values {
		p.P(val.Desc.Name(), " = ", val.Desc.Number(), ",")
	}
	p.Outdent()
	p.P("}")
}

func genMessage(gen *protogen.Plugin, file *protogen.File, p *Printer, msg *protogen.Message) {
	// Generate constants for the repeated and oneof field numbers.
	repeatedFields := prototype.RepeatedFields(msg.Desc)
	oneofFields := prototype.OneofFields(msg.Desc)
	p.P("const __", msg.Desc.Name(), "_repeated: number[] = ", toArray(repeatedFields), ";")
	p.P()
	p.P("const __", msg.Desc.Name(), "_oneof: number[][] = ", to2DArray(oneofFields), ";")
	p.P()

	if msg.Comments.Leading != "" {
		p.P("/**")
		p.C(msg.Comments.Leading)
		p.P(" */")
	}
	p.P("export class ", msg.Desc.Name(), " extends jspb.Message {")
	p.P()
	p.Indent()

	// Generate statuc deserializeBinary method.
	p.P("static deserializeBinary(bytes: Uint8Array): ", msg.Desc.Name(), " {")
	p.Indented(func() {
		p.P("let reader = new jspb.BinaryReader(bytes);")
		p.P("let msg = new ", msg.Desc.Name(), "();")
		p.P("return ", msg.Desc.Name(), ".deserializeBinaryFromReader(msg, reader);")
	})
	p.P("}")
	p.P()

	// Generate other static methods.
	genDeserializeBinaryFromReader(gen, file, p, msg)
	p.P()

	genSerializeBinaryToWriter(gen, file, p, msg)
	p.P()

	genToObject(gen, file, p, msg)
	p.P()

	// Generate constructor.
	msgID := 0
	suggestedPivot := -1
	p.P("constructor(data?: jspb.Message.MessageArray) {")
	p.Indented(func() {
		p.P("super();")
		p.F("jspb.Message.initialize(this, data ?? [], %d, %d, __%[3]s_repeated, __%[3]s_oneof);", msgID, suggestedPivot, msg.Desc.Name())
	})
	p.P("}")
	p.P()

	// Generate serializeBinary method.
	p.P("serializeBinary(): Uint8Array {")
	p.Indented(func() {
		p.P("const writer = new jspb.BinaryWriter();")
		p.P(msg.Desc.Name(), ".serializeBinaryToWriter(this, writer);")
		p.P("return writer.getResultBuffer();")
	})
	p.P("}")

	// Generate toObject method
	p.P("toObject(includeInstance?: boolean): ", msg.Desc.Name(), ".AsObject {")
	p.Indented(func() {
		p.P("return ", msg.Desc.Name(), ".toObject(includeInstance ?? false, this);")
	})
	p.P("}")

	// Generate field methods
	for _, field := range msg.Fields {
		genFieldMethods(gen, file, p, field)
	}

	p.Outdent()
	p.P("}") // class end
	p.P()

	genMessageNamespace(gen, file, p, msg)
}

func genMessageNamespace(gen *protogen.Plugin, file *protogen.File, p *Printer, msg *protogen.Message) {
	p.P("/**")
	p.P(" * Namespace for the ", msg.Desc.Name(), ".")
	p.P(" * Contains nested message and enum declarations.")
	p.P(" */")
	p.P("export namespace ", msg.Desc.Name(), "{")
	p.P()
	p.Indented(func() {
		// Generate AsObject tyoe
		p.P("export type AsObject = {")
		p.Indented(func() {
			for i, field := range msg.Fields {
				var fieldType string
				var optional bool
				if field.Desc.IsMap() {
					fieldType = "Array<[" + prototype.Type(field.Desc.MapKey()) + "," + prototype.Type(field.Desc.MapValue()) + "]>"
				} else if field.Desc.IsList() && field.Desc.Kind() == protoreflect.MessageKind {
					fieldType = "Array<" + prototype.Type(field.Desc) + ".AsObject>"
				} else if field.Desc.IsList() {
					fieldType = "Array<" + prototype.Type(field.Desc) + ">"
				} else if field.Desc.Kind() == protoreflect.MessageKind {
					fieldType = prototype.Type(field.Desc) + ".AsObject"
					optional = true
				} else {
					fieldType = prototype.Type(field.Desc)
				}
				suffix := ","
				if i == len(msg.Fields)-1 {
					suffix = ""
				}
				optFlag := ""
				if optional {
					optFlag = "?"
				}
				p.P(prototype.NormalizedFieldName(field.Desc.JSONName()), optFlag, ": ", fieldType, suffix)
			}

		})
		p.P("}")
		p.P()

		// Generate nested enums.
		for _, enum := range msg.Enums {
			genEnum(gen, file, p, enum)
			p.P()
		}

		// Generate nested messages.
		for _, nested := range msg.Messages {
			genMessage(gen, file, p, nested)
			p.P()
		}
	})
	p.P("}") // namespace end
}

func genDeserializeBinaryFromReader(gen *protogen.Plugin, file *protogen.File, p *Printer, msg *protogen.Message) {
	p.P("static deserializeBinaryFromReader(msg: ", msg.Desc.Name(), ", reader: jspb.BinaryReader): ", msg.Desc.Name(), " {")
	p.Indented(func() {
		p.P("while (reader.nextField()) {")
		p.Indented(func() {
			p.P("if (reader.isEndGroup()) {")
			p.Indented(func() {
				p.P("break;")
			})
			p.P("}") // end if
			p.P("let field = reader.getFieldNumber();")
			p.P("switch (field) {")
			// Generate cases
			for _, field := range msg.Fields {
				p.P("case ", field.Desc.Number(), ": {")
				p.Indented(func() {
					genDeserializeBinaryFromReaderCase(p, field)
					p.P("break;")
				})
				p.P("}") // case end
			}
			p.P("default:")
			p.Indented(func() {
				p.P("reader.skipField();")
				p.P("break;")
			})
			p.P("}") // end switch
		})
		p.P("}") // end while
		p.P("return msg;")
	})
	p.P("}")
}

func genDeserializeBinaryFromReaderCase(p *Printer, field *protogen.Field) {
	if field.Desc.IsMap() {
		p.P("let value = msg.", prototype.Get(field.Desc), "();")
		p.P("reader.readMessage(value, (message, reader) =>")
		p.Indented(func() {
			p.P("jspb.Map.deserializeBinary(")
			p.Indented(func() {
				p.P("message,")
				p.P("reader,")
				p.P("jspb.BinaryReader.prototype.", prototype.BinaryReaderFunc(field.Desc.MapKey()), ",")
				p.P("jspb.BinaryReader.prototype.", prototype.BinaryReaderFunc(field.Desc.MapValue()), ",")
				// TODO: comments
				if field.Desc.MapValue().Kind() == protoreflect.MessageKind {
					p.P(prototype.Type(field.Desc.MapValue()), ".deserializeBinaryFromReader,")
				} else {
					p.P("undefined,")
				}
				p.P(prototype.Default(field.Desc.MapKey()), ",")
				// TODO: comments
				if field.Desc.MapValue().Kind() == protoreflect.MessageKind {
					p.P("undefined,")
				} else {
					p.P(prototype.Default(field.Desc.MapValue()), ",")
				}
			})
			p.P(")")
		})
		p.P(");")
		return
	}
	if field.Desc.IsList() {
		if field.Desc.Kind() == protoreflect.MessageKind {
			// Read operator for repeated, non-packed, wrapper fields.
			// Read operation for scalar, wrapper fields.
			p.P("let value = new ", prototype.Type(field.Desc), "();")
			p.P("reader.readMessage(value,", prototype.Type(field.Desc), ".deserializeBinaryFromReader);")
			p.P("msg.", prototype.Add(field.Desc), "(value);")
			return
		}
		if field.Desc.IsPacked() {
			// Read operation for repeated, non-wrapper, packed fields.
			p.P("let value = reader.", prototype.BinaryReaderFunc(field.Desc), "();")
			p.P("msg.", prototype.Set(field.Desc), "(value);")
			return
		}
		// Read operation for repeated, non-wrapper, non-packed fields.
		p.P("let value = reader.", prototype.BinaryReaderFunc(field.Desc), "();")
		p.P("msg.", prototype.Add(field.Desc), "(value);")
		return
	}
	// Scalar fields.
	if field.Desc.Kind() == protoreflect.MessageKind {
		p.P("let value = new ", prototype.Type(field.Desc), "();")
		p.P("reader.readMessage(value,", prototype.Type(field.Desc), ".deserializeBinaryFromReader);")
		p.P("msg.", prototype.Set(field.Desc), "(value);")
		return
	}
	// Read operation for scalar, non-wrapper fields.
	p.P("let value = reader.", prototype.BinaryReaderFunc(field.Desc), "();")
	p.P("msg.", prototype.Set(field.Desc), "(value);")
	return

}

func genSerializeBinaryToWriter(gen *protogen.Plugin, file *protogen.File, p *Printer, msg *protogen.Message) {
	p.P("static serializeBinaryToWriter(message: ", msg.Desc.Name(), ", writer: jspb.BinaryWriter) {")
	p.Indented(func() {
		for _, field := range msg.Fields {
			p.P("let field", field.Desc.Number(), " = message.", prototype.Get(field.Desc), "();")
			p.P("if (", serializeCompare(field.Desc), ") {")
			p.Indented(func() {
				// TODO: maps
				genSerializeToWriterCase(p, field)
			})
			p.P("}")
		}
	})
	p.P("}")
}

func serializeCompare(fd protoreflect.FieldDescriptor) string {
	if fd.IsMap() {
		return fmt.Sprintf("field%[1]d && field%[1]d.getLength() > 0", fd.Number())
	}
	if fd.IsList() {
		return fmt.Sprintf("field%[1]d && field%[1]d.length > 0", fd.Number())
	}

	switch fd.Kind() {
	case protoreflect.BoolKind:
		return fmt.Sprintf("field%d", fd.Number())
	case protoreflect.BytesKind:
		return fmt.Sprintf("field%d.length > 0", fd.Number())
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
		return fmt.Sprintf("field%d !== 0", fd.Number())
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return fmt.Sprintf("field%d != null", fd.Number())
	case protoreflect.EnumKind:
		return fmt.Sprintf("field%d != 0.0", fd.Number())
	case protoreflect.StringKind:
		return fmt.Sprintf("field%d.length > 0", fd.Number())
	default:
		panic(fmt.Sprintf("unrecognized kind: %v", fd.Kind()))
	}
}

// for scalars and packed
//
// 	writer.write..(<n>, var)
//
// for messages
//
// 	writer.writeMessage(<n>, <var>, <Type>);
//
// for repeated messages
//
// 	writer.writeRepeatedMessage(n, var, Type);
func genSerializeToWriterCase(p *Printer, field *protogen.Field) {
	// Three arguments for message required.
	if field.Desc.IsMap() {
		if field.Desc.MapValue().Kind() == protoreflect.MessageKind {
			p.P("field", field.Desc.Number(), ".serializeBinary(")
			p.Indented(func() {
				p.P(field.Desc.Number(), ",")
				p.P("writer, ")
				p.P("jspb.BinaryWriter.prototype.", prototype.BinaryWriterFunc(field.Desc.MapKey()), ",")
				p.P("jspb.BinaryWriter.prototype.writeMessage, ")
				p.P(prototype.Type(field.Desc.MapValue()), ".serializeBinaryToWriter")
			})
			p.P(");")
			return
		}
		p.P("field", field.Desc.Number(), ".serializeBinary(")
		p.Indented(func() {
			p.P(field.Desc.Number(), ",")
			p.P("writer, ")
			p.P("jspb.BinaryWriter.prototype.", prototype.BinaryWriterFunc(field.Desc.MapKey()), ",")
			p.P("jspb.BinaryWriter.prototype.", prototype.BinaryWriterFunc(field.Desc.MapValue()))
		})
		p.P(");")
		return
	}
	if field.Desc.Kind() == protoreflect.MessageKind {
		if field.Desc.IsList() {
			p.P("writer.writeMessage(", field.Desc.Number(), ", field", field.Desc.Number(), " ,", prototype.Type(field.Desc), ".serializeBinaryToWriter);")
			return
		}
		// Non-repeated, wrapper types.
		p.P("writer.writeMessage(", field.Desc.Number(), ", field", field.Desc.Number(), " ,", prototype.Type(field.Desc), ".serializeBinaryToWriter);")
		return
	}
	// This is for repeated and non-repeated non-wrapper values alike.
	// Note that,in proto3 repeated non-wrapper fields are packed by default,
	// thus BinaryWriterFunc returns the packed version of the write, that
	// accepts a list of values.
	p.P("writer.", prototype.BinaryWriterFunc(field.Desc), "(", field.Desc.Number(), ", field", field.Desc.Number(), ");")
	return
}

// genToObject generates the static toObject method for msg.
func genToObject(gen *protogen.Plugin, file *protogen.File, p *Printer, msg *protogen.Message) {
	p.P("static toObject(includeInstance: boolean, msg: ", msg.Desc.Name(), "): ", msg.Desc.Name(), ".AsObject {")
	p.Indented(func() {
		p.P("return {")
		for i, field := range msg.Fields {
			// All lines expect the last must be suffix with a ","
			isLast := len(msg.Fields) == i-1
			suffix := ","
			if isLast {
				suffix = ""
			}

			// Op that
			getter := prototype.Get(field.Desc)
			op := fmt.Sprintf("msg.%s()", getter)
			if field.Desc.IsMap() {
				op = fmt.Sprintf("msg.%[1]s()?.toObject(includeInstance ?? false) ?? []", getter)
			} else if field.Desc.IsList() && field.Desc.Kind() == protoreflect.MessageKind {
				op = fmt.Sprintf("jspb.Message.toObjectList(msg.%s(), %s.toObject, includeInstance)", getter, prototype.Ctor(field.Desc))
			} else if field.Desc.IsList() && field.Desc.Kind() != protoreflect.MessageKind {
				op = fmt.Sprintf("msg.%s()", getter)
			} else if field.Desc.Kind() == protoreflect.MessageKind {
				op = fmt.Sprintf("msg.%s()?.toObject(includeInstance ?? false)", getter)
			} else {
				// primitive type
				op = fmt.Sprintf("msg.%s()", getter)
			}
			p.P(prototype.NormalizedFieldName(field.Desc.JSONName()), ": ", op, suffix)
		}
		p.P("}")
	})
	p.P("}")
}

// TODO docs:
//
// Need to differentiate between
// - oneof, wrapper fields
// - oneof, scalar field (note there are no maps or repeated fields in oneofs)
// - map fields
// - repated wrapper fields
// - wrapper fields
// - repeated fields
// - scalar field (non-wrapper)
func genFieldMethods(gen *protogen.Plugin, file *protogen.File, p *Printer, field *protogen.Field) {
	if field.Oneof != nil {
		if field.Desc.Kind() == protoreflect.MessageKind {
			// Get for wrapper, oneof fields.
			p.P(prototype.Get(field.Desc), "(): ", prototype.Type(field.Desc), " | undefined {")
			p.Indented(func() {
				p.F("return jspb.Message.getFieldWithDefault(this, %d, undefined) as %s | undefined;", field.Desc.Number(), prototype.Type(field.Desc))
			})
			p.P("}")
			p.P()
		} else {
			// Get for non-wrapper, oneof fields.
			p.P(prototype.Get(field.Desc), "(): ", prototype.Type(field.Desc), "{")
			p.Indented(func() {
				p.F("return jspb.Message.getFieldWithDefault(this, %d, %s);", field.Desc.Number(), prototype.Default(field.Desc))
			})
			p.P("}")
			p.P()
		}
		p.P(prototype.Has(field.Desc), "(): boolean {")
		p.Indented(func() {
			p.F("return jspb.Message.getField(this, %d) !== undefined;", field.Desc.Number())
		})
		p.P("}")
		p.P(prototype.Set(field.Desc), "(value: ", prototype.Type(field.Desc), "): ", field.Parent.Desc.Name(), " {")
		p.Indented(func() {
			if field.Desc.Kind() == protoreflect.MessageKind {
				p.F("jspb.Message.setOneofWrapperField(this, %d, __%s_oneof[%d], value);", field.Desc.Number(), field.Parent.Desc.Name(), field.Desc.ContainingOneof().Index())
			} else {
				p.F("jspb.Message.setOneofField(this, %d, __%s_oneof[%d], value);", field.Desc.Number(), field.Parent.Desc.Name(), field.Desc.ContainingOneof().Index())
			}
			p.P("return this;")
		})
		p.P("}")
		p.P()
		p.P(prototype.Clear(field.Desc), "(): ", field.Parent.Desc.Name(), " {")
		p.Indented(func() {
			p.F("jspb.Message.setOneofField(this, %d, __%s_oneof[%d], undefined);", field.Desc.Number(), field.Parent.Desc.Name(), field.Desc.ContainingOneof().Index())
			p.P("return this;")
		})
		p.P("}")
		p.P()
		return
	}
	if field.Desc.IsMap() {
		// map field
		p.P(prototype.Get(field.Desc), "(): jspb.Map<", prototype.Type(field.Desc.MapKey()), ", ", prototype.Type(field.Desc.MapValue()), "> {")
		p.Indented(func() {
			// Note that getMapField returns undefined only, if noLazyCreate is true.
			p.P("// @ts-ignore: Ignore that getMapField might return undefined")
			p.F("return jspb.Message.getMapField(this, %d, false, %s);", field.Desc.Number(), prototype.Ctor(field.Desc.MapValue()))
		})
		p.P("}")
		p.P()
		p.P(prototype.Clear(field.Desc), "(): ", field.Parent.Desc.Name(), " {")
		p.Indented(func() {
			p.P("this.", prototype.Get(field.Desc), "().clear();")
			p.P("return this;")
		})
		p.P("}")
		p.P()
		return
	}
	if field.Desc.IsList() && field.Desc.Kind() == protoreflect.MessageKind {
		// Generate getter, setter and clearer for repeated, wrapper fields.
		// repeated, non-wrapper field
		p.P(prototype.Get(field.Desc), "(): Array<", prototype.Type(field.Desc), "> {")
		p.Indented(func() {
			p.F("return jspb.Message.getRepeatedWrapperField(this, %s, %d);", prototype.Ctor(field.Desc), field.Desc.Number())
		})
		p.P("}")
		p.P()
		p.P(prototype.Set(field.Desc), "(value: Array<", prototype.Type(field.Desc), ">): ", field.Parent.Desc.Name(), " {")
		p.Indented(func() {
			p.F("jspb.Message.setRepeatedWrapperField(this, %d, value);", field.Desc.Number())
			p.P("return this;")
		})
		p.P("}")
		p.P()
		p.P(prototype.Add(field.Desc), "(value: ", prototype.Type(field.Desc), ", index?: number): ", field.Parent.Desc.Name(), "{")
		p.Indented(func() {
			p.F("jspb.Message.addToRepeatedWrapperField(this, %d, value, %s, index);", field.Desc.Number(), prototype.Ctor(field.Desc))
			p.P("return this;")
		})
		p.P("}")
		p.P()
		p.P(prototype.Clear(field.Desc), "(): ", field.Parent.Desc.Name(), " {")
		p.Indented(func() {
			p.F("jspb.Message.setRepeatedWrapperField(this, %d, undefined);", field.Desc.Number())
			p.P("return this;")
		})
		p.P("}")
		p.P()
		return
	}
	if field.Desc.IsList() {
		// repeated, non-wrapper field
		p.P(prototype.Get(field.Desc), "(): Array<", prototype.Type(field.Desc), "> {")
		p.Indented(func() {
			p.F("return jspb.Message.getField(this, %d) as Array<%s>;", field.Desc.Number(), prototype.Type(field.Desc))
		})
		p.P("}")
		p.P()
		p.P(prototype.Set(field.Desc), "(value: Array<", prototype.Type(field.Desc), ">): ", field.Parent.Desc.Name(), " {")
		p.Indented(func() {
			p.F("jspb.Message.setField(this, %d, value);", field.Desc.Number())
			p.P("return this;")
		})
		p.P("}")
		p.P()
		p.P(prototype.Add(field.Desc), "(value: ", prototype.Type(field.Desc), ", index?: number): ", field.Parent.Desc.Name(), "{")
		p.Indented(func() {
			p.F("jspb.Message.addToRepeatedField(this, %d, value, index);", field.Desc.Number())
			p.P("return this;")

		})
		p.P("}")
		p.P()
		p.P(prototype.Clear(field.Desc), "(): ", field.Parent.Desc.Name(), " {")
		p.Indented(func() {
			p.F("jspb.Message.setField(this, %d, undefined);", field.Desc.Number())
			p.P("return this;")
		})
		p.P("}")
		p.P()
		return
	}
	if field.Desc.Kind() == protoreflect.MessageKind {
		// non-repeated, wrapper field
		p.P(prototype.Get(field.Desc), "(): ", prototype.Type(field.Desc), " {")
		p.Indented(func() {
			p.F("return jspb.Message.getWrapperField(this, %s, %d);", prototype.Type(field.Desc), field.Desc.Number())
		})
		p.P("}")
		p.P()
		p.P(prototype.Set(field.Desc), "(value: ", prototype.Type(field.Desc), "): ", field.Parent.Desc.Name(), " {")
		p.Indented(func() {
			p.F("jspb.Message.setWrapperField(this, %d, value);", field.Desc.Number())
			p.P("return this;")
		})
		p.P("}")
		p.P()
		p.P(prototype.Clear(field.Desc), "(): ", field.Parent.Desc.Name(), " {")
		p.Indented(func() {
			p.F("jspb.Message.setField(this, %d, undefined);", field.Desc.Number())
			p.P("return this;")
		})
		p.P("}")
		p.P()
		return
	}
	// non-repeated, non-wrapper field
	p.P(prototype.Get(field.Desc), "(): ", prototype.Type(field.Desc), "{")
	p.Indented(func() {
		p.F("return jspb.Message.getFieldWithDefault(this, %d, %s);", field.Desc.Number(), prototype.Default(field.Desc))
	})
	p.P("}")
	p.P()
	p.P(prototype.Set(field.Desc), "(value: ", prototype.Type(field.Desc), "): ", field.Parent.Desc.Name(), " {")
	p.Indented(func() {
		p.F("jspb.Message.setField(this, %d, value);", field.Desc.Number())
		p.P("return this;")
	})
	p.P("}")
	p.P()
	p.P(prototype.Clear(field.Desc), "(): ", field.Parent.Desc.Name(), " {")
	p.Indented(func() {
		p.F("jspb.Message.setField(this, %d, undefined);", field.Desc.Number())
		p.P("return this;")
	})
	p.P("}")
	p.P()
}

func genService(gen *protogen.Plugin, file *protogen.File, p *Printer, svc *protogen.Service) {
	p.P("export class ", svc.Desc.Name(), "Client {")
	p.P()
	p.Indent()
	p.P("private client: grpcweb.GrpcWebClientBase;")
	p.P("private hostname: string;")
	p.P()

	// Constructor
	p.P("constructor(hostname: string, options: grpcweb.GrpcWebClientBaseOptions) {")
	p.Indented(func() {
		p.P("this.hostname = hostname;")
		p.P("this.client = new grpcweb.GrpcWebClientBase(options);")
	})
	p.P("}") // constructor end
	p.P()

	// Generate method definitions.
	for _, method := range svc.Methods {
		p.P(strcase.ToLowerCamel(string(method.Desc.Name())), "(")
		p.Indented(func() {
			p.P("request: ", prototype.InputType(method.Desc), ",")
			p.P("metadata: grpcweb.Metadata,")
			p.P("callback: (err: grpcweb.Error, response: ", prototype.OutputType(method.Desc), ") => void")
		})
		p.P("): grpcweb.ClientReadableStream<", prototype.OutputType(method.Desc), "> {")
		p.Indented(func() {
			p.P("return this.client.rpcCall(")
			p.Indented(func() {
				p.P("this.hostname + \"", prototype.Address(method.Desc), "\",")
				p.P("request,")
				p.P("metadata,")
				p.P(methodDescriptorName(method), ",")
				p.P("callback")
			})
			p.P(")")
		})
		p.P("}") // method end
		p.P()
	}

	p.Outdent()
	p.P("}") // service class end

	// Generate method descriptor and info.
	for _, method := range svc.Methods {
		p.P("const ", methodDescriptorName(method), " = new grpcweb.MethodDescriptor<")
		p.Indented(func() {
			p.P(prototype.InputType(method.Desc), ", ")
			p.P(prototype.OutputType(method.Desc))
		})
		p.P(">(")
		p.Indented(func() {
			p.F("\"%s\",", prototype.Address(method.Desc))
			p.F("\"%s\",", prototype.MethodType(method.Desc))
			p.P(prototype.InputType(method.Desc), ",")
			p.P(prototype.OutputType(method.Desc), ",")
			p.P("(req: ", prototype.InputType(method.Desc), ") => req.serializeBinary(),")
			p.P(prototype.OutputType(method.Desc), ".deserializeBinary")
		})
		p.P(");")
		p.P()
		// Generate method info.
		p.P("const ", methodInfoName(method), " = new grpcweb.AbstractClientBase.MethodInfo<")
		p.Indented(func() {
			p.P(prototype.InputType(method.Desc), ", ")
			p.P(prototype.OutputType(method.Desc))
		})
		p.P(">(")
		p.Indented(func() {
			p.P(prototype.OutputType(method.Desc), ",")
			p.P("(req: ", prototype.InputType(method.Desc), ") => req.serializeBinary(),")
			p.P(prototype.OutputType(method.Desc), ".deserializeBinary")
		})
		p.P(");")
		p.P()
	}
}

func methodInfoName(m *protogen.Method) string {
	return "methodInfo_" + string(m.Desc.Parent().Name()) + "_" + string(m.Desc.Name())
}

func methodDescriptorName(m *protogen.Method) string {
	return "methodDescriptor_" + string(m.Desc.Parent().Name()) + "_" + string(m.Desc.Name())
}

func toArray(arr []int32) string {
	var ss []string
	for _, a := range arr {
		ss = append(ss, fmt.Sprintf("%d", a))
	}
	return "[" + strings.Join(ss, ",") + "]"
}

func to2DArray(arr [][]int32) string {
	var ss []string
	for _, a := range arr {
		ss = append(ss, toArray(a))
	}
	return "[" + strings.Join(ss, ",") + "]"
}
