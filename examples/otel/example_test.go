package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"

	"ariga.io/sqlcomment"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	_ "github.com/mattn/go-sqlite3"

	"ariga.io/sqlcomment/examples/ent"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type CustomCommenter struct{}

func (mcc CustomCommenter) Tag(ctx context.Context) sqlcomment.Tags {
	return sqlcomment.Tags{
		"key": "value",
	}
}

func Example_OTELIntegration() {
	tp := initTracer()
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// Create db driver.
	db, err := sql.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	commentedDriver := sqlcomment.NewDriver(dialect.Debug(db),
		sqlcomment.WithTagger(
			// add tracing info with Open Telemetry.
			sqlcomment.NewOTELTagger(),
			// use your custom commenter
			CustomCommenter{},
		),
		// add `db_driver` version tag
		sqlcomment.WithDriverVerTag(),
		// add some global tags to all queries
		sqlcomment.WithTags(sqlcomment.Tags{
			sqlcomment.KeyApplication: "bootcamp",
			sqlcomment.KeyFramework:   "go-chi",
		}))
	// create and configure ent client
	client := ent.NewClient(ent.Driver(commentedDriver))
	defer client.Close()
	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	client.User.Create().SetName("hedwigz").SaveX(context.Background())

	// An HTTP middleware that adds the URL path to sqlcomment tags, under the key "route".
	middleware := func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := sqlcomment.WithTag(r.Context(), "route", r.URL.Path)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
	// Application-level http handler.
	getUsersHandler := func(rw http.ResponseWriter, r *http.Request) {
		users := client.User.Query().AllX(r.Context())
		b, _ := json.Marshal(users)
		rw.WriteHeader(http.StatusOK)
		rw.Write(b)
	}

	backend := otelhttp.NewHandler(middleware(http.HandlerFunc(getUsersHandler)), "app")
	testRequest(backend)
}

func initTracer() *sdktrace.TracerProvider {
	exporter, err := stdout.New(stdout.WithWriter(io.Discard))
	if err != nil {
		log.Fatal(err)
	}
	// For the demonstration, use sdktrace.AlwaysSample sampler to sample all traces.
	// In a production application, use sdktrace.ProbabilitySampler with a desired probability.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String("ExampleService"))),
	)
	otel.SetTracerProvider(tp)
	// Add propagation.TaceContext{} which will be used by OtelTagger to inject trace information.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp
}

func testRequest(handler http.Handler) {
	req := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	w := httptest.NewRecorder()

	// Debug printer should print SQL statement with comment.
	handler.ServeHTTP(w, req)
}
