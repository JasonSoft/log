package gelf

import (
	"bufio"
	"fmt"
	stdlog "log"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/jasonsoft/log/v2"
)

// Gelf is an instance of the Gelf logger
type Gelf struct {
	mutex          sync.Mutex
	conn           net.Conn
	bufferedWriter *bufio.Writer
	url            *url.URL
	isActive       bool
}

// New create a new Gelf instance
func New(connectionString string) log.Handler {
	url, err := url.Parse(connectionString)
	if err != nil {
		panic(fmt.Errorf("graylog connectionString is wrong: %v", err))
	}
	g := &Gelf{
		url:      url,
		isActive: true,
	}
	g.manageConnections()
	return g
}

var empty byte

func (g *Gelf) close() error {
	if g.conn != nil {
		_ = g.conn.Close()
		g.conn = nil
	}
	return nil
}

// BeforeWriting handles the log entry
func (g *Gelf) BeforeWriting(e *log.Entry) error {
	e.Str("version", "1.1").
		Uint8("level", gelfLevel(e.Level)).
		Str("short_message", e.Message).
		Str("timestamp", time.Now().UTC().Format(time.RFC3339))

	e.Message = ""
	return nil
	//items["timestamp"] = float64(time.Now().UTC().UnixNano()) / float64(time.Second)
}

// Write handles the log entry
func (g *Gelf) Write(bytes []byte) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.conn != nil {
		_, err := g.bufferedWriter.Write(append(bytes, empty)) // when we use tcp, we need to add null byte in the end.
		if err != nil {
			_ = g.close()
			return fmt.Errorf("send log to graylog failed: %w", err)
		}

		// msg := fmt.Sprintf("payload size: %d", size)
		// println(msg)

		// msg = fmt.Sprintf("payload body: %s", string(payload))
		// println(msg)
	}

	return nil
}

// Flush all buffer data and close connection
func (g *Gelf) Flush() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	_ = g.bufferedWriter.Flush()
	g.isActive = false

	return g.close()
}

func (g *Gelf) manageConnections() {
	var err error
	if strings.EqualFold(g.url.Scheme, "tcp") {
		g.conn, err = net.Dial("tcp", g.url.Host)
		if err != nil {
			stdlog.Println("gelf tcp connection was failed:", err.Error())
		}
		g.bufferedWriter = bufio.NewWriter(g.conn)
	} else {
		g.conn, err = net.Dial("udp", g.url.Host)
		if err != nil {
			stdlog.Println("gelf udp connection was failed:", err.Error())
		}
		g.bufferedWriter = bufio.NewWriter(g.conn)
	}

	// check connection status every 1 second
	go func() {
		for {
			g.mutex.Lock()
			if g.isActive == false {
				return
			}
			if g.conn == nil {
				// TODO: tcp is hard-code at the point, we need to remove that later
				newConn, err := net.Dial("tcp", g.url.Host)
				if err != nil {
					stdlog.Printf("gelf: create connection failed: %v", err)
					continue
				}
				g.conn = newConn
				g.bufferedWriter = bufio.NewWriter(g.conn)
				stdlog.Println("gelf: created a connection")
			}
			_ = g.bufferedWriter.Flush()
			g.mutex.Unlock()
			time.Sleep(1 * time.Second)
		}
	}()
}

func gelfLevel(level log.Level) uint8 {
	switch level {
	case log.DebugLevel:
		return 7
	case log.InfoLevel:
		return 6
	case log.WarnLevel:
		return 4
	case log.ErrorLevel:
		return 3
	case log.FatalLevel:
		return 2
	default:
		return 1
	}
}
