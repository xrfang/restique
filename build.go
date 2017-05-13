//usr/local/go/bin/go run $0 $@ $(dirname `realpath $0`); exit
package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"time"
)

func assert(err error) {
	if err != nil {
		panic(err)
	}
}

type depSpec struct {
	url    string
	branch string
	commit string
	path   string
	isGit  bool
}

type buildConf struct {
	PRE_COMPILE_EXEC  []string
	POST_COMPILE_EXEC []string
	EXTRA_LDFLAGS     string
	BUILD_TAGS        string
}

var PROJ_ROOT, CMD string

func getDepends() (deps []depSpec) {
	f, err := os.Open("depends")
	if os.IsNotExist(err) {
		return
	}
	assert(err)
	defer f.Close()
	spec := bufio.NewScanner(f)
	for spec.Scan() {
		s := strings.TrimSpace(spec.Text())
		if len(s) == 0 || strings.HasPrefix(s, "#") {
			continue
		}
		spec := strings.Split(s, " ")
		if len(spec) > 2 {
			panic(fmt.Errorf("invalid spec: %s", s))
		}
		var gs depSpec
		gs.isGit = true
		gs.url = spec[0]
		u, err := url.Parse(gs.url)
		if err == nil && (u.Scheme == "http" || u.Scheme == "https") {
			gs.path = path.Join(u.Host, u.Path)
		} else {
			gs.path = gs.url
			u := strings.SplitN(gs.url, "@", 2)
			if len(u) == 2 {
				gs.path = strings.Replace(u[1], ":", "/", -1)
			} else {
				gs.isGit = false
			}
		}
		gs.commit = "HEAD"
		gs.branch = "master"
		if len(spec) == 2 {
			if !gs.isGit {
				panic(fmt.Errorf("cannot specify branch/commit for 'go get' spec"))
			}
			bnc := strings.SplitN(spec[1], "@", 2)
			if bnc[0] != "" {
				gs.branch = bnc[0]
			}
			if len(bnc) == 2 {
				gs.commit = bnc[1]
			}
		}
		deps = append(deps, gs)
	}
	return
}

func updDepends(deps []depSpec, full bool) (depRoots []string) {
	PROJ_SRC := path.Join(PROJ_ROOT, "src")
	rs := make(map[string]int)
	for _, repo := range deps {
		fmt.Printf("clone: %s %s@%s", repo.url, repo.branch, repo.commit)
		root := strings.SplitN(repo.path, "/", 2)[0]
		rs[root] = 1
		cd := path.Join(PROJ_SRC, repo.path)
		fi, err := os.Stat(path.Join(cd, ".git"))
		if err == nil && fi.IsDir() && !full {
			fmt.Println(" ...skipped")
			continue
		}
		fmt.Printf("\n")
		if !repo.isGit {
			assert(run("go", "get", repo.url))
			continue
		}
		args := []string{"clone", repo.url}
		if repo.commit == "HEAD" {
			args = append(args, "--depth", "1")
		}
		args = append(args, "--branch", repo.branch, "--single-branch", cd)
		assert(exec.Command("git", args...).Run())
		if repo.commit != "HEAD" {
			os.Chdir(cd)
			assert(exec.Command("git", "checkout", repo.commit).Run())
		}
	}
	for r := range rs {
		depRoots = append(depRoots, r)
	}
	return
}

func updGitIgnore(roots []string) {
	patterns := make(map[string]int)
	buf, err := ioutil.ReadFile(".gitignore")
	if os.IsNotExist(err) {
		f, err := os.Create(".gitignore")
		assert(err)
		defer f.Close()
		for _, p := range roots {
			f.WriteString(p + "\n")
		}
		return
	}
	assert(err)
	for _, p := range strings.Split(string(buf), "\n") {
		patterns[p] = 1
	}
	for _, p := range roots {
		patterns[p] = 1
	}
	f, err := os.Create(".gitignore")
	assert(err)
	defer f.Close()
	var ps []string
	for p := range patterns {
		p = strings.TrimSpace(p)
		if p != "" {
			ps = append(ps, p)
		}
	}
	sort.Strings(ps)
	for _, p := range ps {
		f.WriteString(p + "\n")
	}
}

func getGitInfo() (branch, hash string, revisions int) {
	o, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	assert(err)
	branch = strings.TrimSpace(string(o))
	o, err = exec.Command("git", "log", "-n1", "--pretty=format:%h").Output()
	assert(err)
	hash = string(o)
	o, err = exec.Command("git", "log", "--oneline").Output()
	revisions = len(strings.Split(string(o), "\n")) - 1
	return
}

func parseConf() (bc buildConf, err error) {
	f, err := os.Open(path.Join(PROJ_ROOT, "build.conf"))
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("[build.conf] not found, continue with defaults...")
			err = nil
		}
		return
	}
	defer f.Close()
	getCmd := func(cmdline string) []string {
		cmdline = strings.TrimSpace(cmdline)
		w := "/ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0192456789."
		if strings.ContainsAny(cmdline[0:1], w) {
			return strings.Split(cmdline, " ")
		}
		return strings.Split(cmdline[1:], cmdline[0:1])
	}
	lines := bufio.NewScanner(f)
	for lines.Scan() {
		line := strings.TrimSpace(lines.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		switch strings.ToUpper(key) {
		case "PRE_COMPILE_EXEC":
			bc.PRE_COMPILE_EXEC = getCmd(kv[1])
		case "POST_COMPILE_EXEC":
			bc.POST_COMPILE_EXEC = getCmd(kv[1])
		case "EXTRA_LDFLAGS":
			bc.EXTRA_LDFLAGS = strings.TrimSpace(kv[1])
		case "BUILD_TAGS":
			bc.BUILD_TAGS = strings.TrimSpace(kv[1])
		default:
			err = fmt.Errorf("Invalid configuration key: %s", key)
			return
		}
	}
	err = lines.Err()
	return
}

func run(args ...string) (err error) {
	exe := args[0]
	if exe == "go" {
		exe = "/usr/local/go/bin/go"
	}
	cmd := exec.Command(exe, args[1:]...)
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "PATH=") {
			cmd.Env = append(cmd.Env, e)
			break
		}
	}
	cmd.Env = append(cmd.Env, "GOPATH="+PROJ_ROOT, "GOBUILD_MODE="+CMD)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	CMD = path.Base(os.Args[0])
	PROJ_ROOT = os.Args[len(os.Args)-1]
	PROJ_NAME := path.Base(PROJ_ROOT)
	assert(os.Chdir(PROJ_ROOT))
	depRoots := updDepends(getDepends(), CMD == "sync")
	updGitIgnore(depRoots)
	if CMD == "sync" {
		return
	}
	fmt.Println()
	bc, err := parseConf()
	if err != nil {
		fmt.Printf("parseConf: %s\n", err)
		return
	}
	if len(bc.PRE_COMPILE_EXEC) > 0 {
		fmt.Printf("PRE_COMPILE_EXEC: %s\n", strings.Join(bc.PRE_COMPILE_EXEC,
			" "))
		err = run(bc.PRE_COMPILE_EXEC...)
		if err != nil {
			fmt.Printf("PRE_COMPILE_EXEC: %s\n", err)
			return
		}
	}
	branch, hash, revs := getGitInfo()
	args := []string{"go", "install"}
	if len(bc.BUILD_TAGS) > 0 {
		args = append(args, "-tags", bc.BUILD_TAGS)
	}
	ldflags := fmt.Sprintf(`%s -X main._G_BRANCH=%s -X main._G_HASH=%s
		-X main._G_REVS=%d -X main._BUILT_=%d`, bc.EXTRA_LDFLAGS, branch,
		hash, revs, time.Now().Unix())
	args = append(args, "-ldflags", ldflags, PROJ_NAME)
	err = run(args...)
	if err != nil {
		fmt.Printf("COMPILE: %s\n", err)
		return
	}
	if len(bc.POST_COMPILE_EXEC) > 0 {
		fmt.Printf("POST_COMPILE_EXEC: %s\n", strings.Join(
			bc.POST_COMPILE_EXEC, " "))
		err = run(bc.POST_COMPILE_EXEC...)
		if err != nil {
			fmt.Printf("POST_COMPILE_EXEC: %s", err)
			return
		}
	}
	if CMD == "run" {
		args := []string{path.Join(PROJ_ROOT, "bin", PROJ_NAME)}
		args = append(args, os.Args[1:len(os.Args)-1]...)
		assert(run(args...))
	}
}
