# sqlcomm**ent**
sqlcomm**ent** is an [ent](https://entgo.io) driver that adds SQL comments following [sqlcommenter specification](https://google.github.io/sqlcommenter/spec/).  
sqlcomment includes support for OpenTelemetry and OpenCensus (see [examples](examples/)).

# Installing
```bash
go install ariga.io/sqlcomment
```

# Basic Usage
```go
import (
  "ariga.io/sqlcomment"
  "entgo.io/ent/dialect/sql"
)
// Create db driver.
db, err := sql.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
if err != nil {
  log.Fatalf("Failed to connect to database: %v", err)
}
// create sqlcomment driver which wraps sqlite driver.
drv := sqlcomment.NewDriver(db),
  sqlcomment.WithDriverVerTag(),
  sqlcomment.WithTags(sqlcomment.Tags{
    sqlcomment.KeyApplication: "my-app",
    sqlcomment.KeyFramework:   "net/http",
  }),
)
// create and configure ent client
client := ent.NewClient(ent.Driver(drv))
```

# Adding context level tags
Suppose you have a REST API and you want to add a tag with the request URL (typically `route` tag). You can achieve that by using `sqlcomment.WithTag(ctx, key, val)` which adds the given tag to the context to later be serialized by sqlcomment driver. (see full example [here](examples/otel/example_test.go))
```go
// Add a middleware to your HTTP server which puts the `route` tag on the context for every request.
middleware := func(next http.Handler) http.Handler {
  fn := func(w http.ResponseWriter, r *http.Request) {
    ctx := sqlcomment.WithTag(r.Context(), "route", r.URL.Path)
    next.ServeHTTP(w, r.WithContext(ctx))
  }
  return http.HandlerFunc(fn)
}
```