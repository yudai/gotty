package app

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"github.com/fatih/structs"
	"github.com/gorilla/websocket"
)

type clientContext struct {
	app        *App
	request    *http.Request
	connection *websocket.Conn
	command    *exec.Cmd
	pty        *os.File
	writeMutex *sync.Mutex
	cmdstring  string
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

type ContextVars struct {
	Command    string
	Pid        int
	Hostname   string
	RemoteAddr string
}

func (context *clientContext) addLog(single string, info string) {
	if context.app.options.EnableAduit == true {
		loggers, err := os.OpenFile(context.app.options.Aduit, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("Fail to find", *loggers, "cServer start Failed")
			os.Exit(1)
		}
		log.SetOutput(loggers)
		log.SetFlags(log.LstdFlags)
		hostname, _ := os.Hostname()
		if context.app.options.EnableBasicAuth == true {
			user := strings.Split(context.app.options.Credential, ":")
			log.Printf(user[0]+"@"+hostname+":"+single+"# %s", info)
		} else {
			log.Printf(hostname+":"+single+"# %s", info)
		}
	}
}

func (context *clientContext) goHandleClient() {
	exit := make(chan bool, 2)

	go func() {
		defer func() { exit <- true }()

		context.processSend()
	}()

	go func() {
		defer func() { exit <- true }()

		context.processReceive()
	}()

	go func() {
		defer context.app.server.FinishRoutine()

		<-exit
		context.pty.Close()

		// Even if the PTY has been closed,
		// Read(0 in processSend() keeps blocking and the process doen't exit
		context.command.Process.Signal(syscall.Signal(context.app.options.CloseSignal))

		context.command.Wait()
		context.connection.Close()
		context.app.connections--
		if context.app.options.MaxConnection != 0 {
			log.Printf("Connection closed: %s, connections: %d/%d",
				context.request.RemoteAddr, context.app.connections, context.app.options.MaxConnection)
		} else {
			log.Printf("Connection closed: %s, connections: %d",
				context.request.RemoteAddr, context.app.connections)
		}
	}()
}

func (context *clientContext) processSend() {
	if err := context.sendInitialize(); err != nil {
		log.Printf(err.Error())
		return
	}

	buf := make([]byte, 1024)

	for {
		size, err := context.pty.Read(buf)
		if err != nil {
			log.Printf("Command exited for: %s", context.request.RemoteAddr)
			return
		}
		if size > 1 && buf[:size][0] != 35 {
			context.addLog("Send", string(buf[:size]))
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
	hostname, _ := os.Hostname()
	titleVars := ContextVars{
		Command:    strings.Join(context.app.command, " "),
		Pid:        context.command.Process.Pid,
		Hostname:   hostname,
		RemoteAddr: context.request.RemoteAddr,
	}

	titleBuffer := new(bytes.Buffer)
	if err := context.app.titleTemplate.Execute(titleBuffer, titleVars); err != nil {
		return err
	}
	if err := context.write(append([]byte{SetWindowTitle}, titleBuffer.Bytes()...)); err != nil {
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

		if len(data[1:]) > 0 {
			if data[1:][0] != 13 {
				context.cmdstring += string(data[1:])
			} else {
				context.addLog("Receive", context.cmdstring)
				context.cmdstring = ""
			}
		}

		switch data[0] {
		case Input:
			if !context.app.options.PermitWrite {
				break
			}

			_, err := context.pty.Write(data[1:])
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

			window := struct {
				row uint16
				col uint16
				x   uint16
				y   uint16
			}{
				uint16(args.Rows),
				uint16(args.Columns),
				0,
				0,
			}
			syscall.Syscall(
				syscall.SYS_IOCTL,
				context.pty.Fd(),
				syscall.TIOCSWINSZ,
				uintptr(unsafe.Pointer(&window)),
			)

		default:
			log.Print("Unknown message type")
			return
		}
	}
}
