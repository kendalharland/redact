package server

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"sync"
	"time"
)

const (
	// userExpiry is how long to keep user state after last activity.
	userExpiry = 1 * time.Hour
	// cleanupInterval is how often to clean up expired users.
	cleanupInterval = 5 * time.Minute
)

// userState tracks the state of a single user's requests.
type userState struct {
	mu          sync.Mutex
	processing  bool                    // true if a request is being processed
	pending     chan func()             // channel for pending request (size 1)
	lastActive  time.Time
}

// RateLimiter limits requests per user based on IP + user agent.
type RateLimiter struct {
	mu    sync.Mutex
	users map[string]*userState
	done  chan struct{}
}

// NewRateLimiter creates a new RateLimiter.
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		users: make(map[string]*userState),
		done:  make(chan struct{}),
	}
	go rl.cleanupLoop()
	return rl
}

// Stop stops the rate limiter's cleanup goroutine.
func (rl *RateLimiter) Stop() {
	close(rl.done)
}

// cleanupLoop periodically removes expired users.
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.done:
			return
		}
	}
}

// cleanup removes users that haven't been active recently.
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, state := range rl.users {
		state.mu.Lock()
		expired := now.Sub(state.lastActive) > userExpiry && !state.processing
		state.mu.Unlock()

		if expired {
			delete(rl.users, key)
		}
	}
}

// getUserKey generates a unique key for a user based on IP and user agent.
func getUserKey(r *http.Request) string {
	ip := getClientIP(r)
	ua := r.UserAgent()

	hash := sha256.Sum256([]byte(ip + "|" + ua))
	return hex.EncodeToString(hash[:16])
}

// getClientIP extracts the client IP from the request.
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain
		if idx := len(xff); idx > 0 {
			for i, c := range xff {
				if c == ',' {
					return xff[:i]
				}
			}
			return xff
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// getOrCreateUser gets or creates a user state for the given request.
func (rl *RateLimiter) getOrCreateUser(r *http.Request) *userState {
	key := getUserKey(r)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	state, ok := rl.users[key]
	if !ok {
		state = &userState{
			pending:    make(chan func(), 1),
			lastActive: time.Now(),
		}
		rl.users[key] = state

		// Start the worker goroutine for this user
		go rl.userWorker(key, state)
	}

	return state
}

// userWorker processes requests for a single user sequentially.
func (rl *RateLimiter) userWorker(key string, state *userState) {
	for {
		select {
		case fn := <-state.pending:
			state.mu.Lock()
			state.processing = true
			state.mu.Unlock()

			fn()

			state.mu.Lock()
			state.processing = false
			state.lastActive = time.Now()
			state.mu.Unlock()

		case <-rl.done:
			return
		}
	}
}

// Wrap wraps an HTTP handler with rate limiting.
// It ensures only one request per user is processed at a time.
// If a request is already being processed, subsequent requests are queued.
// The queue size is 1, with the most recent request taking priority.
func (rl *RateLimiter) Wrap(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := rl.getOrCreateUser(r)

		// Create a channel to signal completion
		done := make(chan struct{})

		// Create the work function
		work := func() {
			defer close(done)
			handler(w, r)
		}

		// Try to enqueue the work
		select {
		case state.pending <- work:
			// Successfully enqueued, wait for completion
			<-done
		default:
			// Queue is full, replace the pending request
			// First, drain the existing pending request
			select {
			case <-state.pending:
				// Drained old request
			default:
				// Nothing to drain
			}

			// Now enqueue our request
			state.pending <- work
			<-done
		}
	}
}
