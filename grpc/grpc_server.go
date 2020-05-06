package grpc

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	strfmt "github.com/go-openapi/strfmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	logcore "github.com/jedrp/go-core/log"
	"github.com/jedrp/go-core/util"
	flags "github.com/jessevdk/go-flags"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

type ServicesRegistrationFunc func(*grpc.Server)

// Server wrapper object to run grpc services
// setting string : SERVER_CONFIG|server-config="host=0.0.0.0;port=80;tlsCert=;tlsCertKey=;"
type Server struct {
	SettingStr string `long:"server-config" description:"the server setting string" env:"SERVER_CONFIG" json:"settingStr,omitempty"`
	host       string
	port       int
	logger     logcore.Logger
	grpcServer *grpc.Server
}

const (
	tlsCert    = "tlsCert"
	tlsCertKey = "tlsCertKey"
	host       = "host"
	port       = "port"
)

func NewServer(servicesRegistrationFunc ServicesRegistrationFunc, logger logcore.Logger) *Server {

	server := &Server{
		logger: logger,
	}
	parser := flags.NewParser(server, flags.IgnoreUnknown)
	ParseConfig(parser)

	if server.SettingStr == "" {
		logger.Warn("empty server setting string")
	} else {
		logger.Infof("server start with setting string: %v", server.SettingStr)
	}
	config, err := util.GetConfig(server.SettingStr)
	if err != nil {
		panic(err)
	}
	cert := config[tlsCert]
	certKey := config[tlsCertKey]
	if v, found := config[host]; found {
		server.host = v
	} else {
		server.host = "127.0.0.1"
	}
	if v, found := config[port]; found {
		server.port, _ = strconv.Atoi(v)
	} else {
		server.port = 0
	}

	formats := strfmt.Default
	var grpcServer *grpc.Server
	grpcOpts := []grpc.ServerOption{grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		UnaryServerRequestContextInterceptor(),
		UnaryServerPanicInterceptor(logger),
		UnaryValidatorServerInterceptor(formats, logger),
	)), grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
		StreamServerRequestInterceptor(),
		grpc_recovery.StreamServerInterceptor(
			grpc_recovery.WithRecoveryHandlerContext(getRecoveryHandlerFuncContextHandler(logger)),
		),
		StreamValidatorServerInterceptor(formats, logger),
	)), grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle: 5 * time.Minute, // <--- This fixes it!
	})}
	if cert != "" || certKey != "" {
		creds, err := credentials.NewServerTLSFromFile(cert, certKey)
		if err != nil {
			logger.Fatal(err)
		}
		grpcOpts = append(grpcOpts, grpc.Creds(creds))
	}

	grpcServer = grpc.NewServer(grpcOpts...)
	if servicesRegistrationFunc != nil {
		servicesRegistrationFunc(grpcServer)
	}

	server.grpcServer = grpcServer

	return server
}

func (s *Server) Serve() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		s.logger.Panicf("failed to listen: %v", err)
	}
	s.logger.Infof("Sever starting serving gRPC at: %s\n", lis.Addr())
	if err = s.grpcServer.Serve(lis); err != nil {
		s.logger.Fatalf("failed to serve: %s", err)
		return err
	}
	return nil
}

func ParseConfig(parser *flags.Parser) {
	if _, err := parser.Parse(); err != nil {
		log.Fatalln(err)
		code := 1
		if fe, ok := err.(*flags.Error); ok {
			if fe.Type == flags.ErrHelp {
				code = 0
			}
		}
		os.Exit(code)
	}
}
