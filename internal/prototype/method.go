package prototype

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// InputType retuns the Message name (in context) of desc.Input().
func InputType(desc protoreflect.MethodDescriptor) string {
	return nameInContext(desc.ParentFile(), desc.Input())
}

// Output retuns the Message name (in context) of desc.Output().
func OutputType(desc protoreflect.MethodDescriptor) string {
	return nameInContext(desc.ParentFile(), desc.Output())
}

// Address returns the method name as expected by grpc web.
//
// For a MethodDescriptor with FullName
//
// 	hello.world.Greeter.Greet
//
// it returns
//
// 	/hello.world.Greeter/Greet
//
func Address(desc protoreflect.MethodDescriptor) string {
	// Replace the last ".", that is in front of the method name with a "/",
	// to obtain the address for m.
	fn := string(desc.FullName())
	idx := strings.LastIndex(fn, ".")
	path := []rune(fn)
	path[idx] = '/'
	return "/" + string(path)
}

// MetMethodType returns "server_streaming" or "unary".
//
// Panics if desc.IsStreamingClient evaluates to true, because GRPC web does not
// support client side streaming.
func MethodType(desc protoreflect.MethodDescriptor) string {
	if desc.IsStreamingServer() {
		return "server_streaming"
	}
	if desc.IsStreamingClient() {
		panic(fmt.Sprintf("Client side streaming method (%s) is not supported", desc.FullName()))
	}
	return "unary"
}
