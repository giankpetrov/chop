package filters

import "sort"

// routerFunc resolves a FilterFunc for a command given its arguments.
type routerFunc func(args []string) FilterFunc

// entry describes a registered built-in filter.
type entry struct {
	filter   FilterFunc // for simple/direct filters (e.g., filterPing)
	router   routerFunc // for subcommand routers (e.g., getGitFilter)
	builtins []string   // subcommands shown in `chop list`; nil = command-level
	hidden   bool       // if true, excluded from `chop list` (autoDetect fallbacks)
}

// registry maps command names to their filter entries.
var registry = map[string]*entry{}

// aliases tracks command names registered via registerAlias.
// These are excluded from ListBuiltins.
var aliases = map[string]bool{}

// registerFilter registers a direct FilterFunc for one or more command names.
func registerFilter(filter FilterFunc, builtins []string, commands ...string) {
	e := &entry{filter: filter, builtins: builtins}
	for _, cmd := range commands {
		registry[cmd] = e
	}
}

// registerRouter registers a sub-router for one or more command names.
func registerRouter(router routerFunc, builtins []string, commands ...string) {
	e := &entry{router: router, builtins: builtins}
	for _, cmd := range commands {
		registry[cmd] = e
	}
}

// registerAlias makes additional command names share an already-registered entry.
// Aliases are excluded from ListBuiltins.
func registerAlias(target string, names ...string) {
	e := registry[target]
	for _, name := range names {
		registry[name] = e
		aliases[name] = true
	}
}

// registerHidden registers a filter that won't appear in `chop list`.
func registerHidden(filter FilterFunc, commands ...string) {
	e := &entry{filter: filter, hidden: true}
	for _, cmd := range commands {
		registry[cmd] = e
	}
}

func init() {
	// --- Routers (commands with subcommand dispatch) ---
	registerRouter(getGitFilter, []string{"status", "log", "diff", "show", "branch", "push", "pull", "fetch"}, "git")
	registerRouter(getNpmFilter, []string{"install", "list", "test"}, "npm")
	registerRouter(getNpxFilter, nil, "npx")
	registerRouter(getPnpmFilter, []string{"install", "list"}, "pnpm")
	registerRouter(getYarnFilter, []string{"install"}, "yarn")
	registerRouter(getBunFilter, []string{"install"}, "bun")
	registerRouter(getDockerFilter, []string{"build", "ps", "logs"}, "docker")
	registerAlias("docker", "podman")
	registerRouter(getDockerComposeFilter, nil, "docker-compose")
	registerRouter(getDotnetFilter, []string{"build", "test"}, "dotnet")
	registerRouter(getKubectlFilter, []string{"get", "describe", "logs"}, "kubectl")
	registerRouter(getHelmFilter, []string{"install", "list"}, "helm")
	registerRouter(getTerraformFilter, []string{"plan", "apply", "init"}, "terraform")
	registerAlias("terraform", "tofu")
	registerRouter(getCargoFilter, []string{"build", "test"}, "cargo")
	registerRouter(getGoFilter, []string{"build", "test"}, "go")
	registerRouter(getGhFilter, nil, "gh")
	registerRouter(getAwsFilter, nil, "aws")
	registerRouter(getAzFilter, nil, "az")
	registerRouter(getGcloudFilter, nil, "gcloud")
	registerRouter(getMavenFilter, nil, "mvn")
	registerRouter(getAngularFilter, nil, "ng")
	registerRouter(getNxFilter, nil, "nx")
	registerRouter(getUvFilter, nil, "uv")
	registerRouter(getComposerFilter, []string{"install"}, "composer")
	registerRouter(getAcliFilter, nil, "acli")

	// Routers with aliases (primary name visible, alternatives hidden)
	registerRouter(getGradleFilter, nil, "gradle")
	registerAlias("gradle", "gradlew")

	registerRouter(getPipFilter, []string{"install"}, "pip")
	registerAlias("pip", "pip3")

	registerRouter(getBundleFilter, []string{"install"}, "bundle")
	registerAlias("bundle", "bundler")

	// --- Simple filters ---
	registerFilter(filterTsc, nil, "tsc")
	registerFilter(filterEslint, nil, "eslint")
	registerFilter(filterEslint, nil, "biome")
	registerFilter(filterGrep, nil, "grep")
	registerFilter(filterGrep, nil, "rg")
	registerFilter(filterCurl, nil, "curl")
	registerFilter(filterHttpie, nil, "http")
	registerFilter(filterPytest, nil, "pytest")
	registerFilter(filterMypy, nil, "mypy")
	registerFilter(filterRuff, nil, "ruff")
	registerFilter(filterRuff, nil, "flake8")
	registerFilter(filterPylint, nil, "pylint")
	registerFilter(filterRspec, nil, "rspec")
	registerFilter(filterRubocop, nil, "rubocop")
	registerFilter(filterAnsiblePlaybook, nil, "ansible-playbook")
	registerFilter(filterMake, nil, "make")
	registerFilter(filterCmake, nil, "cmake")
	registerFilter(filterPing, nil, "ping")
	registerFilter(filterPsCmd, nil, "ps")
	registerFilter(filterNetstat, nil, "ss")
	registerFilter(filterNetstat, nil, "netstat")
	registerFilter(filterDf, nil, "df")

	// Simple with aliases
	registerFilter(filterCompiler, nil, "gcc")
	registerAlias("gcc", "g++", "cc", "c++", "clang", "clang++")

	registerAlias("df", "du")

	// --- Hidden (autoDetect fallbacks, not shown in `chop list`) ---
	registerHidden(filterAutoDetect, "cat", "tail", "less", "more")
	registerHidden(filterAutoDetect, "ls")
	registerHidden(filterAutoDetect, "find")
	registerHidden(filterAutoDetect, "node", "node16", "node18", "node20", "node22")
}

// ListBuiltins returns all commands that have built-in filters,
// auto-generated from the registry. Aliases and hidden entries are excluded.
func ListBuiltins() []BuiltinCommand {
	commands := make([]string, 0, len(registry))
	for cmd := range registry {
		commands = append(commands, cmd)
	}
	sort.Strings(commands)

	var result []BuiltinCommand
	for _, cmd := range commands {
		e := registry[cmd]
		if e.hidden || aliases[cmd] {
			continue
		}
		if len(e.builtins) > 0 {
			for _, sub := range e.builtins {
				result = append(result, BuiltinCommand{Command: cmd, Subcommand: sub})
			}
		} else {
			result = append(result, BuiltinCommand{Command: cmd})
		}
	}
	return result
}