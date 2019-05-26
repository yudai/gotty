package server

import (
	"html/template"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"modernc.org/httpfs"

	"github.com/yudai/gotty/server/assets"
	"github.com/yudai/gotty/server/middleware"
	"github.com/yudai/gotty/wetty"
)

// Server provides a wetty HTTP endpoint.
type Server struct {
	factory       Factory
	upgrader      *websocket.Upgrader
	indexTemplate *template.Template
}

// New creates a new instance of Server.
// Server will use the New() of the factory provided to handle each request.
func New(factory Factory) (*Server, error) {
	indexData, ok := assets.Assets["/index.html"]
	if !ok {
		panic("index not found") // must be in assets.Assets
	}
	indexTemplate, err := template.New("index").Parse(indexData)
	if err != nil {
		panic("index template parse failed") // must be valid
	}

	return &Server{
		factory: factory,
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			Subprotocols:    wetty.Protocols,
		},
		indexTemplate: indexTemplate,
	}, nil
}

// Run starts the main process of the Server.
// The cancelation of ctx will shutdown the server immediately with aborting
// existing connections. Use WithGracefullContext() to support gracefull shutdown.
func (server *Server) Run() error {
	scheme, host, port, path := "http", "127.0.0.1", "8080", "/"
	hostPort := host + ":" + port
	url := scheme + "://" + hostPort + path

	srv := &http.Server{Handler: server.setupHandlers(path)}

	listener, err := net.Listen("tcp", hostPort)
	if err != nil {
		return err // ors.Wrapf(err, "failed to listen at `%s`", hostPort)
	}

	log.Printf("HTTP server is listening at: %s", url)
	return srv.Serve(listener)
}

func (server *Server) setupHandlers(pathPrefix string) http.Handler {
	staticFileHandler := http.FileServer(httpfs.NewFileSystem(assets.Assets, time.Now()))

	siteMux := http.NewServeMux()
	siteMux.Handle(pathPrefix+"js/", http.StripPrefix(pathPrefix, staticFileHandler))
	siteMux.Handle(pathPrefix+"css/", http.StripPrefix(pathPrefix, staticFileHandler))
	siteMux.Handle(pathPrefix+"favicon.png", http.StripPrefix(pathPrefix, staticFileHandler))
	siteMux.HandleFunc(pathPrefix, server.handleIndex)
	siteMux.HandleFunc(pathPrefix+"auth_token.js", server.handleAuthToken)
	siteMux.HandleFunc(pathPrefix+"config.js", server.handleConfig)

	wsMux := http.NewServeMux()
	wsMux.Handle("/", middleware.WrapLogger(middleware.WrapGzip(http.Handler(siteMux))))
	wsMux.HandleFunc(pathPrefix+"ws", server.generateHandleWS())

	return http.Handler(wsMux)
}
