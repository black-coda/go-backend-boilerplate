# ===== Go =====
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with `go test -c`
*.test

# Output of the go coverage tool
*.out
coverage.out
coverage.html

# Go workspace file
go.work
go.work.sum

# Dependency directories (if not using modules)
vendor/

# Go build cache / compiled files
/bin/
/dist/
build/

# Debug files
__debug_bin*
*.log

# GoLand / IDE
.idea/
.vscode/
*.swp
*.swo

# ===== Bun =====
# Dependencies
node_modules/
.pnp
.pnp.js

# Bun lockfile is usually committed, but ignore if you don't want it
# bun.lockb

# Build output
dist/
build/
out/
.next/
.nuxt/

# Cache
.bun/
.turbo/
.cache/

# Env files
.env
.env.local
.env.*.local

# Logs
logs/
*.log
npm-debug.log*
yarn-debug.log*
yarn-error.log*
bun-debug.log*

# Coverage
coverage/
.nyc_output/

# ===== OS =====
.DS_Store
Thumbs.db
desktop.ini

# ===== Editor =====
.idea/
.vscode/
*.swp
*.swo

### What middleware does
Middleware executes before or after a request handler.

Common middleware responsibilities include:
- Authentication
- Authorization
- Logging
- Request tracing
- Rate limiting
- Request validation
- Panic recovery
- Error handling



When errors are wrapped using fmt.Errorf("...: %w", err), simple type assertions (like err.(*MyError)) fail because the custom error is buried inside wrapper layers. errors.As() traverses the entire chain of wrapped errors. If it finds an error matching the type of target, it copies that error's value into target and returns true