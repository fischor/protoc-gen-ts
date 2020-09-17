package prototype

import (
	"path"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// Import is a typescript import.
type Import struct {
	// The Typescript import path.
	//
	// For well known types, eg.: "google/protobuf/timestamp.proto", this
	// would be
	//
	// 	"google-protobuf/google/protobuf/timestamp_pb"
	//
	// For local, e.g. a file "mycom/protobuf/hello.proto", this would be
	//
	// 	"./mycom/protobuf/hello_pb"
	//
	Path string

	// Alias is the alias name of the Import.
	//
	// It concats the package name with the filename. E.g.: for a proto
	// import:
	//
	// 	import "google/protobuf/timestamp.proto"
	//
	// Alias would be "google_protobuf_timestamp_pb".
	Alias string
}

// Imports returns the list of Import's that are present in desc.
func Imports(desc protoreflect.FileDescriptor) []*Import {
	var ii []*Import
	for i := 0; i < desc.Imports().Len(); i++ {
		ii = append(ii, newImport(desc, desc.Imports().Get(i)))
	}
	return ii
}

func newImport(base protoreflect.FileDescriptor, imp protoreflect.FileImport) *Import {
	return &Import{
		Path:  importPath(base, imp),
		Alias: importAlias(imp),
	}
}

// Path returns the path of the typescript fule that protoc-gen-ts would
// generate for desc.
//
// It replaces ".proto" suffix, if present, with "_pb.ts". If no ".proto" suffix
// is present, path is returned.
func Path(desc protoreflect.FileDescriptor) string {
	name := strings.TrimSuffix(desc.Path(), ".proto") + "_pb.ts"
	return name
}

// importPath constructs the Typescript import path.
//
// For well known types, eg.: "google/protobuf/timestamp.proto", importPath
// returns
//
// 	"google-protobuf/google/protobuf/timestamp_pb"
//
// For local, e.g. a file "mycom/protobuf/hello.proto", tmporPath returns
//
// 	"./mycom/protobuf/hello_pb"
//
func importPath(desc protoreflect.FileDescriptor, imp protoreflect.FileImport) string {
	if imp.Package() == "google.protobuf" {
		return "google-protobuf/" + strings.TrimSuffix(imp.Path(), ".proto") + "_pb"
	}
	base := path.Dir(desc.Path())
	relpath, err := filepath.Rel(base, imp.Path())
	if err != nil {
		panic(err)
	}
	path := strings.TrimSuffix(relpath, ".proto") + "_pb"
	if !strings.HasPrefix(path, ".") {
		path = "./" + path
	}
	return path
}

// ImportAlias returns the import name for desc.
//
// It concats the package name with the filename.
// E.g.: if desc describes a proto import:
//
// 	import "google/protobuf/timestamp.proto"
//
// ImportAlias returns "google_protobuf_timestamp_pb".
func importAlias(desc protoreflect.FileDescriptor) string {
	// TODO: pkg might have a leading dot?
	pkg := string(desc.Package())
	base := path.Base(desc.Path())
	alias := strings.Replace(pkg, ".", "_", -1) + "_" +
		strings.TrimSuffix(base, ".proto") + "_pb"
	return alias
}

// nameInContext returns the Message name for msg in context ctx.
//
// If msg is defined in the file described by ctx, the short Message name of msg
// is returned. If msg is defined in a file other that ctx, the stort Message
// name of msg prefix with the import alias for file that defines msg is returned.
func nameInContext(ctx protoreflect.FileDescriptor, msg protoreflect.MessageDescriptor) string {
	if ctx.Path() != msg.ParentFile().Path() {
		alias := importAlias(msg.ParentFile())
		return alias + "." + trimPackagePrefix(msg)
	}
	// If msg is defined in ctx, do not prefix with import alias.
	return trimPackagePrefix(msg)
}

// enumInContext returns the Enum name for enum in context ctx.
//
// If enum is defined in the file described by ctx, the short Enum name of msg
// is returned. If enum is defined in a file other that ctx, the Enum name of
// enum prefix with the import alias for file that defines enum is returned.
func enumNameInContext(ctx protoreflect.FileDescriptor, enum protoreflect.EnumDescriptor) string {
	if ctx.Path() != enum.ParentFile().Path() {
		alias := importAlias(enum.ParentFile())
		return alias + "." + trimPackagePrefix(enum)
	}
	// If enum is defined in ctx, do not prefix with import alias.
	return trimPackagePrefix(enum)
}

// trimPackagePrefix returns the full name for desc with the package prefix
// trimmed.
//
// E.g. for desc with FullName "my.awesome.package.MyMessage.MyNestedMessage" it
// returns "MyMessage.MyNestedMessage".
func trimPackagePrefix(desc protoreflect.Descriptor) string {
	name := string(desc.FullName())
	pkg := string(desc.ParentFile().Package())
	return strings.TrimPrefix(name, pkg+".")
}
