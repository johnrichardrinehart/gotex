package main

import (
	"fmt"
	"github.com/fuzzybear3965/gotex/internal/parser"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
)

func compile(rows []*parser.DBRow, c chan []*parser.DBRow) {
	last := len(rows) - 1
	for i := range rows {
		row := rows[last-i]
		logger.Printf("\tWorking on commit %v.\n", row.ID)
		initRepo(row)
		// Now we need to:
		// 1) Check out the commit previous to this row (if it exists)
		// 2) Change the root name of the previous LaTeX file to <row.ID>.tex
		// 3) Checkout the commit associated with this row
		// 4) Generate the <row.TexRoot>.diff.tex (latexdiff)
		// 5) Generate the <row.TexRoot>.diff.pdf (lualatex -pdf)
		// 6) Generate the <row.TexRoot>.pdf (lualatex -pdf)
		// 7) Move the files to the build/ directory from the repos/ directory

		// Check for the existence of the generated diff pdf, log file,
		// and LaTeX PDF
		repopath := fmt.Sprintf("repos/%v/", row.Path)
		curTeX := fmt.Sprintf("%v.tex", row.TeXRoot)
		oldTeX := fmt.Sprintf("%v.tex", row.ID) // just use this SHA-1 as a reference to the previous SHA-1
		diffTeX := fmt.Sprintf("%v.diff", row.TeXRoot)
		diffPDF := fmt.Sprintf("builds/%v/%v/%v.diff.pdf", row.Path, row.ID, row.TeXRoot)
		rootPDF := fmt.Sprintf("builds/%v/%v/%v.pdf", row.Path, row.ID, row.TeXRoot)

		var latexmkArgs = make([]string, 0)
		latexmkArgs = append(latexmkArgs,
			"-pdf",
			"-recorder",
			"-lualatex",
			"-verbose",
			"-halt-on-error",
			"-file-line-error",
			"-interaction=nonstopmode",
		)

		// Work on the diff PDF
		if exist, err := exists(diffPDF); !exist && err == nil {
			// 1) check out the previous commit (if possible)
			logger.Println("Checking out the previous commit.")
			if !runCommand(exec.Command("git", "checkout", fmt.Sprintf("%v~1", row.ID)), repopath) {
				logger.Println("Couldn't check out previous commit (maybe first commit?)")
			}

			logger.Println("Changing the name of the root tex file (for diffing).")
			// 2) change repos/a/b/main.tex -> repos/a/b/<sha-1>.tex
			if err := os.Rename(fmt.Sprintf("repos/%v/%v.tex", row.Path, row.TeXRoot), fmt.Sprintf("repos/%v/%v", row.Path, oldTeX)); err != nil {
				panic(err)
			}
			logger.Printf("Checking out commit %v.\n", row.ID)
			// 3) check out this row's commit
			if !runCommand(exec.Command("git", "checkout", fmt.Sprintf("%v", row.ID)), repopath) {
				logger.Println("Couldn't check out this commit.")
			}

			logger.Printf("Generating diff LaTeX file.\n")
			// 4) Generate diff tex
			if !runCommand(exec.Command("latexdiff", oldTeX, curTeX, ">", diffTeX), repopath) {
				logger.Printf("latexdiff failed on commit %v.\n", row.ID)
			}

			logger.Printf("Generating diff pdf.\n")
			// 5) Generate diff pdf
			if !runCommand(exec.Command("latexmk", append(latexmkArgs, diffTeX)...), repopath) {
				logger.Println("latexmk the diff failed")
			}

			// Rename the diff pdf
			if err := os.Rename(fmt.Sprintf("repos/%v/%v.pdf", row.Path, row.TeXRoot), diffPDF); err != nil {
				logger.Printf("Could not rename the diff pdf %v.\n", row.TeXRoot)
			}

			// move the diff pdf
			if err := os.Rename(fmt.Sprintf("repos/%v/%v.diff.pdf", row.Path, row.TeXRoot), fmt.Sprintf("build/%v/%v/%v.diff.pdf", row.Path, row.ID, row.TeXRoot)); err != nil {
				logger.Printf("Could not move file %v.\n", diffTeX)
			}
		} else {
			logger.Printf("I've already generated %v.\n", diffPDF)
		}

		if exist, err := exists(rootPDF); !exist && err == nil {
			logger.Printf("Building root PDF file.\n")
			// 6) Build current LaTeX file
			if !runCommand(exec.Command("latexmk", append(latexmkArgs, curTeX)...), repopath) {
				logger.Printf("latexmk %v failed.\n", row.TeXRoot)
			}

			// 7) move everything
			// move the current pdf
			if err := os.Rename(fmt.Sprintf("repos/%v/%v.pdf", row.Path, row.TeXRoot), fmt.Sprintf("builds/%v/%v/%v.pdf", row.Path, row.ID, row.TeXRoot)); err != nil {
				logger.Printf("Could not move file %v.\n", fmt.Sprintf("repos/%v/%v.pdf", row.Path, row.TeXRoot))
			}

			// move the log file
			if err := os.Rename(fmt.Sprintf("repos/%v/%v.log", row.Path, row.TeXRoot), fmt.Sprintf("builds/%v/%v/%v.log", row.Path, row.ID, row.TeXRoot)); err != nil {
				logger.Printf("Could not move file %v.\n", fmt.Sprintf("repos/%v/%v.log", row.Path, row.TeXRoot))
			}
		} else {
			logger.Printf("I've already generated %v.\n", rootPDF)
		}
		// Clean the repo
		if !runCommand(exec.Command("git", "clean", "-f"), repopath) {
			logger.Println("git clean failed.")
		}
		logger.Printf("\n")
	}
	c <- rows
}

func runCommand(c *exec.Cmd, p string) bool {
	curPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	os.Chdir(p)             // change to path to run command
	defer os.Chdir(curPath) // go back to where we started
	// below taken from https://stackoverflow.com/questions/10385551/get-exit-code-go
	if err := c.Start(); err != nil {
		logger.Printf("c.Start: %v", err)
	}
	if _, err := os.Getwd(); err != nil {
		panic(err)
	} else {
		//logger.Printf("Running %v at %v.\n", c.Args, wd)
	}
	if err := c.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0

			// This works on both Unix and Windows. Although package
			// syscall is generally platform dependent, WaitStatus is
			// defined for both Unix and Windows and in both cases has
			// an ExitStatus() method with the same signature.
			if s := exiterr.Success(); !s {
				//logger.Printf("%v failed.\n\n", c.Args)
				return false
			}
		} else {
			//logger.Printf("%v encountered an error: %v.\n\n", c.Args, err)
			return false
		}
	}
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

func initRepo(row *parser.DBRow) {
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
	e, err := exists(filepath.Join(repopath, ".git/"))
	if err != nil {
		panic(err)
	}
	if e {
		// go ahead and fetch the latest version of the repository
		runCommand(exec.Command("git", "reset", "--hard"), repopath)          // reset the state of the repository
		runCommand(exec.Command("git", "pull", "origin", "master"), repopath) // pull the latest version
	} else {
		// this repo is new... go ahead and clone it
		u, err := url.Parse(row.URL)
		if err != nil {
			panic(err)
		}
		scheme := u.Scheme
		url := fmt.Sprintf("%s://%s", scheme, row.Path)
		runCommand(exec.Command("git", "clone", url, repopath), ".") // pull the latest version
	}
}
