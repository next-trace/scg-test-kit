# Public API — scg-test-kit

The following packages are considered PUBLIC and stable:

- github.com/next-trace/scg-test-kit (primary entry point)

All other packages are INTERNAL.
External imports from non-listed packages (including anything under `internal/`) are unsupported and may break without notice.

## Core API

### Harness Creation
- `func New(t testing.TB, opts ...Option) *Harness`
- `func NewHarness(t testing.TB, opts ...Option) *Harness` (Alias for New)
- `func NewUnitHarness(t testing.TB, opts ...Option) *Harness` (Semantic alias for New)
- `func NewIntegrationHarness(t testing.TB, opts ...Option) *Harness` (Semantic alias for New)
- `func NewBrowserHarness(t testing.TB, handler http.Handler, opts ...Option) *Harness`

### Resource Management
- `func WithResource(name string, value any, cleanup func() error) Option`
- `func Resource[T any](h *Harness, name string) (T, bool)`
- `func (h *Harness) RegisterCleanup(fn func())`
- `func (h *Harness) Cleanup()` (Idempotent, automatically called by `t.Cleanup`)
- `func (h *Harness) Close()` (Alias for Cleanup)

### HTTP Helpers
- `func WithHTTPServer(handler http.Handler) Option`
- `func Get(t testing.TB, h *Harness, path string, target any) *http.Response`
- `func Post(t testing.TB, h *Harness, path string, body any, target any)`

### JSON Helpers
- `func EncodeJSON(t testing.TB, value any) io.Reader`
- `func DecodeJSON(t testing.TB, reader io.Reader, target any)`
