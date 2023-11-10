package radar

import (
	"fmt"
	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"github.com/vela-ssoc/vela-kit/iputil"
	"github.com/vela-ssoc/vela-kit/thread"
	"github.com/vela-ssoc/vela-radar/fingerprintx/plugins"
	"github.com/vela-ssoc/vela-radar/fingerprintx/scan"
	"github.com/vela-ssoc/vela-radar/host"
	"github.com/vela-ssoc/vela-radar/port"
	"github.com/vela-ssoc/vela-radar/port/syn"
	"github.com/vela-ssoc/vela-radar/util"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/netip"
	"sync"
	"testing"
	"time"
)

func TestScan(t *testing.T) {
	pn := false

	// 解析端口字符串并且优先发送 TopTcpPorts 中的端口, eg: 1-65535,top1000
	ports, err := port.ShuffleParseAndMergeTopPorts("1-65535")
	if err != nil {
		log.Fatal(err)
	}

	fn := func(tx port.OpenIpPort) {
		t.Logf("ip:%s port:%d", tx.Ip.String(), tx.Port)

		addr, _ := netip.AddrFromSlice(tx.Ip)

		target := plugins.Target{
			Address: netip.AddrPortFrom(addr, tx.Port),
			Host:    "localhost",
		}

		cfg := scan.Config{
			UDP:            false,
			FastMode:       false,
			DefaultTimeout: time.Second,
			Verbose:        false,
		}

		srv, err := scan.Do(target, cfg)
		if err != nil {
			t.Logf("callback fingerprint fail %v", err)
			return
		}

		if srv == nil {
			return
		}

		h := Host{
			IP:        tx.Ip,
			Port:      tx.Port,
			Protocol:  srv.Protocol,
			TLS:       srv.TLS,
			Transport: srv.Transport,
			Version:   srv.Version,
			Banner:    srv.Raw,
		}

		t.Logf("%s\n", h.String())
	}

	var wgFinger sync.WaitGroup
	fingerPool, _ := thread.NewPoolWithFunc(300, func(v interface{}) {
		tx := v.(port.OpenIpPort)
		fn(tx)
		wgFinger.Done()
	})

	call := func(v port.OpenIpPort) {
		wgFinger.Add(1)
		fingerPool.Invoke(v)
	}

	// parse ip
	it, startIp, _ := iputil.NewIter("172.31.61.0/24")

	// scanner
	ss, err := syn.NewSynScanner(startIp, call, port.Option{
		Rate:    1500,
		Timeout: 800,
	})

	// port scan func
	portScan := func(ip net.IP) {
		for _, _port := range ports { // port
			ss.WaitLimiter() // limit rate
			ss.Scan(ip, _port)
		}
	}

	// host group scan func
	var wgHostScan sync.WaitGroup
	hostScan, _ := thread.NewPoolWithFunc(200, func(ip interface{}) {
		_ip := ip.(net.IP)
		portScan(_ip)
		wgHostScan.Done()
	})
	defer hostScan.Release()

	var wgPing sync.WaitGroup

	// Pool - ping and port scan
	poolPing, _ := thread.NewPoolWithFunc(300, func(ip interface{}) {
		_ip := ip.(net.IP)
		if host.IsLive(_ip.String(), false, 800*time.Millisecond) {
			wgHostScan.Add(1)
			hostScan.Invoke(_ip)
		}
		wgPing.Done()
	})

	defer poolPing.Release()

	shuffle := util.NewShuffle(it.TotalNum())    // shuffle
	for i := uint64(0); i < it.TotalNum(); i++ { // ip index
		ip := make(net.IP, len(it.GetIpByIndex(0)))
		copy(ip, it.GetIpByIndex(shuffle.Get(i))) // Note: dup copy []byte when concurrent (GetIpByIndex not to do dup copy)
		if !pn {                                  // ping
			wgPing.Add(1)
			_ = poolPing.Invoke(ip)
		} else {
			wgHostScan.Add(1)
			hostScan.Invoke(ip)
		}
	}
	wgPing.Wait()
	wgHostScan.Wait()
	ss.Wait()
	ss.Close()
}

func TestWeb(t *testing.T) {
	resp, err := http.DefaultClient.Get("http://172.31.61.168:8081")
	if err != nil {
		log.Fatal(err)
	}
	data, _ := ioutil.ReadAll(resp.Body) // Ignoring error for example

	wappalyzerClient, err := wappalyzer.New()
	fingerprints := wappalyzerClient.Fingerprint(resp.Header, data)
	fmt.Printf("%v\n", fingerprints)

	// Output: map[Acquia Cloud Platform:{} Amazon EC2:{} Apache:{} Cloudflare:{} Drupal:{} PHP:{} Percona:{} React:{} Varnish:{}]
}
