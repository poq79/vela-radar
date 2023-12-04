package tcp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/vela-ssoc/vela-radar/port"
	limiter "golang.org/x/time/rate"
)

var DefaultTcpOption = port.Option{
	Rate:    1000,
	Timeout: 800,
}

type TcpScanner struct {
	ports    []uint16 // 指定端口
	callback func(port.OpenIpPort)
	limiter  *limiter.Limiter
	ctx      context.Context
	timeout  time.Duration
	isDone   bool
	option   port.Option
	wg       sync.WaitGroup
}

// NewTcpScanner Tcp扫描器
func NewTcpScanner(callback func(port.OpenIpPort), option port.Option) (ts *TcpScanner, err error) {
	// option verify
	if option.Rate < 10 {
		err = errors.New("rate can not set < 10")
		return
	}
	if option.Timeout <= 0 {
		err = errors.New("timeout can not set to 0")
		return
	}

	ts = &TcpScanner{
		callback: callback,
		limiter:  limiter.NewLimiter(limiter.Every(time.Second/time.Duration(option.Rate)), option.Rate/10),
		ctx:      context.Background(),
		timeout:  time.Duration(option.Timeout) * time.Millisecond,
		option:   option,
	}

	return
}

// Scan 对指定IP和dis port进行扫描
func (ts *TcpScanner) Scan(ip net.IP, dst uint16) error {
	if ts.isDone {
		return errors.New("scanner is closed")
	}
	ts.wg.Add(1)
	go func() {
		defer ts.wg.Done()
		//fmt.Println(1)
		openIpPort := port.OpenIpPort{
			Ip:   ip,
			Port: dst,
		}
		conn, _ := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, dst), ts.timeout)
		if conn != nil {
			conn.Close()
		} else {
			ts.callback(port.OpenIpPort{
				Ip:   nil,
				Port: 0,
			})
			return
		}
		ts.callback(openIpPort)
	}()
	return nil
}

/*
func (ts *TcpScanner) Scan(ip net.IP, dst uint16) error {
	if ts.isDone {
		return errors.New("scanner is closed")
	}
	//fmt.Println(1)
	openIpPort := port.OpenIpPort{
		Ip:   ip,
		Port: dst,
	}
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, dst), ts.timeout)
	if err != nil {
		return err
	}

	if conn == nil {
		return nil
	}
	conn.Close()
	ts.callback(openIpPort)
	return nil
}*/

func (ts *TcpScanner) Wait() {
	ts.wg.Wait()
}

// Close chan
func (ts *TcpScanner) Close() {
	ts.isDone = true
}

// WaitLimiter Waiting for the speed limit
func (ts *TcpScanner) WaitLimiter() error {
	return ts.limiter.Wait(ts.ctx)
}
