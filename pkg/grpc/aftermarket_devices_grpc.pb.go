// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: pkg/grpc/aftermarket_devices.proto

package grpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	AftermarketDeviceService_ListAftermarketDevicesForUser_FullMethodName = "/devices.AftermarketDeviceService/ListAftermarketDevicesForUser"
)

// AftermarketDeviceServiceClient is the client API for AftermarketDeviceService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AftermarketDeviceServiceClient interface {
	ListAftermarketDevicesForUser(ctx context.Context, in *ListAftermarketDevicesForUserRequest, opts ...grpc.CallOption) (*ListAftermarketDevicesForUserResponse, error)
}

type aftermarketDeviceServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAftermarketDeviceServiceClient(cc grpc.ClientConnInterface) AftermarketDeviceServiceClient {
	return &aftermarketDeviceServiceClient{cc}
}

func (c *aftermarketDeviceServiceClient) ListAftermarketDevicesForUser(ctx context.Context, in *ListAftermarketDevicesForUserRequest, opts ...grpc.CallOption) (*ListAftermarketDevicesForUserResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListAftermarketDevicesForUserResponse)
	err := c.cc.Invoke(ctx, AftermarketDeviceService_ListAftermarketDevicesForUser_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AftermarketDeviceServiceServer is the server API for AftermarketDeviceService service.
// All implementations must embed UnimplementedAftermarketDeviceServiceServer
// for forward compatibility.
type AftermarketDeviceServiceServer interface {
	ListAftermarketDevicesForUser(context.Context, *ListAftermarketDevicesForUserRequest) (*ListAftermarketDevicesForUserResponse, error)
	mustEmbedUnimplementedAftermarketDeviceServiceServer()
}

// UnimplementedAftermarketDeviceServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedAftermarketDeviceServiceServer struct{}

func (UnimplementedAftermarketDeviceServiceServer) ListAftermarketDevicesForUser(context.Context, *ListAftermarketDevicesForUserRequest) (*ListAftermarketDevicesForUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAftermarketDevicesForUser not implemented")
}
func (UnimplementedAftermarketDeviceServiceServer) mustEmbedUnimplementedAftermarketDeviceServiceServer() {
}
func (UnimplementedAftermarketDeviceServiceServer) testEmbeddedByValue() {}

// UnsafeAftermarketDeviceServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AftermarketDeviceServiceServer will
// result in compilation errors.
type UnsafeAftermarketDeviceServiceServer interface {
	mustEmbedUnimplementedAftermarketDeviceServiceServer()
}

func RegisterAftermarketDeviceServiceServer(s grpc.ServiceRegistrar, srv AftermarketDeviceServiceServer) {
	// If the following call pancis, it indicates UnimplementedAftermarketDeviceServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&AftermarketDeviceService_ServiceDesc, srv)
}

func _AftermarketDeviceService_ListAftermarketDevicesForUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListAftermarketDevicesForUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AftermarketDeviceServiceServer).ListAftermarketDevicesForUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AftermarketDeviceService_ListAftermarketDevicesForUser_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AftermarketDeviceServiceServer).ListAftermarketDevicesForUser(ctx, req.(*ListAftermarketDevicesForUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// AftermarketDeviceService_ServiceDesc is the grpc.ServiceDesc for AftermarketDeviceService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AftermarketDeviceService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "devices.AftermarketDeviceService",
	HandlerType: (*AftermarketDeviceServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListAftermarketDevicesForUser",
			Handler:    _AftermarketDeviceService_ListAftermarketDevicesForUser_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/grpc/aftermarket_devices.proto",
}
