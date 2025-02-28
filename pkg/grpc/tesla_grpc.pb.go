// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: pkg/grpc/tesla.proto

// more types here: https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#google.protobuf.Empty

// import "google/protobuf/timestamp.proto";
// import "google/protobuf/empty.proto";
// import "pkg/grpc/aftermarket_devices.proto";

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
	TeslaService_CheckFleetTelemetryCapable_FullMethodName = "/tesla.TeslaService/CheckFleetTelemetryCapable"
)

// TeslaServiceClient is the client API for TeslaService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TeslaServiceClient interface {
	// Tries to determine whether the vehicle with the given id is Fleet Telemetry-capable.
	// Note that if the vehicle is not already enrolled then calling this function will
	// attempt to enroll it.
	//
	// Will fail if the credentials are expired or if the device is not minted.
	CheckFleetTelemetryCapable(ctx context.Context, in *CheckFleetTelemetryCapableRequest, opts ...grpc.CallOption) (*CheckFleetTelemetryCapableResponse, error)
}

type teslaServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTeslaServiceClient(cc grpc.ClientConnInterface) TeslaServiceClient {
	return &teslaServiceClient{cc}
}

func (c *teslaServiceClient) CheckFleetTelemetryCapable(ctx context.Context, in *CheckFleetTelemetryCapableRequest, opts ...grpc.CallOption) (*CheckFleetTelemetryCapableResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CheckFleetTelemetryCapableResponse)
	err := c.cc.Invoke(ctx, TeslaService_CheckFleetTelemetryCapable_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TeslaServiceServer is the server API for TeslaService service.
// All implementations must embed UnimplementedTeslaServiceServer
// for forward compatibility.
type TeslaServiceServer interface {
	// Tries to determine whether the vehicle with the given id is Fleet Telemetry-capable.
	// Note that if the vehicle is not already enrolled then calling this function will
	// attempt to enroll it.
	//
	// Will fail if the credentials are expired or if the device is not minted.
	CheckFleetTelemetryCapable(context.Context, *CheckFleetTelemetryCapableRequest) (*CheckFleetTelemetryCapableResponse, error)
	mustEmbedUnimplementedTeslaServiceServer()
}

// UnimplementedTeslaServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedTeslaServiceServer struct{}

func (UnimplementedTeslaServiceServer) CheckFleetTelemetryCapable(context.Context, *CheckFleetTelemetryCapableRequest) (*CheckFleetTelemetryCapableResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckFleetTelemetryCapable not implemented")
}
func (UnimplementedTeslaServiceServer) mustEmbedUnimplementedTeslaServiceServer() {}
func (UnimplementedTeslaServiceServer) testEmbeddedByValue()                      {}

// UnsafeTeslaServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TeslaServiceServer will
// result in compilation errors.
type UnsafeTeslaServiceServer interface {
	mustEmbedUnimplementedTeslaServiceServer()
}

func RegisterTeslaServiceServer(s grpc.ServiceRegistrar, srv TeslaServiceServer) {
	// If the following call pancis, it indicates UnimplementedTeslaServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&TeslaService_ServiceDesc, srv)
}

func _TeslaService_CheckFleetTelemetryCapable_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckFleetTelemetryCapableRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TeslaServiceServer).CheckFleetTelemetryCapable(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TeslaService_CheckFleetTelemetryCapable_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TeslaServiceServer).CheckFleetTelemetryCapable(ctx, req.(*CheckFleetTelemetryCapableRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// TeslaService_ServiceDesc is the grpc.ServiceDesc for TeslaService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TeslaService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tesla.TeslaService",
	HandlerType: (*TeslaServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CheckFleetTelemetryCapable",
			Handler:    _TeslaService_CheckFleetTelemetryCapable_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/grpc/tesla.proto",
}
