package router

// router.go - HTTP request routing and dispatching
//
// Responsibilities:
// - Build route table from configuration at startup
// - Match incoming requests to routes using longest-prefix matching
// - Support atomic route table replacement for hot-reload
// - Return 404 if no route matches request path
//
// Key Functions:
// - New(routes []Route) *Router: Create router with initial route table
// - Match(path string) (*Route, bool): Find matching route for request path (longest prefix wins)
// - Reload(routes []Route): Atomically replace route table (using sync/atomic pointer swap)
//
// Inputs:
// - HTTP request path (string)
// - Route configuration from config package
//
// Outputs:
// - Matched Route struct (or nil if no match)
// - bool indicating whether a match was found

type Router struct {
	// TODO: Implement router with atomic route table
}

// New creates a new router with the given routes
func New(routes []Route) *Router {
	// TODO: Implement router initialization
	return nil
}

// Match finds the route with the longest matching prefix
func (r *Router) Match(path string) (*Route, bool) {
	// TODO: Implement longest-prefix matching
	return nil, false
}
