package api

import (
	"context"

	"github.com/go-chi/chi/v5"
)

// chiRouteCtx returns a context with chi URL param injected so we can
// drive handlers without spinning up a full router.
func chiRouteCtx(key, val string) context.Context {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, val)
	return context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
}
