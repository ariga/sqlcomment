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
