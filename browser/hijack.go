package browser

import (
	"log"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// ensureHijacked sets up request interception on a tab if a policy is configured
// and the tab hasn't been hijacked yet.
func (b *Browser) ensureHijacked(t *tab) {
	if b.policy == nil || t.hijackRouter != nil {
		return
	}
	b.setupHijackRouter(t)
}

// setupHijackRouter configures request interception on a tab's page.
// Each deny rule's URL pattern is registered as a CDP Fetch pattern so Chrome
// only pauses requests matching deny URLs. The handler evaluates the full
// policy (deny + allow) and aborts or continues accordingly.
func (b *Browser) setupHijackRouter(t *tab) {
	router := t.page.HijackRequests()

	patterns := b.policy.Network.DenyPatterns()

	handler := func(h *rod.Hijack) {
		allowed, rule, err := b.policy.Network.IsAllowed(h.Request.Req())
		if err != nil {
			log.Printf("policy error: %v", err)
			h.ContinueRequest(&proto.FetchContinueRequest{})
			return
		}
		if allowed {
			h.ContinueRequest(&proto.FetchContinueRequest{})
			return
		}
		log.Printf("blocked by policy: %s %s (rule: %s)", h.Request.Method(), h.Request.URL(), rule)
		t.sendErr(&ErrBlocked{
			Method: h.Request.Method(),
			URL:    h.Request.URL().String(),
			Rule:   rule.String(),
		})
		h.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
	}

	for _, pattern := range patterns {
		if err := router.Add(pattern, "", handler); err != nil {
			log.Printf("hijack router add %q: %v", pattern, err)
		}
	}

	go router.Run()
	t.hijackRouter = router
}
