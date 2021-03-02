// Code generated by protoc-bin.sh-gen-go-grpc. DO NOT EDIT.

package proto

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// SourceLiveEventsClient is the client API for SourceLiveEvents service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SourceLiveEventsClient interface {
	FetchSourceLiveEvents(ctx context.Context, in *FetchSourceLiveEventsRequest, opts ...grpc.CallOption) (SourceLiveEvents_FetchSourceLiveEventsClient, error)
	StopSourceLiveEvents(ctx context.Context, in *StopSourceLiveEventsRequest, opts ...grpc.CallOption) (*StopSourceLiveEventsResponse, error)
}

type sourceLiveEventsClient struct {
	cc grpc.ClientConnInterface
}

func NewSourceLiveEventsClient(cc grpc.ClientConnInterface) SourceLiveEventsClient {
	return &sourceLiveEventsClient{cc}
}

func (c *sourceLiveEventsClient) FetchSourceLiveEvents(ctx context.Context, in *FetchSourceLiveEventsRequest, opts ...grpc.CallOption) (SourceLiveEvents_FetchSourceLiveEventsClient, error) {
	stream, err := c.cc.NewStream(ctx, &_SourceLiveEvents_serviceDesc.Streams[0], "/proto.SourceLiveEvents/fetchSourceLiveEvents", opts...)
	if err != nil {
		return nil, err
	}
	x := &sourceLiveEventsFetchSourceLiveEventsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type SourceLiveEvents_FetchSourceLiveEventsClient interface {
	Recv() (*FetchSourceLiveEventsResponse, error)
	grpc.ClientStream
}

type sourceLiveEventsFetchSourceLiveEventsClient struct {
	grpc.ClientStream
}

func (x *sourceLiveEventsFetchSourceLiveEventsClient) Recv() (*FetchSourceLiveEventsResponse, error) {
	m := new(FetchSourceLiveEventsResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *sourceLiveEventsClient) StopSourceLiveEvents(ctx context.Context, in *StopSourceLiveEventsRequest, opts ...grpc.CallOption) (*StopSourceLiveEventsResponse, error) {
	out := new(StopSourceLiveEventsResponse)
	err := c.cc.Invoke(ctx, "/proto.SourceLiveEvents/stopSourceLiveEvents", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SourceLiveEventsServer is the server API for SourceLiveEvents service.
// All implementations must embed UnimplementedSourceLiveEventsServer
// for forward compatibility
type SourceLiveEventsServer interface {
	FetchSourceLiveEvents(*FetchSourceLiveEventsRequest, SourceLiveEvents_FetchSourceLiveEventsServer) error
	StopSourceLiveEvents(context.Context, *StopSourceLiveEventsRequest) (*StopSourceLiveEventsResponse, error)
	mustEmbedUnimplementedSourceLiveEventsServer()
}

// UnimplementedSourceLiveEventsServer must be embedded to have forward compatible implementations.
type UnimplementedSourceLiveEventsServer struct {
}

func (UnimplementedSourceLiveEventsServer) FetchSourceLiveEvents(*FetchSourceLiveEventsRequest, SourceLiveEvents_FetchSourceLiveEventsServer) error {
	return status.Errorf(codes.Unimplemented, "method FetchSourceLiveEvents not implemented")
}
func (UnimplementedSourceLiveEventsServer) StopSourceLiveEvents(context.Context, *StopSourceLiveEventsRequest) (*StopSourceLiveEventsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StopSourceLiveEvents not implemented")
}
func (UnimplementedSourceLiveEventsServer) mustEmbedUnimplementedSourceLiveEventsServer() {}

// UnsafeSourceLiveEventsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SourceLiveEventsServer will
// result in compilation errors.
type UnsafeSourceLiveEventsServer interface {
	mustEmbedUnimplementedSourceLiveEventsServer()
}

func RegisterSourceLiveEventsServer(s grpc.ServiceRegistrar, srv SourceLiveEventsServer) {
	s.RegisterService(&_SourceLiveEvents_serviceDesc, srv)
}

func _SourceLiveEvents_FetchSourceLiveEvents_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(FetchSourceLiveEventsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SourceLiveEventsServer).FetchSourceLiveEvents(m, &sourceLiveEventsFetchSourceLiveEventsServer{stream})
}

type SourceLiveEvents_FetchSourceLiveEventsServer interface {
	Send(*FetchSourceLiveEventsResponse) error
	grpc.ServerStream
}

type sourceLiveEventsFetchSourceLiveEventsServer struct {
	grpc.ServerStream
}

func (x *sourceLiveEventsFetchSourceLiveEventsServer) Send(m *FetchSourceLiveEventsResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _SourceLiveEvents_StopSourceLiveEvents_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StopSourceLiveEventsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SourceLiveEventsServer).StopSourceLiveEvents(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.SourceLiveEvents/stopSourceLiveEvents",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SourceLiveEventsServer).StopSourceLiveEvents(ctx, req.(*StopSourceLiveEventsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _SourceLiveEvents_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.SourceLiveEvents",
	HandlerType: (*SourceLiveEventsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "stopSourceLiveEvents",
			Handler:    _SourceLiveEvents_StopSourceLiveEvents_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "fetchSourceLiveEvents",
			Handler:       _SourceLiveEvents_FetchSourceLiveEvents_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "proto/rudder-server/rudder-server.proto",
}
