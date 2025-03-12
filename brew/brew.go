package brew

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"

	"github.com/charmbracelet/log"
)

type Brew struct {
	log *log.Logger
}

func New(log *log.Logger) *Brew {
	os.Setenv("HOMEBREW_NO_INSTALL_FROM_API", "1")
	os.Setenv("HOMEBREW_NO_ANALYTICS", "1")
	b := &Brew{log}
	b.Run("brew", "tap", "homebrew/core", "--force")
	return b
}

func (b *Brew) GetPackages(packages ...string) []Package {

	// Get all dependencies
	dependencyList := b.RunWithOutput(
		"brew",
		append([]string{
			"deps",
			"--formula",
			"--full-name",
			"--union",
			"--topological",
			"--include-implicit",
			"--include-build",
			"--include-test",
		}, packages...)...,
	)

	// Add root packages to the list
	dependencySet := set[string]{}
	dependencySet.Add(dependencyList...)
	for _, pkg := range packages {
		if !dependencySet.Contains(pkg) {
			dependencyList = append(dependencyList, pkg)
		}
	}

	// Get package information
	packageDataBlob := b.RunWithFullOutput(
		"brew",
		append([]string{
			"info",
			"--formula",
			"--json",
		}, dependencyList...)...,
	)

	// Parse package information
	packageDataJson := make([]struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		Versions struct {
			Stable string `json:"stable"`
		} `json:"versions"`
		Revision int `json:"revision"`
		Bottles  struct {
			Stable struct {
				Files set[string] `json:"files"`
			} `json:"stable"`
		} `json:"bottle"`
		PostInstallDefined bool `json:"post_install_defined"`
	}, 0)

	if err := json.Unmarshal(packageDataBlob, &packageDataJson); err != nil {
		b.log.Fatal(err)
	}

	// Normalize package information
	packageData := make([]Package, len(packageDataJson))
	for i, packageJson := range packageDataJson {
		packageData[i] = Package{
			b,
			packageJson.Name,
			packageJson.FullName,
			packageJson.Versions.Stable,
			packageJson.Bottles.Stable.Files,
			packageJson.PostInstallDefined,
		}
	}

	return packageData
}

func (b *Brew) Run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		b.log.Fatal(err)
	}
}

func (b *Brew) RunWithOutput(name string, args ...string) []string {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		b.log.Fatal(err)
	}
	scanner := bufio.NewScanner(stdout)
	lines := make([]string, 0)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			lines = append(lines, line)
		}
	}()
	if err := cmd.Run(); err != nil {
		b.log.Fatal(err)
	}
	return lines
}

func (b *Brew) RunWithFullOutput(name string, args ...string) []byte {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	blob, err := cmd.Output()
	if err != nil {
		b.log.Fatal(err)
	}
	return blob
}
