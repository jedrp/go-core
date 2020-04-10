package rest

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jedrp/go-core/log"
	"github.com/jedrp/go-core/util"
	"github.com/jessevdk/go-flags"
	"golang.org/x/net/netutil"
)

// Server wrapper object to run rest service
// setting string : SERVER_CONFIG|server-config="host=0.0.0.0;port=80;tlsCert=;tlsCertKey=;"
type Server struct {
	SettingStr     string `long:"server-config" description:"the server setting string" env:"SERVER_CONFIG" json:"settingStr,omitempty"`
	host           string
	port           string
	logger         log.Logger
	httpServer     *http.Server
	listenLimit    int
	cleanupTimeout int
	keepAlive      int
	readTimeout    int
	writeTimeout   int
	listener       net.Listener
}

const (
	tlsCACert      = "tlsCACert"
	tlsCert        = "tlsCert"
	tlsCertKey     = "tlsCertKey"
	host           = "host"
	port           = "port"
	listenLimit    = "listenLimit"
	cleanupTimeout = "cleanupTimeout"
	readTimeout    = "readTimeout"
	writeTimeout   = "writeTimeout"
	keepAlive      = "keepAlive"
)

func NewServer(handler http.Handler, logger log.Logger) *Server {
	server := &Server{
		logger:       logger,
		readTimeout:  30,
		writeTimeout: 60,
		keepAlive:    180,
	}

	parser := flags.NewParser(server, flags.IgnoreUnknown)
	parseConfig(parser)

	if server.SettingStr == "" {
		panic("empty server setting string")
	}

	config, err := util.GetConfig(server.SettingStr)

	if v, found := config[host]; found {
		server.host = v
	}

	if v, found := config[port]; found {
		server.port = v
	} else {
		server.port = "0"
	}

	setConfigInt(config, &server.listenLimit, listenLimit)
	setConfigInt(config, &server.cleanupTimeout, cleanupTimeout)
	setConfigInt(config, &server.readTimeout, readTimeout)
	setConfigInt(config, &server.writeTimeout, writeTimeout)
	setConfigInt(config, &server.keepAlive, keepAlive)

	cert := config[tlsCert]
	certKey := config[tlsCertKey]
	certCA := config[tlsCACert]

	listener, err := net.Listen("tcp", net.JoinHostPort(server.host, server.port))

	if err != nil {
		panic(err)
	}
	httpServer := new(http.Server)
	httpServer.MaxHeaderBytes = 1024
	httpServer.ReadTimeout = time.Duration(server.readTimeout) * time.Second
	httpServer.WriteTimeout = time.Duration(server.writeTimeout) * time.Second
	httpServer.SetKeepAlivesEnabled(int64(server.keepAlive) > 0)
	if server.listenLimit > 0 {
		listener = netutil.LimitListener(listener, server.listenLimit)
	}
	if server.cleanupTimeout > 0 {
		httpServer.IdleTimeout = time.Duration(server.cleanupTimeout) * time.Second
	}
	httpServer.Handler = HandlePanicMiddleware(handler, logger)

	//https
	if cert != "" && certKey != "" {
		httpServer.TLSConfig = &tls.Config{
			// Causes servers to use Go's default ciphersuite preferences,
			// which are tuned to avoid attacks. Does nothing on clients.
			PreferServerCipherSuites: true,
			// Only use curves which have assembly implementations
			// https://github.com/golang/go/tree/master/src/crypto/elliptic
			CurvePreferences: []tls.CurveID{tls.CurveP256},
			// Use modern tls mode https://wiki.mozilla.org/Security/Server_Side_TLS#Modern_compatibility
			NextProtos: []string{"h2", "http/1.1"},
			// https://www.owasp.org/index.php/Transport_Layer_Protection_Cheat_Sheet#Rule_-_Only_Support_Strong_Protocols
			MinVersion: tls.VersionTLS12,
			// These ciphersuites support Forward Secrecy: https://en.wikipedia.org/wiki/Forward_secrecy
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			},
		}
		// build standard config from server options
		httpServer.TLSConfig.Certificates = make([]tls.Certificate, 1)
		httpServer.TLSConfig.Certificates[0], err = tls.LoadX509KeyPair(string(cert), string(certKey))
		if err != nil {
			panic(err)
		}

		if certCA != "" {
			// include specified CA certificate
			caCert, caCertErr := ioutil.ReadFile(string(certCA))
			if caCertErr != nil {
				panic(caCertErr)
			}
			caCertPool := x509.NewCertPool()
			ok := caCertPool.AppendCertsFromPEM(caCert)
			if !ok {
				panic(fmt.Errorf("cannot parse CA certificate"))
			}
			httpServer.TLSConfig.ClientCAs = caCertPool
			httpServer.TLSConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}
		if len(httpServer.TLSConfig.Certificates) == 0 && httpServer.TLSConfig.GetCertificate == nil {
			server.logger.Fatalf("no certificate was configured for TLS")
		}

		// must have at least one certificate or panics
		httpServer.TLSConfig.BuildNameToCertificate()
		server.listener = tls.NewListener(listener, httpServer.TLSConfig)
	} else {
		server.listener = listener
	}
	server.httpServer = httpServer
	return server
}

func (s *Server) Serve() error {
	s.logger.Infof("Sever starting serving REST(https)  at: %s", s.listener.Addr())
	if err := s.httpServer.Serve(s.listener); err != nil && err != http.ErrServerClosed {
		s.logger.Fatalf("%v", err)
		return err
	}
	return nil
}
func setConfigInt(config map[string]string, settingField *int, key string) {
	if v, found := config[key]; found {
		intValue, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}
		*settingField = intValue
	}
}
func parseConfig(parser *flags.Parser) {
	if _, err := parser.Parse(); err != nil {
		panic(err)
		code := 1
		if fe, ok := err.(*flags.Error); ok {
			if fe.Type == flags.ErrHelp {
				code = 0
			}
		}
		os.Exit(code)
	}
}
