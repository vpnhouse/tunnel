package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"github.com/vpnhouse/tunnel/pkg/xdns/dnsbase"
	"go.uber.org/zap"
)

const (
	pluginName = "blocklist"
)

// blocklistPlugin implements coredns' plugin.Handler interface
type blocklistPlugin struct {
	Next plugin.Handler

	mu   sync.Mutex
	looq dnsbase.LookupInterface
}

func (b *blocklistPlugin) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := &request.Request{W: w, Req: r}

	if b.mustPass(ctx, state.Name()) {
		return plugin.NextOrFailure(b.Name(), b.Next, ctx, w, r)
	}

	blockedCount.WithLabelValues(metrics.WithServer(ctx)).Inc()
	b.logBlock(state)

	m := &dns.Msg{}
	m.SetReply(r)
	// non-existent domain, recursion not available
	m.Rcode = dns.RcodeNameError
	m.RecursionAvailable = false
	m.RecursionDesired = false
	// m.Answer is empty - we have nothing to reply with
	if err := w.WriteMsg(m); err != nil {
		fmt.Println("XXX failed to write msg", err)
	}
	return dns.RcodeRefused, nil
}

func (b *blocklistPlugin) logBlock(r *request.Request) {
	q := r.Req.Question[0]
	// TODO(nikonov): optionally write blocklog to file
	zap.L().Warn("blocking request", zap.String("from", r.RemoteAddr()), zap.String("query", q.String()))
}

func (b *blocklistPlugin) mustPass(ctx context.Context, name string) bool {
	start := time.Now()
	defer func() {
		lookupDurationHist.WithLabelValues(metrics.WithServer(ctx)).Observe(float64(time.Since(start)))
	}()

	b.mu.Lock()
	defer b.mu.Unlock()
	v, _ := b.looq.Lookup(name)
	return v == nil // the name is not in block lists
}

func (*blocklistPlugin) Name() string {
	return pluginName
}

func (*blocklistPlugin) Ready() bool { return true }

func New(dbpath string) error {
	if isRegistered() {
		// the plugin has already been registered,
		// we don't want to replace it in the runtime (yet).
		return nil
	}

	rd, err := dnsbase.NewFTLReader(dbpath)
	if err != nil {
		return fmt.Errorf("failed to initialize blacklist db: %v", err)
	}

	dnsserver.Directives = append([]string{pluginName}, dnsserver.Directives...)
	setupFn := func(c *caddy.Controller) error {
		dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
			return &blocklistPlugin{Next: next, looq: rd}
		})
		return nil
	}
	plugin.Register(pluginName, setupFn)
	return nil
}

func isRegistered() bool {
	list := caddy.ListPlugins()
	expected := "dns." + pluginName
	for _, name := range list["others"] {
		if name == expected {
			return true
		}
	}
	return false
}
