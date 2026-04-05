package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var packageNameRE = regexp.MustCompile(`([a-zA-Z0-9._-]+\.(js|ts|mjs|cjs|py|rb|go))`)

func IsDevProcess(processName, command string) bool {
	name := strings.ToLower(strings.TrimSpace(processName))
	cmd := strings.ToLower(command)

	systemApps := []string{
		"spotify",
		"raycast",
		"tableplus",
		"postman",
		"linear",
		"cursor",
		"controlce",
		"rapportd",
		"slack",
		"discord",
		"firefox",
		"chrome",
		"google",
		"safari",
		"figma",
		"notion",
		"zoom",
		"teams",
		"code",
		"iterm2",
		"warp",
		"arc",
		"loginwindow",
		"windowserver",
		"systemuiserver",
		"kernel_task",
		"launchd",
		"mdworker",
		"mds_stores",
		"cfprefsd",
		"coreaudio",
		"airportd",
		"bluetoothd",
		"sharingd",
		"usernoted",
		"notificationcenter",
		"cloudd",
	}
	for _, app := range systemApps {
		if strings.HasPrefix(name, app) {
			return false
		}
	}

	devNames := map[string]struct{}{
		"node":           {},
		"python":         {},
		"python3":        {},
		"ruby":           {},
		"java":           {},
		"go":             {},
		"cargo":          {},
		"deno":           {},
		"bun":            {},
		"php":            {},
		"uvicorn":        {},
		"gunicorn":       {},
		"flask":          {},
		"rails":          {},
		"npm":            {},
		"npx":            {},
		"yarn":           {},
		"pnpm":           {},
		"tsc":            {},
		"tsx":            {},
		"esbuild":        {},
		"rollup":         {},
		"turbo":          {},
		"nx":             {},
		"jest":           {},
		"vitest":         {},
		"mocha":          {},
		"pytest":         {},
		"cypress":        {},
		"playwright":     {},
		"rustc":          {},
		"dotnet":         {},
		"gradle":         {},
		"mvn":            {},
		"mix":            {},
		"elixir":         {},
		"docker":         {},
		"docker-sandbox": {},
	}
	if _, ok := devNames[name]; ok {
		return true
	}

	if strings.HasPrefix(name, "python") || strings.HasPrefix(name, "com.docke") || strings.HasPrefix(name, "docker") {
		return true
	}

	indicators := []*regexp.Regexp{
		regexp.MustCompile(`\bnode\b`),
		regexp.MustCompile(`\bnext([\s-]|$)`),
		regexp.MustCompile(`\bvite\b`),
		regexp.MustCompile(`\bnuxt\b`),
		regexp.MustCompile(`\bwebpack\b`),
		regexp.MustCompile(`\bremix\b`),
		regexp.MustCompile(`\bastro\b`),
		regexp.MustCompile(`\bgulp\b`),
		regexp.MustCompile(`\bng serve\b`),
		regexp.MustCompile(`\bgatsby\b`),
		regexp.MustCompile(`\bflask\b`),
		regexp.MustCompile(`\bdjango\b|manage\.py`),
		regexp.MustCompile(`\buvicorn\b`),
		regexp.MustCompile(`\brails\b`),
		regexp.MustCompile(`\bcargo\b`),
		regexp.MustCompile(`\bgo run\b`),
	}
	for _, re := range indicators {
		if re.MatchString(cmd) {
			return true
		}
	}

	return false
}

func DetectFrameworkFromImage(image string) string {
	img := strings.ToLower(image)
	switch {
	case img == "":
		return "Docker"
	case strings.Contains(img, "postgres"):
		return "PostgreSQL"
	case strings.Contains(img, "redis"):
		return "Redis"
	case strings.Contains(img, "mysql"), strings.Contains(img, "mariadb"):
		return "MySQL"
	case strings.Contains(img, "mongo"):
		return "MongoDB"
	case strings.Contains(img, "nginx"):
		return "nginx"
	case strings.Contains(img, "localstack"):
		return "LocalStack"
	case strings.Contains(img, "rabbitmq"):
		return "RabbitMQ"
	case strings.Contains(img, "kafka"):
		return "Kafka"
	case strings.Contains(img, "elasticsearch"), strings.Contains(img, "opensearch"):
		return "Elasticsearch"
	case strings.Contains(img, "minio"):
		return "MinIO"
	default:
		return "Docker"
	}
}

func FindProjectRoot(dir string) string {
	markers := []string{
		"package.json",
		"Cargo.toml",
		"go.mod",
		"pyproject.toml",
		"Gemfile",
		"pom.xml",
		"build.gradle",
	}

	current := dir
	for depth := 0; current != "/" && depth < 15; depth++ {
		for _, marker := range markers {
			if _, err := os.Stat(filepath.Join(current, marker)); err == nil {
				return current
			}
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return dir
}

func DetectFramework(projectRoot string) string {
	pkgPath := filepath.Join(projectRoot, "package.json")
	if data, err := os.ReadFile(pkgPath); err == nil {
		var pkg struct {
			Dependencies    map[string]string `json:"dependencies"`
			DevDependencies map[string]string `json:"devDependencies"`
		}
		if err := json.Unmarshal(data, &pkg); err == nil {
			deps := map[string]string{}
			for k, v := range pkg.Dependencies {
				deps[k] = v
			}
			for k, v := range pkg.DevDependencies {
				deps[k] = v
			}

			switch {
			case hasDep(deps, "next"):
				return "Next.js"
			case hasDep(deps, "nuxt") || hasDep(deps, "nuxt3"):
				return "Nuxt"
			case hasDep(deps, "@sveltejs/kit"):
				return "SvelteKit"
			case hasDep(deps, "svelte"):
				return "Svelte"
			case hasDep(deps, "@remix-run/react") || hasDep(deps, "remix"):
				return "Remix"
			case hasDep(deps, "astro"):
				return "Astro"
			case hasDep(deps, "vite"):
				return "Vite"
			case hasDep(deps, "@angular/core"):
				return "Angular"
			case hasDep(deps, "vue"):
				return "Vue"
			case hasDep(deps, "react"):
				return "React"
			case hasDep(deps, "express"):
				return "Express"
			case hasDep(deps, "fastify"):
				return "Fastify"
			case hasDep(deps, "hono"):
				return "Hono"
			case hasDep(deps, "koa"):
				return "Koa"
			case hasDep(deps, "nestjs") || hasDep(deps, "@nestjs/core"):
				return "NestJS"
			case hasDep(deps, "gatsby"):
				return "Gatsby"
			case hasDep(deps, "webpack-dev-server"):
				return "Webpack"
			case hasDep(deps, "esbuild"):
				return "esbuild"
			case hasDep(deps, "parcel"):
				return "Parcel"
			}
		}
	}

	switch {
	case fileExists(filepath.Join(projectRoot, "vite.config.ts")) || fileExists(filepath.Join(projectRoot, "vite.config.js")):
		return "Vite"
	case fileExists(filepath.Join(projectRoot, "next.config.js")) || fileExists(filepath.Join(projectRoot, "next.config.mjs")):
		return "Next.js"
	case fileExists(filepath.Join(projectRoot, "angular.json")):
		return "Angular"
	case fileExists(filepath.Join(projectRoot, "Cargo.toml")):
		return "Rust"
	case fileExists(filepath.Join(projectRoot, "go.mod")):
		return "Go"
	case fileExists(filepath.Join(projectRoot, "manage.py")):
		return "Django"
	case fileExists(filepath.Join(projectRoot, "Gemfile")):
		return "Ruby"
	default:
		return ""
	}
}

func DetectFrameworkFromCommand(command, processName string) string {
	cmd := strings.ToLower(command)
	switch {
	case strings.Contains(cmd, "next"):
		return "Next.js"
	case strings.Contains(cmd, "vite"):
		return "Vite"
	case strings.Contains(cmd, "nuxt"):
		return "Nuxt"
	case strings.Contains(cmd, "angular") || strings.Contains(cmd, "ng serve"):
		return "Angular"
	case strings.Contains(cmd, "webpack"):
		return "Webpack"
	case strings.Contains(cmd, "remix"):
		return "Remix"
	case strings.Contains(cmd, "astro"):
		return "Astro"
	case strings.Contains(cmd, "gatsby"):
		return "Gatsby"
	case strings.Contains(cmd, "flask"):
		return "Flask"
	case strings.Contains(cmd, "django") || strings.Contains(cmd, "manage.py"):
		return "Django"
	case strings.Contains(cmd, "uvicorn"):
		return "FastAPI"
	case strings.Contains(cmd, "rails"):
		return "Rails"
	case strings.Contains(cmd, "cargo"), strings.Contains(cmd, "rustc"):
		return "Rust"
	default:
		return DetectFrameworkFromName(processName)
	}
}

func DetectFrameworkFromName(processName string) string {
	switch strings.ToLower(processName) {
	case "node":
		return "Node.js"
	case "python", "python3":
		return "Python"
	case "ruby":
		return "Ruby"
	case "java":
		return "Java"
	case "go":
		return "Go"
	default:
		return ""
	}
}

func SummarizeCommand(command, processName string) string {
	parts := strings.Fields(command)
	meaningful := make([]string, 0, 3)
	for i, part := range parts {
		if i == 0 || strings.HasPrefix(part, "-") {
			continue
		}
		if strings.Contains(part, "/") {
			meaningful = append(meaningful, filepath.Base(part))
		} else if match := packageNameRE.FindStringSubmatch(part); len(match) > 1 {
			meaningful = append(meaningful, match[1])
		} else {
			meaningful = append(meaningful, part)
		}
		if len(meaningful) >= 3 {
			break
		}
	}
	if len(meaningful) > 0 {
		return strings.Join(meaningful, " ")
	}
	return processName
}

func hasDep(deps map[string]string, name string) bool {
	_, ok := deps[name]
	return ok
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
