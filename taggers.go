package sqlcomment

import (
	"context"
	"fmt"
	"runtime/debug"
)

// driverVersionTagger adds `db_driver` tag with "ent:<version>"
type driverVersionTagger struct {
	version string
}

func NewDriverVersionTagger() driverVersionTagger {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return driverVersionTagger{"ent"}
	}
	for _, d := range info.Deps {
		if d.Path == "entgo.io/ent" {
			return driverVersionTagger{fmt.Sprintf("ent:%s", d.Version)}
		}
	}
	return driverVersionTagger{"ent"}
}

func (dv driverVersionTagger) Tag(ctx context.Context) Tags {
	return Tags{
		KeyDBDriver: dv.version,
	}
}

type contextKey struct{}

// WithTag stores the key and val pair on the context.
// for example, if you want to add `route` tag to your SQL comment, put the url path on request context:
//	middleware := func(next http.Handler) http.Handler {
//		fn := func(w http.ResponseWriter, r *http.Request) {
//			c := sqlcomment.WithTag(r.Context(), "route", r.URL.Path)
//			next.ServeHTTP(w, r.WithContext(c))
//		}
//		return http.HandlerFunc(fn)
//	}
func WithTag(ctx context.Context, key, val string) context.Context {
	t, ok := ctx.Value(contextKey{}).(*Tags)
	if !ok {
		return context.WithValue(ctx, contextKey{}, &Tags{key: val})
	}
	tags := *t
	tags[key] = val
	return context.WithValue(ctx, contextKey{}, tags)
}

// FromContext returns the tags stored in ctx, if any.
func FromContext(ctx context.Context) Tags {
	t, ok := ctx.Value(contextKey{}).(*Tags)
	if !ok {
		return Tags{}
	}
	return *t
}

type contextTagger struct{}

func (ct contextTagger) Tag(ctx context.Context) Tags {
	return FromContext(ctx)
}

type staticTagger struct {
	tags Tags
}

// NewStaticTagger returns an Tagger which adds tags to every SQL comment.
func NewStaticTagger(tags Tags) staticTagger {
	return staticTagger{tags}
}

func (st staticTagger) Tag(ctx context.Context) Tags {
	return st.tags
}
