package app

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"strings"
	"sync"

	"github.com/fatih/structs"
	"github.com/gorilla/websocket"
	"github.com/yudai/gotty/backends"
)

type clientContext struct {
	backends.ClientContext
	app        *App
	connection *websocket.Conn
	writeMutex *sync.Mutex
}

const (
	Input          = '0'
	Ping           = '1'
	ResizeTerminal = '2'
)

const (
	Output         = '0'
	Pong           = '1'
	SetWindowTitle = '2'
	SetPreferences = '3'
	SetReconnect   = '4'
)

type argResizeTerminal struct {
	Columns float64
	Rows    float64
}

func (context *clientContext) goHandleClient() {
	exit := make(chan bool, 3)

	context.Start(exit)

	go func() {
		defer func() { exit <- true }()
		context.processSend()
	}()

	go func() {
		defer func() { exit <- true }()
		context.processReceive()
	}()

	<-exit
	context.TearDown()
}

func (context *clientContext) processSend() {
	if err := context.sendInitialize(); err != nil {
		log.Printf(err.Error())
		return
	}

	buf := make([]byte, 1024)

	for {
		size, err := context.OutputReader().Read(buf)
		if err != nil {
			log.Printf("failed to read output from terminal backend: %v", err)
			return
		}
		safeMessage := base64.StdEncoding.EncodeToString([]byte(buf[:size]))
		if err = context.write(append([]byte{Output}, []byte(safeMessage)...)); err != nil {
			log.Printf(err.Error())
			return
		}
	}
}

func (context *clientContext) write(data []byte) error {
	context.writeMutex.Lock()
	defer context.writeMutex.Unlock()
	return context.connection.WriteMessage(websocket.TextMessage, data)
}

func (context *clientContext) sendInitialize() error {
	windowTitle, err := context.WindowTitle()
	if err != nil {
		return err
	}
	if err := context.write(append([]byte{SetWindowTitle}, []byte(windowTitle)...)); err != nil {
		return err
	}

	prefStruct := structs.New(context.app.options.Preferences)
	prefMap := prefStruct.Map()
	htermPrefs := make(map[string]interface{})
	for key, value := range prefMap {
		rawKey := prefStruct.Field(key).Tag("hcl")
		if _, ok := context.app.options.RawPreferences[rawKey]; ok {
			htermPrefs[strings.Replace(rawKey, "_", "-", -1)] = value
		}
	}
	prefs, err := json.Marshal(htermPrefs)
	if err != nil {
		return err
	}

	if err := context.write(append([]byte{SetPreferences}, prefs...)); err != nil {
		return err
	}
	if context.app.options.EnableReconnect {
		reconnect, _ := json.Marshal(context.app.options.ReconnectTime)
		if err := context.write(append([]byte{SetReconnect}, reconnect...)); err != nil {
			return err
		}
	}
	return nil
}

func (context *clientContext) processReceive() {
	for {
		_, data, err := context.connection.ReadMessage()
		if err != nil {
			log.Print(err.Error())
			return
		}
		if len(data) == 0 {
			log.Print("An error has occured")
			return
		}

		switch data[0] {
		case Input:
			if !context.app.options.PermitWrite {
				break
			}

			_, err := context.InputWriter().Write(data[1:])
			if err != nil {
				return
			}

		case Ping:
			if err := context.write([]byte{Pong}); err != nil {
				log.Print(err.Error())
				return
			}
		case ResizeTerminal:
			var args argResizeTerminal
			err = json.Unmarshal(data[1:], &args)
			if err != nil {
				log.Print("Malformed remote command")
				return
			}
			rows := uint16(context.app.options.Height)
			if rows == 0 {
				rows = uint16(args.Rows)
			}

			columns := uint16(context.app.options.Width)
			if columns == 0 {
				columns = uint16(args.Columns)
			}

			context.ResizeTerminal(columns, rows)
		default:
			log.Print("Unknown message type")
			return
		}
	}
}
