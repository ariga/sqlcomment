package sqlcomment

import (
	"context"
)

type (
	// ctxOptions allows injecting runtime options.
	ctxOptions struct {
		skip bool // i.e. skip entry.
		tags Tags
	}
	ctxKeyType struct{}
)

var ctxOptionsKey ctxKeyType

// Skip returns a new Context that tells the Driver
// to skip the commenting on Query.
//
//	client.T.Query().All(sqlcomment.Skip(ctx))
//
func Skip(ctx context.Context) context.Context {
	c, ok := ctx.Value(ctxOptionsKey).(*ctxOptions)
	if !ok {
		return context.WithValue(ctx, ctxOptionsKey, &ctxOptions{skip: true})
	}
	c.skip = true
	return ctx
}

// WithTag stores the key and val pair on the context.
// for example, if you want to add `route` tag to your SQL comment, put the url path on request context:
//	middleware := func(next http.Handler) http.Handler {
//		fn := func(w http.ResponseWriter, r *http.Request) {
//			ctx := sqlcomment.WithTag(r.Context(), "route", r.URL.Path)
//			next.ServeHTTP(w, r.WithContext(ctx))
//		}
//		return http.HandlerFunc(fn)
//	}
func WithTag(ctx context.Context, key, val string) context.Context {
	t, ok := ctx.Value(ctxOptionsKey).(*ctxOptions)
	if !ok {
		return context.WithValue(ctx, ctxOptionsKey, &ctxOptions{tags: Tags{key: val}})
	}
	t.tags[key] = val
	return ctx
}

// FromContext returns the tags stored in ctx, if any.
func FromContext(ctx context.Context) Tags {
	t, ok := ctx.Value(ctxOptionsKey).(*ctxOptions)
	if !ok {
		return Tags{}
	}
	return t.tags
}
