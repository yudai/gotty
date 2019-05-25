package server

import (
	"html/template"
	"log"
	"net"
	"net/http"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"

	"github.com/yudai/gotty/pkg/randomstring"
	"github.com/yudai/gotty/server/middleware"
	"github.com/yudai/gotty/webtty"
)

// Server provides a webtty HTTP endpoint.
type Server struct {
	factory Factory
	options *Options

	upgrader      *websocket.Upgrader
	indexTemplate *template.Template
}

// New creates a new instance of Server.
// Server will use the New() of the factory provided to handle each request.
func New(factory Factory, options *Options) (*Server, error) {
	indexData, err := Asset("static/index.html")
	if err != nil {
		panic("index not found") // must be in bindata
	}
	indexTemplate, err := template.New("index").Parse(string(indexData))
	if err != nil {
		panic("index template parse failed") // must be valid
	}

	return &Server{
		factory: factory,
		options: options,

		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			Subprotocols:    webtty.Protocols,
		},
		indexTemplate: indexTemplate,
	}, nil
}

// Run starts the main process of the Server.
// The cancelation of ctx will shutdown the server immediately with aborting
// existing connections. Use WithGracefullContext() to support gracefull shutdown.
func (server *Server) Run() error {
	path := "/"
	if server.options.EnableRandomUrl {
		path = "/" + randomstring.Generate(8) + "/"
	}

	srv := &http.Server{Handler: server.setupHandlers(path)}

	hostPort := "127.0.0.1:8080"
	listener, err := net.Listen("tcp", hostPort)
	if err != nil {
		return errors.Wrapf(err, "failed to listen at `%s`", hostPort)
	}

	scheme := "http"
	host, port, _ := net.SplitHostPort(listener.Addr().String())
	log.Printf("HTTP server is listening at: %s", scheme+"://"+host+":"+port+path)

	srvErr := make(chan error, 1)
	go func() {
		err = srv.Serve(listener)
		if err != nil {
			srvErr <- err
		}
	}()

	select {
	case err = <-srvErr:
		if err == http.ErrServerClosed { // by gracefull ctx
			err = nil
		}
	}

	return err
}

func (server *Server) setupHandlers(pathPrefix string) http.Handler {
	staticFileHandler := http.FileServer(
		&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, Prefix: "static"},
	)

	siteMux := http.NewServeMux()
	siteMux.HandleFunc(pathPrefix, server.handleIndex)
	siteMux.Handle(pathPrefix+"js/", http.StripPrefix(pathPrefix, staticFileHandler))
	siteMux.Handle(pathPrefix+"favicon.png", http.StripPrefix(pathPrefix, staticFileHandler))
	siteMux.Handle(pathPrefix+"css/", http.StripPrefix(pathPrefix, staticFileHandler))
	siteMux.HandleFunc(pathPrefix+"auth_token.js", server.handleAuthToken)
	siteMux.HandleFunc(pathPrefix+"config.js", server.handleConfig)

	wsMux := http.NewServeMux()
	wsMux.Handle("/", middleware.WrapLogger(middleware.WrapGzip(http.Handler(siteMux))))
	wsMux.HandleFunc(pathPrefix+"ws", server.generateHandleWS())

	return http.Handler(wsMux)
}
