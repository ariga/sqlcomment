package sqlcomment

import (
	"context"

	"go.opentelemetry.io/otel"
)

type (
	// OTELTagger is a Tagger that adds `traceparent` and `tracestate` tags to the SQL comment.
	OTELTagger struct{}
	// CommentCarrier implements propagation.TextMapCarrier in order to retrieve trace information from OTEL.
	CommentCarrier Tags
)

// NewOTELTagger adds OTEL trace information as SQL tags.
func NewOTELTagger() OTELTagger {
	return OTELTagger{}
}

// Tag finds trace information on the given context and returns SQL tags with trace information.
func (ot OTELTagger) Tag(ctx context.Context) Tags {
	c := NewCommentCarrier()
	otel.GetTextMapPropagator().Inject(ctx, c)
	return Tags(c)
}

func NewCommentCarrier() CommentCarrier {
	return make(CommentCarrier)
}

// Get returns the value associated with the passed key.
func (c CommentCarrier) Get(key string) string {
	return string(c[key])
}

// Set stores the key-value pair.
func (c CommentCarrier) Set(key string, value string) {
	c[key] = value
}

// Keys lists the keys stored in this carrier.
func (c CommentCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, string(k))
	}
	return keys
}
