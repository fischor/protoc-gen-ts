package prototype

import "google.golang.org/protobuf/reflect/protoreflect"

// RepeRepeatedFields returns the list of numbers of all repeated field of desc.
func RepeatedFields(desc protoreflect.MessageDescriptor) []int32 {
	var nn []int32
	fds := desc.Fields()
	for i := 0; i < fds.Len(); i++ {
		fd := fds.Get(i)
		if fd.Cardinality() == protoreflect.Repeated {
			nn = append(nn, int32(fd.Number()))
		}
	}
	return nn
}

// Oneofs returns the matrix pf oneof field numbers for all oneof fields present
// in desc.
func OneofFields(desc protoreflect.MessageDescriptor) [][]int32 {
	var union [][]int32
	oofs := desc.Oneofs()
	for i := 0; i < oofs.Len(); i++ {
		union = append(union, oneofs(oofs.Get(i)))
	}
	return union
}

func oneofs(desc protoreflect.OneofDescriptor) []int32 {
	var union []int32
	fds := desc.Fields()
	for i := 0; i < fds.Len(); i++ {
		union = append(union, int32(fds.Get(i).Number()))
	}
	return union
}
