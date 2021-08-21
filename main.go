package main

import (
	"errors"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

var clientAddress string
var clientNetwork string
var dialer = &net.Dialer{
	Timeout:   5 * time.Second,
	DualStack: true,
	KeepAlive: time.Minute,
}
var defaultTimeout = 5 * time.Minute

func copyWithTimeout(at, bt time.Duration, a, b net.Conn) error {
	buf := make([]byte, 1024)
	for {
		a.SetReadDeadline(time.Now().Add(at))
		n, err := a.Read(buf)
		if n != 0 {
			buf = buf[:n]
			b.SetWriteDeadline(time.Now().Add(bt))
			n2, err := b.Write(buf)
			if err != nil {
				return err
			}
			if n != n2 {
				return io.ErrShortWrite
			}
		}
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}
	}
}

func handleConn(a net.Conn) {
	b, err := dialer.Dial(clientNetwork, clientAddress)
	if err != nil {
		a.Close()
		log.Println(err)
		return
	}
	defer func() {
		time.Sleep(time.Second)
		a.Close()
		b.Close()
	}()
	done := make(chan struct{}, 1)
	var doneFlag int32 = 0
	go func() {
		err := copyWithTimeout(defaultTimeout, defaultTimeout, a, b)
		if err != nil && err != io.EOF {
			log.Println(err)
		}
		if atomic.CompareAndSwapInt32(&doneFlag, 0, 1) {
			close(done)
		}
	}()
	go func() {
		err := copyWithTimeout(defaultTimeout, defaultTimeout, b, a)
		if err != nil && err != io.EOF {
			log.Println(err)
		}
		if atomic.CompareAndSwapInt32(&doneFlag, 0, 1) {
			close(done)
		}
	}()
	<-done
}

func main() {
	if len(os.Args) < 3 {
		log.Println("Usage:", os.Args[0], " address-from address-to [timeout]\r\n\r\nEg:\r\n", os.Args[0], "127.0.0.1:9222 unix:/tmp/chrome-run/.devtools.sock 5m\r\n", os.Args[0], "unix:/tmp/chrome-run/.devtools.sock 127.0.0.1:9222 5m")
		os.Exit(1)
		return
	}
	var tcpAddress string
	var listenAddress string
	var listenNetwork string
	if strings.HasPrefix(os.Args[1], "unix:") {
		// listen to unix then connect tcp
		listenAddress = os.Args[1][5:]
		clientAddress = os.Args[2]
		listenNetwork = "unix"
		clientNetwork = "tcp"
		tcpAddress = clientAddress
	} else if strings.HasPrefix(os.Args[2], "unix:") {
		// listen tcp then connect unix
		listenAddress = os.Args[1]
		clientAddress = os.Args[2][5:]
		listenNetwork = "tcp"
		clientNetwork = "unix"
		tcpAddress = listenAddress
	} else {
		log.Println("Read no unix: prefix")
		os.Exit(1)
		return
	}

	if len(os.Args) > 3 {
		var err error
		defaultTimeout, err = time.ParseDuration(os.Args[3])
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		log.Println("Timeout:", defaultTimeout)
	}

	_, _, err := net.SplitHostPort(tcpAddress)
	if err != nil {
		log.Println(err)
		return
	}

	if listenNetwork == "unix" {
		os.Remove(listenAddress)
	}
	log.Println("Listening", listenAddress)
	ln, err := net.Listen(listenNetwork, listenAddress)
	if err != nil {
		log.Println(err)
		os.Exit(1)
		return
	}
	if listenNetwork == "unix" {
		os.Chmod(listenAddress, os.ModePerm)
	}
	for {
		conn, e := ln.Accept()
		if e != nil {
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				log.Println("Accept temp err:", ne)
				continue
			}
			log.Println("Accept err:", e)
			return
		}
		go handleConn(conn)
	}
}
