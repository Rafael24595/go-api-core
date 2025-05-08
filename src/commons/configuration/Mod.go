package configuration

import (
	"bufio"
	"os"
	"strings"
)

type Mod struct {
	Module       string
	Version      string
	Dependencies map[string]Dependency
}

type Dependency struct {
	Module   string
	Version  string
	Replace  string
	Indirect bool
}

func DecodeMod(file *os.File) *Mod {
	version := ""
	module := ""
	dependencies := make(map[string]Dependency)
	
	inRequireBlock := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		if strings.HasPrefix(line, "module ") {
			module = strings.TrimSpace(strings.TrimPrefix(line, "module"))
			continue
		}

		if strings.HasPrefix(line, "go ") {
			version = strings.TrimSpace(strings.TrimPrefix(line, "go"))
			continue
		}

		if strings.HasPrefix(line, "require (") {
			inRequireBlock = true
			continue
		}

		if inRequireBlock {
			if line == ")" {
				inRequireBlock = false
				continue
			}
			
			dependency := makeDependency(line)
			dependencies[dependency.Module] = *dependency
			continue
		}

		if strings.HasPrefix(line, "require ") {
			dep := strings.TrimSpace(strings.TrimPrefix(line, "require"))
			dependency := makeDependency(dep)
			dependencies[dependency.Module] = *dependency
			continue
		}

		if strings.HasPrefix(line, "replace ") {
			fragments := strings.Split(strings.TrimSpace(strings.TrimPrefix(line, "replace")), "=>")
			if len(fragments) < 2 {
				panic("Bad format")
			}

			module := strings.TrimSpace(fragments[0])
			replace := strings.TrimSpace(fragments[1])
			if dependency, ok := dependencies[module]; ok {
				dependency.Replace = replace
				dependencies[module] = dependency
			}
		}
	}

	return &Mod{
		Module:       module,
		Version:      version,
		Dependencies: dependencies,
	}
}

func makeDependency(line string) *Dependency {
	fragments := strings.Split(strings.TrimSpace(line), " ")

	if len(fragments) < 2 {
		panic("Bad format")
	}

	module := fragments[0]
	version := fragments[1]
	indirect := false

	for i := 2; i < len(fragments); i++ {
		fragment := fragments[i]
		if fragment == "indirect" {
			indirect = true
		}
	}

	return &Dependency{
		Module:   module,
		Version:  version,
		Replace:  "",
		Indirect: indirect,
	}
}
