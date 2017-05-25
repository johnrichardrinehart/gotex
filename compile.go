package main

import (
	"fmt"
	"github.com/fuzzybear3965/gotex/internal/parser"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
)

func compile(c chan []*parser.DBRow) {
	rows := <-c
	for i, row := range rows {
		//buildPath := filepath.Join("./builds", row.Path, row.ID)
		err := os.MkdirAll(filepath.Join("./builds", row.Path, row.ID), os.ModePerm)
		if err != nil {
			panic(err)
		}
		repopath := filepath.Join("./repos", row.Path)
		err = os.MkdirAll(repopath, os.ModePerm)
		if err != nil {
			panic(err)
		}
		// only do git clone/pull stuff the first time (all repopaths should be the same for
		// all rows)
		if i == 0 {
			// Do Git stuff
			e, err := exists(filepath.Join(repopath, ".git/"))
			if err != nil {
				panic(err)
			}
			if e {
				// go ahead and fetch the latest version of the repository (in case we
				// haven't already)
				runCommand(exec.Command("git", "reset", "--hard"), repopath)          // reset the state of the repository
				runCommand(exec.Command("git", "pull", "origin", "master"), repopath) // pull the latest version
			} else { // clone the latest version
				u, err := url.Parse(row.URL)
				if err != nil {
					panic(err)
				}
				scheme := u.Scheme
				url := fmt.Sprintf("%s://%s", scheme, row.Path)
				runCommand(exec.Command("git", "clone", url, repopath), ".") // pull the latest version
			}
		}

		if runCommand(exec.Command("git", "checkout", row.ID), repopath) {
			// Do latexmk stuff for each commit ID
			if runCommand(exec.Command("latexmk", "-lualatex", "-verbose", "-halt-on-error", "-file-line-error", "-interaction=nonstopmode"), repopath) {
				fmt.Println("Compiled just great!")
				// clean up after compilation
				if runCommand(exec.Command("git", "clean", "-f"), repopath) {
					fmt.Println("Successfully executed git clean.")
				} else {
					fmt.Println("git clean failed.")
				}
			} else {
				fmt.Println("latexmk failed.")
			}
		} else {
			fmt.Printf("git checkout failed for id %v.\n", row.ID)
		}
		fmt.Println()
	}
}

func runCommand(c *exec.Cmd, p string) bool {
	curPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	os.Chdir(p) // change to path to run command
	// below taken from https://stackoverflow.com/questions/10385551/get-exit-code-go
	if err := c.Start(); err != nil {
		fmt.Printf("c.Start: %v", err)
	}
	if wd, err := os.Getwd(); err != nil {
		panic(err)
	} else {
		fmt.Printf("Running %v at %v.\n", c.Args, wd)
	}
	if err := c.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0

			// This works on both Unix and Windows. Although package
			// syscall is generally platform dependent, WaitStatus is
			// defined for both Unix and Windows and in both cases has
			// an ExitStatus() method with the same signature.
			if s := exiterr.Success(); !s {
				fmt.Printf("%v failed.\n\n", c.Args)
				return false
			}
		} else {
			fmt.Printf("%v encountered an error: %v.\n\n", c.Args, err)
			return false
		}
	}
	os.Chdir(curPath) // go back to where we started
	return true
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
