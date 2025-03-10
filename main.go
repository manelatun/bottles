package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"os/exec"
)

func main() {
	os.Setenv("HOMEBREW_NO_INSTALL_FROM_API", "1")
	os.Setenv("HOMEBREW_NO_INSTALL_UPGRADE", "1")
	os.Setenv("HOMEBREW_NO_AUTO_UPDATE", "1")
	os.Setenv("HOMEBREW_NO_ANALYTICS", "1")

	tap := exec.Command("brew", "tap", "homebrew/core", "--force")
	tap.Stdout = os.Stdout
	tap.Stderr = os.Stderr
	tap.Stdin = os.Stdin
	if err := tap.Run(); err != nil {
		log.Fatal(err)
	}

	packages := Packages(os.Args[1])
	postinstalls := PostInstalls(packages)

	for _, pkg := range packages {
		build := exec.Command("brew", "install", pkg, "--build-bottle", "--verbose")
		build.Stdout = os.Stdout
		build.Stderr = os.Stderr
		build.Stdin = os.Stdin
		if err := build.Run(); err != nil {
			log.Fatal(err)
		}

		bottle := exec.Command("brew", "bottle", pkg, "--verbose")
		bottle.Stdout = os.Stdout
		bottle.Stderr = os.Stderr
		bottle.Stdin = os.Stdin
		if err := bottle.Run(); err != nil {
			log.Fatal(err)
		}

		if postinstalls[pkg] {
			postinstall := exec.Command("brew", "postinstall", pkg, "--verbose")
			postinstall.Stdout = os.Stdout
			postinstall.Stderr = os.Stderr
			postinstall.Stdin = os.Stdin
			if err := postinstall.Run(); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func Packages(pkg string) []string {
	list := make([]string, 0)
	cmd := exec.Command("brew",
		"deps",
		pkg,
		"--topological",
		"--missing",
		"--full-name",
		"--include-implicit",
		"--include-build",
	)
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			list = append(list, line)
		}
	}()
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return append(list, pkg)
}

func PostInstalls(packages []string) map[string]bool {
	cmd := exec.Command("brew", append([]string{"info", "--json"}, packages...)...)
	cmd.Stderr = os.Stderr
	blob, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	infos := make([]struct {
		Name               string `json:"name"`
		PostInstallDefined bool   `json:"post_install_defined"`
	}, 0)
	if err := json.Unmarshal(blob, &infos); err != nil {
		log.Fatal(err)
	}

	postinstalls := make(map[string]bool)
	for _, info := range infos {
		postinstalls[info.Name] = info.PostInstallDefined
	}

	return postinstalls
}
