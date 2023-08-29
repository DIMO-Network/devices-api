package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/DIMO-Network/devices-api/internal/appmetrics"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCPanicker struct {
	Logger *zerolog.Logger
}

func (pr *GRPCPanicker) GRPCPanicRecoveryHandler(p any) (err error) {
	appmetrics.GRPCPanicCount.Inc()

	pr.Logger.Err(fmt.Errorf("%s", p)).Str("stack", string(debug.Stack())).Msg("grpc recovered from panic")
	return status.Errorf(codes.Internal, "%s", p)
}
