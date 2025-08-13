package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/template"
)

type Generator struct {
	ProjectRoot string
	Feature     string
}

type QueryInfo struct {
	Name        string
	Type        string // :one, :many, :exec
	HTTPMethod  string
	URLPath     string
	HandlerName string
	ServiceName string
	Params      []ParamInfo
	ReturnType  string
	SQLComment  string
}

type ParamInfo struct {
	Name string
	Type string
}

type APIGenerationData struct {
	Feature        string
	Package        string
	Queries        []QueryInfo
	Imports        []string
	HasHealthCheck bool
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	feature := os.Args[1]
	generator := &Generator{
		ProjectRoot: ".",
		Feature:     feature,
	}

	fmt.Printf("🔄 Generating API for feature: %s (Windows)\n", feature)

	if err := generator.Generate(); err != nil {
		fmt.Printf("❌ Error generating API: %v\n", err)
		os.Exit(1)
	}

	printSuccess(feature)
}

func printHelp() {
	fmt.Println("🚀 Civil Registry API Generator (Windows)")
	fmt.Println("==========================================")
	fmt.Println("")
	fmt.Println("Usage: go run cmd/genapi/main.go <feature-name>")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/genapi/main.go civil")
	fmt.Println("  go run cmd/genapi/main.go user")
	fmt.Println("  go run cmd/genapi/main.go document")
	fmt.Println("")
	fmt.Println("💡 Tip: Use 'task gen-api <feature>' for easier usage!")
}

func printSuccess(feature string) {
	fmt.Printf("✅ Successfully generated API files for feature: %s\n", feature)
	fmt.Println("")
	fmt.Println("📁 Generated files:")
	fmt.Printf("   📄 internal\\generated\\api\\%s\\handlers.go    - HTTP handlers with Swagger docs\n", feature)
	fmt.Printf("   📄 internal\\generated\\api\\%s\\service.go     - Business logic layer\n", feature)
	fmt.Printf("   📄 internal\\generated\\api\\%s\\router.go      - Chi router configuration\n", feature)
	fmt.Println("")
	fmt.Println("🎯 Next steps:")
	fmt.Printf("   1. Add router to main router: %s.%sRouter(queries, log)\n", feature, strings.Title(feature))
	fmt.Printf("   2. Run: task dev\n")
	fmt.Printf("   3. Test: curl http://localhost:8080/api/%s/health\n", feature)
	fmt.Println("")
	if runtime.GOOS == "windows" {
		fmt.Println("💡 Windows Tip: Use PowerShell or Windows Terminal for best experience!")
	}
}

func (g *Generator) Generate() error {
	// Parse queries from repository
	queries, err := g.parseQueries()
	if err != nil {
		return fmt.Errorf("failed to parse queries: %w", err)
	}

	if len(queries) == 0 {
		fmt.Printf("⚠️  No queries found for feature '%s'\n", g.Feature)
		fmt.Println("💡 Make sure you have:")
		fmt.Printf("   - SQL file: queries\\%s.sql\n", g.Feature)
		fmt.Printf("   - Generated repo: internal\\repository\\%s.sql.go\n", g.Feature)
		fmt.Println("   - Run 'task gen' first to generate repository code")
		return nil
	}

	fmt.Printf("📊 Found %d queries for feature '%s'\n", len(queries), g.Feature)
	for _, q := range queries {
		fmt.Printf("   🔹 %s %s → %s\n", q.HTTPMethod, q.URLPath, q.Name)
	}

	// Create API directory (Windows-compatible) - UPDATED PATH
	apiDir := filepath.Join("internal", "generated", "api", g.Feature)
	if err := os.MkdirAll(apiDir, 0755); err != nil {
		return fmt.Errorf("failed to create API directory: %w", err)
	}

	// Prepare generation data
	data := APIGenerationData{
		Feature:        g.Feature,
		Package:        g.Feature,
		Queries:        queries,
		Imports:        g.getRequiredImports(queries),
		HasHealthCheck: g.needsHealthCheck(queries),
	}

	// Generate files
	if err := g.generateHandlers(data); err != nil {
		return fmt.Errorf("failed to generate handlers: %w", err)
	}

	if err := g.generateService(data); err != nil {
		return fmt.Errorf("failed to generate service: %w", err)
	}

	if err := g.generateRouter(data); err != nil {
		return fmt.Errorf("failed to generate router: %w", err)
	}

	return nil
}

// Rest of the parsing functions remain the same...
func (g *Generator) parseQueries() ([]QueryInfo, error) {
	var queries []QueryInfo

	// Parse generated repository files - UPDATED PATH
	repoDir := filepath.Join("internal", "generated", "repository") // ✅ Updated to match your structure
	pattern := filepath.Join(repoDir, "*.sql.go")

	fmt.Printf("🔍 Looking for repository files in: %s\n", pattern)

	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	fmt.Printf("📁 Found %d repository files: %v\n", len(files), files)

	for _, file := range files {
		// Skip system files
		basename := filepath.Base(file)
		if strings.Contains(basename, "db.go") || strings.Contains(basename, "models.go") {
			fmt.Printf("⏭️  Skipping system file: %s\n", basename)
			continue
		}

		fmt.Printf("📖 Parsing file: %s\n", file)

		fileQueries, err := g.parseRepositoryFile(file)
		if err != nil {
			fmt.Printf("⚠️  Warning: failed to parse %s: %v\n", file, err)
			continue
		}

		fmt.Printf("📋 Found %d queries in %s\n", len(fileQueries), basename)

		// Filter queries by feature
		for _, q := range fileQueries {
			if g.isRelevantQuery(q.Name, basename) {
				fmt.Printf("✅ Including query: %s (from %s)\n", q.Name, basename)
				queries = append(queries, q)
			} else {
				fmt.Printf("❌ Excluding query: %s (from %s) - doesn't match feature '%s'\n", q.Name, basename, g.Feature)
			}
		}
	}

	fmt.Printf("🎯 Final result: %d relevant queries for feature '%s'\n", len(queries), g.Feature)
	return queries, nil
}
func (g *Generator) parseRepositoryFile(filename string) ([]QueryInfo, error) {
	var queries []QueryInfo

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// Parse functions
	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Recv != nil && len(fn.Recv.List) > 0 {
				// This is a method
				if recv, ok := fn.Recv.List[0].Type.(*ast.StarExpr); ok {
					if ident, ok := recv.X.(*ast.Ident); ok && ident.Name == "Queries" {
						query := g.parseQueryFunction(fn)
						if query.Name != "" {
							queries = append(queries, query)
						}
					}
				}
			}
		}
		return true
	})

	return queries, nil
}

func (g *Generator) parseQueryFunction(fn *ast.FuncDecl) QueryInfo {
	query := QueryInfo{
		Name: fn.Name.Name,
	}

	// Determine query type and HTTP mapping
	if strings.HasPrefix(query.Name, "Get") || strings.HasPrefix(query.Name, "List") || strings.HasPrefix(query.Name, "Find") {
		if strings.Contains(query.Name, "ByID") || strings.Contains(query.Name, "ById") {
			query.Type = ":one"
		} else {
			query.Type = ":many"
		}
		query.HTTPMethod = "GET"
		query.HandlerName = query.Name
		query.ServiceName = query.Name
	} else if strings.HasPrefix(query.Name, "Create") || strings.HasPrefix(query.Name, "Insert") {
		query.Type = ":one"
		query.HTTPMethod = "POST"
		query.HandlerName = query.Name
		query.ServiceName = query.Name
	} else if strings.HasPrefix(query.Name, "Update") {
		query.Type = ":one"
		query.HTTPMethod = "PUT"
		query.HandlerName = query.Name
		query.ServiceName = query.Name
	} else if strings.HasPrefix(query.Name, "Delete") {
		query.Type = ":exec"
		query.HTTPMethod = "DELETE"
		query.HandlerName = query.Name
		query.ServiceName = query.Name
	}

	// Generate URL path - IMPROVED
	query.URLPath = g.generateURLPath(query.Name)

	// Parse parameters
	if fn.Type.Params != nil {
		for _, param := range fn.Type.Params.List {
			if len(param.Names) > 0 {
				paramName := param.Names[0].Name
				if paramName == "ctx" {
					continue
				}

				paramType := ""
				if param.Type != nil {
					paramType = g.typeToString(param.Type)
				}

				query.Params = append(query.Params, ParamInfo{
					Name: paramName,
					Type: paramType,
				})
			}
		}
	}

	// Parse return type
	if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
		if len(fn.Type.Results.List) >= 1 {
			returnType := g.typeToString(fn.Type.Results.List[0].Type)
			query.ReturnType = returnType
		}
	}

	return query
}

func (g *Generator) typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.ArrayType:
		return "[]" + g.typeToString(t.Elt)
	case *ast.StarExpr:
		return "*" + g.typeToString(t.X)
	case *ast.SelectorExpr:
		return g.typeToString(t.X) + "." + t.Sel.Name
	default:
		return "interface{}"
	}
}

// IMPROVED URL generation
func (g *Generator) generateURLPath(queryName string) string {
	if strings.HasPrefix(queryName, "Get") || strings.HasPrefix(queryName, "List") {
		if strings.Contains(queryName, "ByID") || strings.Contains(queryName, "ById") {
			return "/{id}" // Clean: GET /api/post/{id}
		}
		if strings.Contains(strings.ToLower(queryName), "public") {
			return "/" // Clean: GET /api/post/ for GetPublicPosts
		}
		// For other Get queries, use the name without "Get"
		basePath := strings.TrimPrefix(camelToKebab(queryName), "get-")
		basePath = strings.TrimPrefix(basePath, g.Feature+"-")
		if basePath == "" {
			return "/"
		}
		return "/" + basePath
	} else if strings.HasPrefix(queryName, "Create") {
		return "/" // Clean: POST /api/post/ for CreatePost
	} else if strings.HasPrefix(queryName, "Update") {
		return "/{id}" // Clean: PUT /api/post/{id}
	} else if strings.HasPrefix(queryName, "Delete") {
		return "/{id}" // Clean: DELETE /api/post/{id}
	}

	return "/" + camelToKebab(queryName)
}

func (g *Generator) isRelevantQuery(queryName, filename string) bool {
	// Check if the file matches the feature
	basename := filepath.Base(filename)
	filenameBase := strings.TrimSuffix(basename, ".sql.go")

	// More flexible matching - if generating for "post" and file is "post.sql.go", include all queries
	return strings.EqualFold(filenameBase, g.Feature) ||
		strings.Contains(strings.ToLower(queryName), strings.ToLower(g.Feature)) ||
		g.Feature == "civil" // Special case for civil (includes all)
}

func (g *Generator) needsHealthCheck(queries []QueryInfo) bool {
	for _, q := range queries {
		if strings.ToLower(q.Name) == "healthcheck" || strings.ToLower(q.Name) == "health" {
			return false
		}
	}
	return true
}

// IMPROVED Template generation functions
// ... keep all your existing code until generateHandlers function ...

func (g *Generator) generateHandlers(data APIGenerationData) error {
	// Analyze what imports are actually needed
	needsStrconv := false
	needsUUID := false

	for _, q := range data.Queries {
		if strings.Contains(q.URLPath, "{id}") {
			// Check if any query uses non-UUID parameters (would need strconv)
			hasNonUUID := false
			for _, param := range q.Params {
				if !strings.Contains(param.Type, "uuid") && param.Name != "ctx" {
					hasNonUUID = true
					break
				}
			}
			if hasNonUUID {
				needsStrconv = true
			} else {
				needsUUID = true
			}
		}
	}

	tmpl := `// Code generated by genapi. DO NOT EDIT manually.
package {{.Package}}

import (
	"encoding/json"
	"net/http"
	{{if .NeedsStrconv}}"strconv"{{end}}
	
	"github.com/go-chi/chi/v5"
	{{if .NeedsUUID}}"github.com/google/uuid"{{end}}
	"go.uber.org/zap"
)

type Handlers struct {
	service *Service
	logger  *zap.SugaredLogger
}

func NewHandlers(service *Service, logger *zap.SugaredLogger) *Handlers {
	return &Handlers{
		service: service,
		logger:  logger,
	}
}

{{range .Queries}}
{{if eq .HTTPMethod "POST"}}
// {{.HandlerName}} creates a new {{$.Feature}}
// @Summary Create {{$.Feature}}
// @Description Create a new {{$.Feature}} record
// @Tags {{$.Feature}}
// @Accept json
// @Produce json
// @Param request body {{.HandlerName}}Request true "{{$.Feature}} data"
// @Success 201 {object} map[string]interface{} "Created {{$.Feature}}"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/{{$.Feature}}{{.URLPath}} [{{.HTTPMethod | lower}}]
func (h *Handlers) {{.HandlerName}}(w http.ResponseWriter, r *http.Request) {
{{else if contains .URLPath "{id}"}}
// {{.HandlerName}} retrieves a {{$.Feature}} by ID
// @Summary Get {{$.Feature}} by ID
// @Description Get a specific {{$.Feature}} by its ID
// @Tags {{$.Feature}}
// @Accept json
// @Produce json
// @Param id path string true "{{$.Feature}} ID"
// @Success 200 {object} map[string]interface{} "{{$.Feature}} found"
// @Failure 400 {object} map[string]interface{} "Invalid ID"
// @Failure 404 {object} map[string]interface{} "{{$.Feature}} not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/{{$.Feature}}{{.URLPath}} [{{.HTTPMethod | lower}}]
func (h *Handlers) {{.HandlerName}}(w http.ResponseWriter, r *http.Request) {
{{else}}
// {{.HandlerName}} retrieves all {{$.Feature}}s
// @Summary Get all {{$.Feature}}s
// @Description Retrieve all {{$.Feature}} records
// @Tags {{$.Feature}}
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "List of {{$.Feature}}s"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/{{$.Feature}}{{.URLPath}} [{{.HTTPMethod | lower}}]
func (h *Handlers) {{.HandlerName}}(w http.ResponseWriter, r *http.Request) {
{{end}}
	{{if contains .URLPath "{id}"}}
	idParam := chi.URLParam(r, "id")
	{{if contains (printf "%v" .Params) "uuid"}}
	id, err := uuid.Parse(idParam)
	if err != nil {
		h.logger.Errorf("Invalid UUID: %v", err)
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}
	{{else}}
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		h.logger.Errorf("Invalid ID: %v", err)
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}
	{{end}}
	{{end}}

	{{if eq .HTTPMethod "POST"}}
	var req {{.HandlerName}}Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Errorf("Failed to decode request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.service.{{.ServiceName}}(r.Context(), req.Title, req.Body)
	{{else if contains .URLPath "{id}"}}
	result, err := h.service.{{.ServiceName}}(r.Context(), id)
	{{else}}
	result, err := h.service.{{.ServiceName}}(r.Context())
	{{end}}
	if err != nil {
		h.logger.Errorf("Service error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	{{if eq .HTTPMethod "POST"}}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "{{$.Feature}} created successfully",
		"data": result,
	})
	{{else if eq .Type ":many"}}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  result,
		"count": len(result),
	})
	{{else}}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": result,
	})
	{{end}}
}

{{if eq .HTTPMethod "POST"}}
type {{.HandlerName}}Request struct {
	Title string ` + "`json:\"title\" example:\"My Post Title\"`" + `
	Body  string ` + "`json:\"body\" example:\"This is the post content\"`" + `
}
{{end}}
{{end}}

{{if .HasHealthCheck}}
// HealthCheck checks the health of the {{.Feature}} service
// @Summary Health check
// @Description Check if the {{.Feature}} service is healthy
// @Tags {{.Feature}}
// @Produce json
// @Success 200 {object} map[string]interface{} "Service is healthy"
// @Failure 503 {object} map[string]interface{} "Service is unhealthy"
// @Router /api/{{.Feature}}/health [get]
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	err := h.service.HealthCheck(r.Context())
	if err != nil {
		http.Error(w, "Service unhealthy", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"service": "{{.Feature}}-api",
		"version": "1.0.0",
	})
}
{{end}}
`

	// Create extended template data
	templateData := struct {
		APIGenerationData
		NeedsStrconv bool
		NeedsUUID    bool
	}{
		APIGenerationData: data,
		NeedsStrconv:      needsStrconv,
		NeedsUUID:         needsUUID,
	}

	return g.writeTemplateWithData("handlers.go", tmpl, templateData)
}

// Add this new function to handle the extended template data
func (g *Generator) writeTemplateWithData(filename, tmpl string, data interface{}) error {
	// Add custom template functions
	funcMap := template.FuncMap{
		"title":      strings.Title,
		"lower":      strings.ToLower,
		"snakeCase":  toSnakeCase,
		"contains":   strings.Contains,
		"methodName": methodName,
	}

	t, err := template.New(filename).Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	// Use filepath.Join for Windows compatibility
	file, err := os.Create(filepath.Join("internal", "generated", "api", g.Feature, filename))
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, data)
}

// Update getRequiredImports to not include unused imports
func (g *Generator) getRequiredImports(queries []QueryInfo) []string {
	// Don't include imports here since we handle them in templates
	return []string{}
}
func (g *Generator) generateService(data APIGenerationData) error {
	tmpl := `// Code generated by genapi. DO NOT EDIT manually.
package {{.Package}}

import (
	"context"
	"fmt"

	"github.com/eif-courses/civilregistry/internal/generated/repository"  // ✅ Correct path
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Service struct {
	repo   *repository.Queries
	logger *zap.SugaredLogger
}

func NewService(repo *repository.Queries, logger *zap.SugaredLogger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

{{range .Queries}}
{{if eq .HTTPMethod "POST"}}
func (s *Service) {{.ServiceName}}(ctx context.Context, title, body string) (*repository.Post, error) {
	s.logger.Infof("Creating new {{.ServiceName}}: %s", title)

	if title == "" || body == "" {
		return nil, fmt.Errorf("title and body are required")
	}

	result, err := s.repo.{{.Name}}(ctx, repository.CreatePostParams{
		Title: title,
		Body:  body,
	})
	if err != nil {
		s.logger.Errorf("Failed {{.ServiceName}}: %v", err)
		return nil, fmt.Errorf("failed {{.ServiceName}}: %w", err)
	}

	s.logger.Infof("{{.ServiceName}} completed successfully with ID: %s", result.ID)
	return &result, nil
}
{{else if contains .URLPath "{id}"}}
func (s *Service) {{.ServiceName}}(ctx context.Context, id uuid.UUID) (*repository.Post, error) {
	s.logger.Infof("{{.ServiceName}} called for ID: %s", id)

	result, err := s.repo.{{.Name}}(ctx, id)
	if err != nil {
		s.logger.Errorf("Failed {{.ServiceName}}: %v", err)
		return nil, fmt.Errorf("failed {{.ServiceName}}: %w", err)
	}

	s.logger.Info("{{.ServiceName}} completed successfully")
	return &result, nil
}
{{else}}
func (s *Service) {{.ServiceName}}(ctx context.Context) ([]repository.Post, error) {
	s.logger.Info("{{.ServiceName}} called")

	result, err := s.repo.{{.Name}}(ctx)
	if err != nil {
		s.logger.Errorf("Failed {{.ServiceName}}: %v", err)
		return nil, fmt.Errorf("failed {{.ServiceName}}: %w", err)
	}

	s.logger.Infof("{{.ServiceName}} returned %d items", len(result))
	return result, nil
}
{{end}}
{{end}}

{{if .HasHealthCheck}}
func (s *Service) HealthCheck(ctx context.Context) error {
	s.logger.Info("Performing health check")
	return nil
}
{{end}}
`

	return g.writeTemplate("service.go", tmpl, data)
}
func (g *Generator) generateRouter(data APIGenerationData) error {
	tmpl := `// Code generated by genapi. DO NOT EDIT manually.
package {{.Package}}

import (
	"github.com/eif-courses/civilregistry/internal/generated/repository"  // ✅ Correct path
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func {{.Feature | title}}Router(queries *repository.Queries, log *zap.SugaredLogger) chi.Router {
	r := chi.NewRouter()

	// Create service with repository
	service := NewService(queries, log)
	handlers := NewHandlers(service, log)

	{{if .HasHealthCheck}}r.Get("/health", handlers.HealthCheck){{end}}
	{{range .Queries}}r.{{.HTTPMethod | methodName}}("{{.URLPath}}", handlers.{{.HandlerName}})
	{{end}}

	return r
}
`

	return g.writeTemplate("router.go", tmpl, data)
}
func (g *Generator) writeTemplate(filename, tmpl string, data interface{}) error {
	// Add custom template functions
	funcMap := template.FuncMap{
		"title":      strings.Title,
		"lower":      strings.ToLower,
		"snakeCase":  toSnakeCase,
		"contains":   strings.Contains,
		"methodName": methodName,
	}

	t, err := template.New(filename).Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	// Use filepath.Join for Windows compatibility - UPDATED PATH
	file, err := os.Create(filepath.Join("internal", "generated", "api", g.Feature, filename))
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, data)
}

// Utility functions
func camelToKebab(s string) string {
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	kebab := re.ReplaceAllString(s, "${1}-${2}")
	return strings.ToLower(kebab)
}

func toSnakeCase(s string) string {
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := re.ReplaceAllString(s, "${1}_${2}")
	return strings.ToLower(snake)
}

func methodName(httpMethod string) string {
	switch strings.ToUpper(httpMethod) {
	case "GET":
		return "Get"
	case "POST":
		return "Post"
	case "PUT":
		return "Put"
	case "DELETE":
		return "Delete"
	case "PATCH":
		return "Patch"
	default:
		return "Get"
	}
}
