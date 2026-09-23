package scanner

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var packageNameRE = regexp.MustCompile(`([a-zA-Z0-9._-]+\.(js|ts|mjs|cjs|py|rb|go))`)

// Detection rules stay in priority order when several names match.
type frameworkRule struct {
	name    string
	markers []string
}

var imageFrameworks = []frameworkRule{
	{"PostgreSQL", []string{"postgres"}},
	{"Redis", []string{"redis"}},
	{"MySQL", []string{"mysql", "mariadb"}},
	{"MongoDB", []string{"mongo"}},
	{"nginx", []string{"nginx"}},
	{"LocalStack", []string{"localstack"}},
	{"RabbitMQ", []string{"rabbitmq"}},
	{"Kafka", []string{"kafka"}},
	{"Elasticsearch", []string{"elasticsearch", "opensearch"}},
	{"MinIO", []string{"minio"}},
}

var commandFrameworks = []frameworkRule{
	{"Next.js", []string{"next"}},
	{"Vite", []string{"vite"}},
	{"Nuxt", []string{"nuxt"}},
	{"Angular", []string{"angular", "ng serve"}},
	{"Webpack", []string{"webpack"}},
	{"Remix", []string{"remix"}},
	{"Astro", []string{"astro"}},
	{"Gatsby", []string{"gatsby"}},
	{"Flask", []string{"flask"}},
	{"Django", []string{"django", "manage.py"}},
	{"FastAPI", []string{"uvicorn"}},
	{"Rails", []string{"rails"}},
	{"Rust", []string{"cargo", "rustc"}},
}

var packageFrameworks = []frameworkRule{
	{"Next.js", []string{"next"}},
	{"Nuxt", []string{"nuxt", "nuxt3"}},
	{"SvelteKit", []string{"@sveltejs/kit"}},
	{"Svelte", []string{"svelte"}},
	{"Remix", []string{"@remix-run/react", "remix"}},
	{"Astro", []string{"astro"}},
	{"Vite", []string{"vite"}},
	{"Angular", []string{"@angular/core"}},
	{"Vue", []string{"vue"}},
	{"React", []string{"react"}},
	{"Express", []string{"express"}},
	{"Fastify", []string{"fastify"}},
	{"Hono", []string{"hono"}},
	{"Koa", []string{"koa"}},
	{"NestJS", []string{"nestjs", "@nestjs/core"}},
	{"Gatsby", []string{"gatsby"}},
	{"Webpack", []string{"webpack-dev-server"}},
	{"esbuild", []string{"esbuild"}},
	{"Parcel", []string{"parcel"}},
}

func IsDevPort(port PortInfo) bool {
	if IsDevProcess(port.ProcessName, port.Command) {
		return true
	}
	return isUVXCommand(port.ParentCommand)
}

var systemApps = []string{
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

var devNames = map[string]struct{}{
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

var devCommandIndicators = []*regexp.Regexp{
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

func IsDevProcess(processName, command string) bool {
	name := strings.ToLower(strings.TrimSpace(processName))
	if hasAnyPrefix(name, systemApps) {
		return false
	}
	if _, ok := devNames[name]; ok {
		return true
	}
	if hasAnyPrefix(name, []string{"python", "com.docke", "docker"}) {
		return true
	}
	return hasDevCommandIndicator(strings.ToLower(command))
}

func hasAnyPrefix(name string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func hasDevCommandIndicator(command string) bool {
	for _, re := range devCommandIndicators {
		if re.MatchString(command) {
			return true
		}
	}
	return false
}

func isUVXCommand(command string) bool {
	_, ok := uvxPackageName(command)
	return ok
}

func uvxProjectLabel(command string) (string, bool) {
	packageName, ok := uvxPackageName(command)
	if !ok {
		return "", false
	}
	return "uvx " + packageName, true
}

func uvxPackageName(command string) (string, bool) {
	fields := strings.Fields(command)
	for i, field := range fields {
		name := filepath.Base(strings.ToLower(field))
		if name == "uvx" {
			return uvxPackageNameFromArgs(fields[i+1:])
		}
		if name == "uv" && isUVToolUVX(fields[i+1:]) {
			return uvxPackageNameFromArgs(fields[i+3:])
		}
	}
	return "", false
}

func isUVToolUVX(args []string) bool {
	return len(args) >= 2 && strings.ToLower(args[0]) == "tool" && strings.ToLower(args[1]) == "uvx"
}

func uvxPackageNameFromArgs(args []string) (string, bool) {
	for i := 0; i < len(args); i++ {
		arg := strings.TrimSpace(args[i])
		switch {
		case arg == "":
			continue
		case arg == "--" || arg == "--from":
			return uvxNextPackage(args, i)
		case strings.HasPrefix(arg, "--from="):
			return cleanUVXPackageName(strings.TrimPrefix(arg, "--from="))
		case strings.HasPrefix(arg, "-"):
			if uvxOptionTakesValue(arg) {
				i++
			}
		default:
			return cleanUVXPackageName(arg)
		}
	}
	return "", false
}

func uvxNextPackage(args []string, index int) (string, bool) {
	if index+1 >= len(args) {
		return "", false
	}
	return cleanUVXPackageName(args[index+1])
}

func uvxOptionTakesValue(option string) bool {
	option, _, _ = strings.Cut(option, "=")
	valueOptions := map[string]struct{}{
		"--active-groups":           {},
		"--config-setting":          {},
		"--config-settings-package": {},
		"--config-file":             {},
		"--constraint":              {},
		"--default-index":           {},
		"--dependency-mode":         {},
		"--directory":               {},
		"--env-file":                {},
		"--exclude-newer":           {},
		"--exclude-newer-package":   {},
		"--extra-index-url":         {},
		"--find-links":              {},
		"--fork-strategy":           {},
		"--group":                   {},
		"--index":                   {},
		"--index-strategy":          {},
		"--index-url":               {},
		"--keyring-provider":        {},
		"--link-mode":               {},
		"--override":                {},
		"--prerelease":              {},
		"--project":                 {},
		"--python":                  {},
		"--python-platform":         {},
		"--python-preference":       {},
		"--refresh-package":         {},
		"--reinstall-package":       {},
		"--resolution":              {},
		"--upgrade-package":         {},
		"--with":                    {},
		"--with-editable":           {},
		"--with-requirements":       {},
	}
	_, ok := valueOptions[option]
	return ok
}

func cleanUVXPackageName(value string) (string, bool) {
	value = strings.Trim(strings.TrimSpace(value), "'\"")
	if value == "" {
		return "", false
	}
	if strings.Contains(value, "/") {
		value = filepath.Base(value)
	}

	cutAt := len(value)
	for _, separator := range []string{"[", "==", ">=", "<=", "!=", "~=", ">", "<", ";"} {
		if index := strings.Index(value, separator); index >= 0 && index < cutAt {
			cutAt = index
		}
	}
	value = value[:cutAt]
	return value, value != ""
}

func DetectFrameworkFromImage(image string) string {
	if name := frameworkContaining(strings.ToLower(image), imageFrameworks); name != "" {
		return name
	}
	return "Docker"
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

// projectLabel names a project by its parent and own directory, such as "ports/frontend".
// The parent makes common names like "frontend" distinct in the port list.
func projectLabel(projectRoot string) string {
	projectRoot = filepath.Clean(projectRoot)
	name := filepath.Base(projectRoot)
	parent := filepath.Base(filepath.Dir(projectRoot))
	if projectRoot == string(filepath.Separator) || isRootParent(parent) {
		return name
	}
	return filepath.Join(parent, name)
}

func isRootParent(parent string) bool {
	return parent == string(filepath.Separator) || parent == "."
}

func DetectFramework(projectRoot string) string {
	if framework := frameworkFromPackageJSON(filepath.Join(projectRoot, "package.json")); framework != "" {
		return framework
	}
	return frameworkFromProjectFiles(projectRoot)
}

func frameworkFromPackageJSON(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return ""
	}
	deps := map[string]string{}
	maps.Copy(deps, pkg.Dependencies)
	maps.Copy(deps, pkg.DevDependencies)

	for _, framework := range packageFrameworks {
		for _, dep := range framework.markers {
			if hasDep(deps, dep) {
				return framework.name
			}
		}
	}
	return ""
}

var projectFileFrameworks = []frameworkRule{
	{"Vite", []string{"vite.config.ts", "vite.config.js"}},
	{"Next.js", []string{"next.config.js", "next.config.mjs"}},
	{"Angular", []string{"angular.json"}},
	{"Rust", []string{"Cargo.toml"}},
	{"Go", []string{"go.mod"}},
	{"Django", []string{"manage.py"}},
	{"Ruby", []string{"Gemfile"}},
}

func frameworkFromProjectFiles(projectRoot string) string {
	for _, framework := range projectFileFrameworks {
		for _, filename := range framework.markers {
			if fileExists(filepath.Join(projectRoot, filename)) {
				return framework.name
			}
		}
	}
	return ""
}

func DetectFrameworkFromCommand(command, processName string) string {
	if name := frameworkContaining(strings.ToLower(command), commandFrameworks); name != "" {
		return name
	}
	return DetectFrameworkFromName(processName)
}

func frameworkContaining(text string, rules []frameworkRule) string {
	for _, rule := range rules {
		for _, marker := range rule.markers {
			if strings.Contains(text, marker) {
				return rule.name
			}
		}
	}
	return ""
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
		meaningful = append(meaningful, commandPartLabel(part))
		if len(meaningful) >= 3 {
			break
		}
	}
	if len(meaningful) > 0 {
		return strings.Join(meaningful, " ")
	}
	return processName
}

func commandPartLabel(part string) string {
	if strings.Contains(part, "/") {
		return filepath.Base(part)
	}
	if match := packageNameRE.FindStringSubmatch(part); len(match) > 1 {
		return match[1]
	}
	return part
}

func hasDep(deps map[string]string, name string) bool {
	_, ok := deps[name]
	return ok
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
