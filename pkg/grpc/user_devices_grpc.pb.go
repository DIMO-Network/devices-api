// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.4.0
// - protoc             v5.27.1
// source: pkg/grpc/user_devices.proto

package grpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.62.0 or later.
const _ = grpc.SupportPackageIsVersion8

const (
	UserDeviceService_GetUserDevice_FullMethodName                 = "/devices.UserDeviceService/GetUserDevice"
	UserDeviceService_GetUserDeviceByTokenId_FullMethodName        = "/devices.UserDeviceService/GetUserDeviceByTokenId"
	UserDeviceService_GetUserDeviceByVIN_FullMethodName            = "/devices.UserDeviceService/GetUserDeviceByVIN"
	UserDeviceService_GetUserDeviceByEthAddr_FullMethodName        = "/devices.UserDeviceService/GetUserDeviceByEthAddr"
	UserDeviceService_ListUserDevicesForUser_FullMethodName        = "/devices.UserDeviceService/ListUserDevicesForUser"
	UserDeviceService_ApplyHardwareTemplate_FullMethodName         = "/devices.UserDeviceService/ApplyHardwareTemplate"
	UserDeviceService_GetUserDeviceByAutoPIUnitId_FullMethodName   = "/devices.UserDeviceService/GetUserDeviceByAutoPIUnitId"
	UserDeviceService_GetClaimedVehiclesGrowth_FullMethodName      = "/devices.UserDeviceService/GetClaimedVehiclesGrowth"
	UserDeviceService_CreateTemplate_FullMethodName                = "/devices.UserDeviceService/CreateTemplate"
	UserDeviceService_RegisterUserDeviceFromVIN_FullMethodName     = "/devices.UserDeviceService/RegisterUserDeviceFromVIN"
	UserDeviceService_UpdateDeviceIntegrationStatus_FullMethodName = "/devices.UserDeviceService/UpdateDeviceIntegrationStatus"
	UserDeviceService_GetAllUserDevice_FullMethodName              = "/devices.UserDeviceService/GetAllUserDevice"
	UserDeviceService_UpdateUserDeviceMetadata_FullMethodName      = "/devices.UserDeviceService/UpdateUserDeviceMetadata"
	UserDeviceService_ClearMetaTransactionRequests_FullMethodName  = "/devices.UserDeviceService/ClearMetaTransactionRequests"
	UserDeviceService_StopUserDeviceIntegration_FullMethodName     = "/devices.UserDeviceService/StopUserDeviceIntegration"
)

// UserDeviceServiceClient is the client API for UserDeviceService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UserDeviceServiceClient interface {
	GetUserDevice(ctx context.Context, in *GetUserDeviceRequest, opts ...grpc.CallOption) (*UserDevice, error)
	GetUserDeviceByTokenId(ctx context.Context, in *GetUserDeviceByTokenIdRequest, opts ...grpc.CallOption) (*UserDevice, error)
	GetUserDeviceByVIN(ctx context.Context, in *GetUserDeviceByVINRequest, opts ...grpc.CallOption) (*UserDevice, error)
	GetUserDeviceByEthAddr(ctx context.Context, in *GetUserDeviceByEthAddrRequest, opts ...grpc.CallOption) (*UserDevice, error)
	ListUserDevicesForUser(ctx context.Context, in *ListUserDevicesForUserRequest, opts ...grpc.CallOption) (*ListUserDevicesForUserResponse, error)
	ApplyHardwareTemplate(ctx context.Context, in *ApplyHardwareTemplateRequest, opts ...grpc.CallOption) (*ApplyHardwareTemplateResponse, error)
	GetUserDeviceByAutoPIUnitId(ctx context.Context, in *GetUserDeviceByAutoPIUnitIdRequest, opts ...grpc.CallOption) (*UserDeviceAutoPIUnitResponse, error)
	GetClaimedVehiclesGrowth(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ClaimedVehiclesGrowth, error)
	CreateTemplate(ctx context.Context, in *CreateTemplateRequest, opts ...grpc.CallOption) (*CreateTemplateResponse, error)
	RegisterUserDeviceFromVIN(ctx context.Context, in *RegisterUserDeviceFromVINRequest, opts ...grpc.CallOption) (*RegisterUserDeviceFromVINResponse, error)
	UpdateDeviceIntegrationStatus(ctx context.Context, in *UpdateDeviceIntegrationStatusRequest, opts ...grpc.CallOption) (*UserDevice, error)
	GetAllUserDevice(ctx context.Context, in *GetAllUserDeviceRequest, opts ...grpc.CallOption) (UserDeviceService_GetAllUserDeviceClient, error)
	// used to update metadata properties, currently only ones needed by valuations-api
	UpdateUserDeviceMetadata(ctx context.Context, in *UpdateUserDeviceMetadataRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	ClearMetaTransactionRequests(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ClearMetaTransactionRequestsResponse, error)
	StopUserDeviceIntegration(ctx context.Context, in *StopUserDeviceIntegrationRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type userDeviceServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewUserDeviceServiceClient(cc grpc.ClientConnInterface) UserDeviceServiceClient {
	return &userDeviceServiceClient{cc}
}

func (c *userDeviceServiceClient) GetUserDevice(ctx context.Context, in *GetUserDeviceRequest, opts ...grpc.CallOption) (*UserDevice, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UserDevice)
	err := c.cc.Invoke(ctx, UserDeviceService_GetUserDevice_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userDeviceServiceClient) GetUserDeviceByTokenId(ctx context.Context, in *GetUserDeviceByTokenIdRequest, opts ...grpc.CallOption) (*UserDevice, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UserDevice)
	err := c.cc.Invoke(ctx, UserDeviceService_GetUserDeviceByTokenId_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userDeviceServiceClient) GetUserDeviceByVIN(ctx context.Context, in *GetUserDeviceByVINRequest, opts ...grpc.CallOption) (*UserDevice, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UserDevice)
	err := c.cc.Invoke(ctx, UserDeviceService_GetUserDeviceByVIN_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userDeviceServiceClient) GetUserDeviceByEthAddr(ctx context.Context, in *GetUserDeviceByEthAddrRequest, opts ...grpc.CallOption) (*UserDevice, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UserDevice)
	err := c.cc.Invoke(ctx, UserDeviceService_GetUserDeviceByEthAddr_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userDeviceServiceClient) ListUserDevicesForUser(ctx context.Context, in *ListUserDevicesForUserRequest, opts ...grpc.CallOption) (*ListUserDevicesForUserResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListUserDevicesForUserResponse)
	err := c.cc.Invoke(ctx, UserDeviceService_ListUserDevicesForUser_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userDeviceServiceClient) ApplyHardwareTemplate(ctx context.Context, in *ApplyHardwareTemplateRequest, opts ...grpc.CallOption) (*ApplyHardwareTemplateResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ApplyHardwareTemplateResponse)
	err := c.cc.Invoke(ctx, UserDeviceService_ApplyHardwareTemplate_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userDeviceServiceClient) GetUserDeviceByAutoPIUnitId(ctx context.Context, in *GetUserDeviceByAutoPIUnitIdRequest, opts ...grpc.CallOption) (*UserDeviceAutoPIUnitResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UserDeviceAutoPIUnitResponse)
	err := c.cc.Invoke(ctx, UserDeviceService_GetUserDeviceByAutoPIUnitId_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userDeviceServiceClient) GetClaimedVehiclesGrowth(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ClaimedVehiclesGrowth, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ClaimedVehiclesGrowth)
	err := c.cc.Invoke(ctx, UserDeviceService_GetClaimedVehiclesGrowth_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userDeviceServiceClient) CreateTemplate(ctx context.Context, in *CreateTemplateRequest, opts ...grpc.CallOption) (*CreateTemplateResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateTemplateResponse)
	err := c.cc.Invoke(ctx, UserDeviceService_CreateTemplate_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userDeviceServiceClient) RegisterUserDeviceFromVIN(ctx context.Context, in *RegisterUserDeviceFromVINRequest, opts ...grpc.CallOption) (*RegisterUserDeviceFromVINResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RegisterUserDeviceFromVINResponse)
	err := c.cc.Invoke(ctx, UserDeviceService_RegisterUserDeviceFromVIN_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userDeviceServiceClient) UpdateDeviceIntegrationStatus(ctx context.Context, in *UpdateDeviceIntegrationStatusRequest, opts ...grpc.CallOption) (*UserDevice, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UserDevice)
	err := c.cc.Invoke(ctx, UserDeviceService_UpdateDeviceIntegrationStatus_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userDeviceServiceClient) GetAllUserDevice(ctx context.Context, in *GetAllUserDeviceRequest, opts ...grpc.CallOption) (UserDeviceService_GetAllUserDeviceClient, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &UserDeviceService_ServiceDesc.Streams[0], UserDeviceService_GetAllUserDevice_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &userDeviceServiceGetAllUserDeviceClient{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type UserDeviceService_GetAllUserDeviceClient interface {
	Recv() (*UserDevice, error)
	grpc.ClientStream
}

type userDeviceServiceGetAllUserDeviceClient struct {
	grpc.ClientStream
}

func (x *userDeviceServiceGetAllUserDeviceClient) Recv() (*UserDevice, error) {
	m := new(UserDevice)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *userDeviceServiceClient) UpdateUserDeviceMetadata(ctx context.Context, in *UpdateUserDeviceMetadataRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, UserDeviceService_UpdateUserDeviceMetadata_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userDeviceServiceClient) ClearMetaTransactionRequests(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ClearMetaTransactionRequestsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ClearMetaTransactionRequestsResponse)
	err := c.cc.Invoke(ctx, UserDeviceService_ClearMetaTransactionRequests_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userDeviceServiceClient) StopUserDeviceIntegration(ctx context.Context, in *StopUserDeviceIntegrationRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, UserDeviceService_StopUserDeviceIntegration_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UserDeviceServiceServer is the server API for UserDeviceService service.
// All implementations must embed UnimplementedUserDeviceServiceServer
// for forward compatibility
type UserDeviceServiceServer interface {
	GetUserDevice(context.Context, *GetUserDeviceRequest) (*UserDevice, error)
	GetUserDeviceByTokenId(context.Context, *GetUserDeviceByTokenIdRequest) (*UserDevice, error)
	GetUserDeviceByVIN(context.Context, *GetUserDeviceByVINRequest) (*UserDevice, error)
	GetUserDeviceByEthAddr(context.Context, *GetUserDeviceByEthAddrRequest) (*UserDevice, error)
	ListUserDevicesForUser(context.Context, *ListUserDevicesForUserRequest) (*ListUserDevicesForUserResponse, error)
	ApplyHardwareTemplate(context.Context, *ApplyHardwareTemplateRequest) (*ApplyHardwareTemplateResponse, error)
	GetUserDeviceByAutoPIUnitId(context.Context, *GetUserDeviceByAutoPIUnitIdRequest) (*UserDeviceAutoPIUnitResponse, error)
	GetClaimedVehiclesGrowth(context.Context, *emptypb.Empty) (*ClaimedVehiclesGrowth, error)
	CreateTemplate(context.Context, *CreateTemplateRequest) (*CreateTemplateResponse, error)
	RegisterUserDeviceFromVIN(context.Context, *RegisterUserDeviceFromVINRequest) (*RegisterUserDeviceFromVINResponse, error)
	UpdateDeviceIntegrationStatus(context.Context, *UpdateDeviceIntegrationStatusRequest) (*UserDevice, error)
	GetAllUserDevice(*GetAllUserDeviceRequest, UserDeviceService_GetAllUserDeviceServer) error
	// used to update metadata properties, currently only ones needed by valuations-api
	UpdateUserDeviceMetadata(context.Context, *UpdateUserDeviceMetadataRequest) (*emptypb.Empty, error)
	ClearMetaTransactionRequests(context.Context, *emptypb.Empty) (*ClearMetaTransactionRequestsResponse, error)
	StopUserDeviceIntegration(context.Context, *StopUserDeviceIntegrationRequest) (*emptypb.Empty, error)
	mustEmbedUnimplementedUserDeviceServiceServer()
}

// UnimplementedUserDeviceServiceServer must be embedded to have forward compatible implementations.
type UnimplementedUserDeviceServiceServer struct {
}

func (UnimplementedUserDeviceServiceServer) GetUserDevice(context.Context, *GetUserDeviceRequest) (*UserDevice, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserDevice not implemented")
}
func (UnimplementedUserDeviceServiceServer) GetUserDeviceByTokenId(context.Context, *GetUserDeviceByTokenIdRequest) (*UserDevice, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserDeviceByTokenId not implemented")
}
func (UnimplementedUserDeviceServiceServer) GetUserDeviceByVIN(context.Context, *GetUserDeviceByVINRequest) (*UserDevice, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserDeviceByVIN not implemented")
}
func (UnimplementedUserDeviceServiceServer) GetUserDeviceByEthAddr(context.Context, *GetUserDeviceByEthAddrRequest) (*UserDevice, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserDeviceByEthAddr not implemented")
}
func (UnimplementedUserDeviceServiceServer) ListUserDevicesForUser(context.Context, *ListUserDevicesForUserRequest) (*ListUserDevicesForUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListUserDevicesForUser not implemented")
}
func (UnimplementedUserDeviceServiceServer) ApplyHardwareTemplate(context.Context, *ApplyHardwareTemplateRequest) (*ApplyHardwareTemplateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ApplyHardwareTemplate not implemented")
}
func (UnimplementedUserDeviceServiceServer) GetUserDeviceByAutoPIUnitId(context.Context, *GetUserDeviceByAutoPIUnitIdRequest) (*UserDeviceAutoPIUnitResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserDeviceByAutoPIUnitId not implemented")
}
func (UnimplementedUserDeviceServiceServer) GetClaimedVehiclesGrowth(context.Context, *emptypb.Empty) (*ClaimedVehiclesGrowth, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetClaimedVehiclesGrowth not implemented")
}
func (UnimplementedUserDeviceServiceServer) CreateTemplate(context.Context, *CreateTemplateRequest) (*CreateTemplateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateTemplate not implemented")
}
func (UnimplementedUserDeviceServiceServer) RegisterUserDeviceFromVIN(context.Context, *RegisterUserDeviceFromVINRequest) (*RegisterUserDeviceFromVINResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterUserDeviceFromVIN not implemented")
}
func (UnimplementedUserDeviceServiceServer) UpdateDeviceIntegrationStatus(context.Context, *UpdateDeviceIntegrationStatusRequest) (*UserDevice, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateDeviceIntegrationStatus not implemented")
}
func (UnimplementedUserDeviceServiceServer) GetAllUserDevice(*GetAllUserDeviceRequest, UserDeviceService_GetAllUserDeviceServer) error {
	return status.Errorf(codes.Unimplemented, "method GetAllUserDevice not implemented")
}
func (UnimplementedUserDeviceServiceServer) UpdateUserDeviceMetadata(context.Context, *UpdateUserDeviceMetadataRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateUserDeviceMetadata not implemented")
}
func (UnimplementedUserDeviceServiceServer) ClearMetaTransactionRequests(context.Context, *emptypb.Empty) (*ClearMetaTransactionRequestsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ClearMetaTransactionRequests not implemented")
}
func (UnimplementedUserDeviceServiceServer) StopUserDeviceIntegration(context.Context, *StopUserDeviceIntegrationRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StopUserDeviceIntegration not implemented")
}
func (UnimplementedUserDeviceServiceServer) mustEmbedUnimplementedUserDeviceServiceServer() {}

// UnsafeUserDeviceServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UserDeviceServiceServer will
// result in compilation errors.
type UnsafeUserDeviceServiceServer interface {
	mustEmbedUnimplementedUserDeviceServiceServer()
}

func RegisterUserDeviceServiceServer(s grpc.ServiceRegistrar, srv UserDeviceServiceServer) {
	s.RegisterService(&UserDeviceService_ServiceDesc, srv)
}

func _UserDeviceService_GetUserDevice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserDeviceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserDeviceServiceServer).GetUserDevice(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserDeviceService_GetUserDevice_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserDeviceServiceServer).GetUserDevice(ctx, req.(*GetUserDeviceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserDeviceService_GetUserDeviceByTokenId_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserDeviceByTokenIdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserDeviceServiceServer).GetUserDeviceByTokenId(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserDeviceService_GetUserDeviceByTokenId_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserDeviceServiceServer).GetUserDeviceByTokenId(ctx, req.(*GetUserDeviceByTokenIdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserDeviceService_GetUserDeviceByVIN_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserDeviceByVINRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserDeviceServiceServer).GetUserDeviceByVIN(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserDeviceService_GetUserDeviceByVIN_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserDeviceServiceServer).GetUserDeviceByVIN(ctx, req.(*GetUserDeviceByVINRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserDeviceService_GetUserDeviceByEthAddr_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserDeviceByEthAddrRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserDeviceServiceServer).GetUserDeviceByEthAddr(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserDeviceService_GetUserDeviceByEthAddr_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserDeviceServiceServer).GetUserDeviceByEthAddr(ctx, req.(*GetUserDeviceByEthAddrRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserDeviceService_ListUserDevicesForUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListUserDevicesForUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserDeviceServiceServer).ListUserDevicesForUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserDeviceService_ListUserDevicesForUser_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserDeviceServiceServer).ListUserDevicesForUser(ctx, req.(*ListUserDevicesForUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserDeviceService_ApplyHardwareTemplate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ApplyHardwareTemplateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserDeviceServiceServer).ApplyHardwareTemplate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserDeviceService_ApplyHardwareTemplate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserDeviceServiceServer).ApplyHardwareTemplate(ctx, req.(*ApplyHardwareTemplateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserDeviceService_GetUserDeviceByAutoPIUnitId_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserDeviceByAutoPIUnitIdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserDeviceServiceServer).GetUserDeviceByAutoPIUnitId(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserDeviceService_GetUserDeviceByAutoPIUnitId_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserDeviceServiceServer).GetUserDeviceByAutoPIUnitId(ctx, req.(*GetUserDeviceByAutoPIUnitIdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserDeviceService_GetClaimedVehiclesGrowth_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserDeviceServiceServer).GetClaimedVehiclesGrowth(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserDeviceService_GetClaimedVehiclesGrowth_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserDeviceServiceServer).GetClaimedVehiclesGrowth(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserDeviceService_CreateTemplate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateTemplateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserDeviceServiceServer).CreateTemplate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserDeviceService_CreateTemplate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserDeviceServiceServer).CreateTemplate(ctx, req.(*CreateTemplateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserDeviceService_RegisterUserDeviceFromVIN_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterUserDeviceFromVINRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserDeviceServiceServer).RegisterUserDeviceFromVIN(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserDeviceService_RegisterUserDeviceFromVIN_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserDeviceServiceServer).RegisterUserDeviceFromVIN(ctx, req.(*RegisterUserDeviceFromVINRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserDeviceService_UpdateDeviceIntegrationStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateDeviceIntegrationStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserDeviceServiceServer).UpdateDeviceIntegrationStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserDeviceService_UpdateDeviceIntegrationStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserDeviceServiceServer).UpdateDeviceIntegrationStatus(ctx, req.(*UpdateDeviceIntegrationStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserDeviceService_GetAllUserDevice_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetAllUserDeviceRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(UserDeviceServiceServer).GetAllUserDevice(m, &userDeviceServiceGetAllUserDeviceServer{ServerStream: stream})
}

type UserDeviceService_GetAllUserDeviceServer interface {
	Send(*UserDevice) error
	grpc.ServerStream
}

type userDeviceServiceGetAllUserDeviceServer struct {
	grpc.ServerStream
}

func (x *userDeviceServiceGetAllUserDeviceServer) Send(m *UserDevice) error {
	return x.ServerStream.SendMsg(m)
}

func _UserDeviceService_UpdateUserDeviceMetadata_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateUserDeviceMetadataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserDeviceServiceServer).UpdateUserDeviceMetadata(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserDeviceService_UpdateUserDeviceMetadata_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserDeviceServiceServer).UpdateUserDeviceMetadata(ctx, req.(*UpdateUserDeviceMetadataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserDeviceService_ClearMetaTransactionRequests_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserDeviceServiceServer).ClearMetaTransactionRequests(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserDeviceService_ClearMetaTransactionRequests_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserDeviceServiceServer).ClearMetaTransactionRequests(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserDeviceService_StopUserDeviceIntegration_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StopUserDeviceIntegrationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserDeviceServiceServer).StopUserDeviceIntegration(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserDeviceService_StopUserDeviceIntegration_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserDeviceServiceServer).StopUserDeviceIntegration(ctx, req.(*StopUserDeviceIntegrationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// UserDeviceService_ServiceDesc is the grpc.ServiceDesc for UserDeviceService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var UserDeviceService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "devices.UserDeviceService",
	HandlerType: (*UserDeviceServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetUserDevice",
			Handler:    _UserDeviceService_GetUserDevice_Handler,
		},
		{
			MethodName: "GetUserDeviceByTokenId",
			Handler:    _UserDeviceService_GetUserDeviceByTokenId_Handler,
		},
		{
			MethodName: "GetUserDeviceByVIN",
			Handler:    _UserDeviceService_GetUserDeviceByVIN_Handler,
		},
		{
			MethodName: "GetUserDeviceByEthAddr",
			Handler:    _UserDeviceService_GetUserDeviceByEthAddr_Handler,
		},
		{
			MethodName: "ListUserDevicesForUser",
			Handler:    _UserDeviceService_ListUserDevicesForUser_Handler,
		},
		{
			MethodName: "ApplyHardwareTemplate",
			Handler:    _UserDeviceService_ApplyHardwareTemplate_Handler,
		},
		{
			MethodName: "GetUserDeviceByAutoPIUnitId",
			Handler:    _UserDeviceService_GetUserDeviceByAutoPIUnitId_Handler,
		},
		{
			MethodName: "GetClaimedVehiclesGrowth",
			Handler:    _UserDeviceService_GetClaimedVehiclesGrowth_Handler,
		},
		{
			MethodName: "CreateTemplate",
			Handler:    _UserDeviceService_CreateTemplate_Handler,
		},
		{
			MethodName: "RegisterUserDeviceFromVIN",
			Handler:    _UserDeviceService_RegisterUserDeviceFromVIN_Handler,
		},
		{
			MethodName: "UpdateDeviceIntegrationStatus",
			Handler:    _UserDeviceService_UpdateDeviceIntegrationStatus_Handler,
		},
		{
			MethodName: "UpdateUserDeviceMetadata",
			Handler:    _UserDeviceService_UpdateUserDeviceMetadata_Handler,
		},
		{
			MethodName: "ClearMetaTransactionRequests",
			Handler:    _UserDeviceService_ClearMetaTransactionRequests_Handler,
		},
		{
			MethodName: "StopUserDeviceIntegration",
			Handler:    _UserDeviceService_StopUserDeviceIntegration_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetAllUserDevice",
			Handler:       _UserDeviceService_GetAllUserDevice_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "pkg/grpc/user_devices.proto",
}
