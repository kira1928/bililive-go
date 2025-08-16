package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/bililive-go/bililive-go/src/instance"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
)

// Client connects to remote server over WebSocket and proxies HTTP traffic to local web UI.
type Client struct {
	serverAddr string
	token      string
	localAddr  string
}

// NewClient creates a streaming proxy client.
func NewClient(serverAddr, token, localAddr string) *Client {
	return &Client{serverAddr: serverAddr, token: token, localAddr: localAddr}
}

type muxListener struct {
	session *yamux.Session
	addr    net.Addr
}

func newMuxListener(session *yamux.Session, addr net.Addr) net.Listener {
	return &muxListener{session: session, addr: addr}
}

func (l *muxListener) Accept() (net.Conn, error) {
	return l.session.Accept()
}
func (l *muxListener) Close() error {
	return l.session.Close()
}
func (l *muxListener) Addr() net.Addr {
	return l.addr
}

var ErrListenerClosed = fmt.Errorf("listener closed")

// Run establishes WebSocket and serves HTTP proxy until disconnected, then retries.
func (c *Client) Run(ctx context.Context) {
	inst := instance.GetInstance(ctx)
	for {
		// 检查配置的 remote mode 状态，动态开关
		if !inst.Config.RemoteMode {
			inst.Logger.Info("remote mode disabled, skip WS connect")
			time.Sleep(5 * time.Second)
			continue
		}
		// determine WebSocket URL: support ws://, wss://, http://, https:// or plain host
		var wsURL url.URL
		if u, err := url.Parse(c.serverAddr); err == nil && u.Scheme != "" {
			// use provided scheme or map http->ws, https->wss
			switch u.Scheme {
			case "ws", "wss":
				wsURL = *u
			case "http":
				u.Scheme = "ws"
				wsURL = *u
			case "https":
				u.Scheme = "wss"
				wsURL = *u
			default:
				wsURL = url.URL{Scheme: "ws", Host: u.Host, Path: u.Path}
			}
		} else {
			wsURL = url.URL{Scheme: "ws", Host: c.serverAddr, Path: "/ws"}
		}
		wsURL.Path = "/ws"
		inst.Logger.Infof("dialing %s", wsURL.String())
		// Dial WebSocket without headers
		ws, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
		if err != nil {
			inst.Logger.Errorf("dial error: %v, retry in 5s", err)
			time.Sleep(5 * time.Second)
			continue
		}
		// Send token as first text message for authentication
		if err := ws.WriteMessage(websocket.TextMessage, []byte(c.token)); err != nil {
			inst.Logger.Errorf("auth send error: %v", err)
			ws.Close()
			time.Sleep(5 * time.Second)
			continue
		}
		inst.Logger.Info("auth token sent, waiting for server response")
		// Read server authentication response
		msgType, msg, err := ws.ReadMessage()
		if err != nil || msgType != websocket.TextMessage {
			inst.Logger.Errorf("failed to read auth response: %v", err)
			ws.Close()
			time.Sleep(5 * time.Second)
			continue
		}
		resp := string(msg)
		if resp != "OK" {
			inst.Logger.Errorf("authentication failed: %s", resp)
			ws.Close()
			time.Sleep(5 * time.Second)
			continue
		}
		inst.Logger.Info("authentication succeeded, starting proxy")

		rawConn := ws.UnderlyingConn()

		session, err := yamux.Server(rawConn, nil)
		if err != nil {
			inst.Logger.Errorf("failed to create yamux session: %v", err)
			ws.Close()
			time.Sleep(5 * time.Second)
			continue
		}

		target, _ := url.Parse("http://" + c.localAddr)
		proxy := httputil.NewSingleHostReverseProxy(target)
		listener := newMuxListener(session, rawConn.LocalAddr())

		// Serve 会不断 Accept yamux 子流，并在每个子流上处理 HTTP
		if err := http.Serve(listener, proxy); err != nil && err != io.EOF {
			inst.Logger.Println("http.Serve 结束：", err)
		}

		inst.Logger.Info("disconnected, retrying in 5s")
		time.Sleep(5 * time.Second)
	}
}
