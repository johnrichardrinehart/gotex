package main

import (
	"bytes"
	"fmt"
	"github.com/fuzzybear3965/gotex/internal/parser"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
)

func compile(rows []*parser.Commit, c chan []*parser.Commit) {
	last := len(rows) - 1
	for i := range rows {
		row := rows[last-i]
		logger.Info.Printf("-- Compiling on commit %v.--\n", row.ID)
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

		bdir := filepath.Join("builds", row.Path, row.ID) // relative path
		rdir := filepath.Join("repos", row.Path)          // relative path
		diff(bdir, rdir, row.TeXRoot, latexmkArgs)

		if exist, err := exists(rootPDF); !exist && err == nil {
			logger.Info.Printf("Building root PDF file.\n")
			// 6) Build current LaTeX file
			if !runCommand(exec.Command("latexmk", append(latexmkArgs, curTeX)...), repopath) {
				logger.Error.Printf("latexmk %v failed on commit %v.\n", row.TeXRoot, row.ID)
			}

			// 7) move everything
			// move the current pdf
			if err := os.Rename(filepath.Join(WORKINGDIR, "repos", row.Path, fmt.Sprintf("%v.pdf", row.TeXRoot)), filepath.Join(WORKINGDIR, "builds", row.Path, row.ID, fmt.Sprintf("%v.pdf", row.TeXRoot))); err != nil {
				logger.Error.Printf("Could not move file %v.\n", fmt.Sprintf("repos/%v/%v.pdf", row.Path, row.TeXRoot))
			}

			// move the log file
			if err := os.Rename(filepath.Join(WORKINGDIR, "repos", row.Path, fmt.Sprintf("%v.log", row.TeXRoot)), filepath.Join(WORKINGDIR, "builds", row.Path, row.ID, fmt.Sprintf("%v.log", row.TeXRoot))); err != nil {
				logger.Error.Printf("Could not move file %v.\n", fmt.Sprintf("repos/%v/%v.log", row.Path, row.TeXRoot))
			}
		} else {
			logger.Error.Printf("I've already generated %v.\n", rootPDF)
		}
		// Clean the repo
		if !runCommand(exec.Command("git", "clean", "-f"), repopath) {
			logger.Error.Println("git clean failed.")
		}
		logger.Anything.Println() // create a space between commits
	}
	c <- rows
}

// func diff() builds a diff PDF named <commit>.diff.pdf - that is, the diff pdf
// is named for the older of the two committed LaTeX files used by latexdiff
//
// bdirrel is the directory where the builds are stored relative to the
// starting directory of gotex
// commit is the SHA-1 hash of the working commit
// fname is the filename of the root document to diff
// largs is an array of latexmk args to pass to exec.Command
func diff(bdirrel string, rdirrel string, fname string, largs []string) {
	commit := filepath.Base(bdirrel)
	bdirabs := filepath.Join(WORKINGDIR, bdirrel)
	rdirabs := filepath.Join(WORKINGDIR, rdirrel)
	// Work on the diff PDF if it hasn't been made, already
	diffpdfpath := filepath.Join(bdirabs, fmt.Sprintf("%v.diff.pdf", commit))
	logger.Info.Printf("Building diff PDF for %v.\n", filepath.Join(bdirrel, fmt.Sprintf("%v.diff.pdf", commit)))
	if exist, err := exists(diffpdfpath); !exist && err == nil {
		// 1) check out the previous commit (if possible)
		logger.Debug.Printf("Checking out old commit %v~1.\n", commit)
		if !runCommand(exec.Command("git", "checkout", fmt.Sprintf("%v~1", commit)), rdirabs) {
			logger.Warning.Println("Couldn't check out old commit (maybe this is the first commit?).")
		}
		logger.Debug.Printf("Changing the name of the root tex file to temp.tex (for diffing)\n.")
		// 2) change repos/a/b/main.tex -> repos/a/b/temp.tex
		if err := os.Rename(filepath.Join(rdirabs, fmt.Sprintf("%v.tex", fname)), filepath.Join(rdirabs, "temp.tex")); err != nil {
			logger.Fatal.Println(err)
		}
		logger.Debug.Printf("Checking out new commit %v.\n", commit)
		// 3) check out this row's commit
		if !runCommand(exec.Command("git", "checkout", fmt.Sprintf("%v", commit)), rdirabs) {
			logger.Error.Printf("Couldn't check out commit %v.", commit)
		}
		logger.Debug.Println("Generating diff LaTeX file.")
		// 4) Generate diff tex
		c := exec.Command("latexdiff", "temp.tex", fmt.Sprintf("%v.tex", fname))
		if stdout, err := c.StdoutPipe(); err != nil {
			logger.Error.Println(err)
		} else {
			curPath, err := os.Getwd() // store current path for later
			if err != nil {
				logger.Error.Println(err)
			}
			// chdir to repository
			if err := os.Chdir(rdirabs); err != nil {
				logger.Error.Println(err)
			}
			// start latexdiff
			if err := c.Start(); err != nil {
				logger.Error.Println(err)
			}
			// read the output into diff.tex
			f, err := os.Create("diff.tex")
			if err != nil {
				logger.Error.Println(err)
			}
			var b bytes.Buffer
			if _, err := b.ReadFrom(stdout); err != nil {
				logger.Error.Println(err)
			}
			if _, err := b.WriteTo(f); err != nil {
				logger.Error.Println(err)
			}
			f.Close()
			if err := c.Wait(); err != nil {
				logger.Error.Println(err)
			}
			// chdir back to where we started
			if err := os.Chdir(curPath); err != nil {
				logger.Error.Println(err)
			}
		}
		logger.Debug.Println("Generating diff pdf.")
		// 5) Generate diff pdf
		if !runCommand(exec.Command("latexmk", append(largs, "diff.tex")...), rdirabs) {
			logger.Error.Printf("latexmk the diff failed for commit %v and ancestor.", commit)
		}
		// move the diff pdf
		if err := os.Rename(filepath.Join(rdirabs, "diff.pdf"), filepath.Join(bdirabs, fmt.Sprintf("%v.diff.pdf", commit))); err != nil {
			logger.Error.Printf("Could not move diff.pdf to %v.\n", bdirabs)
		}
	} else {
		logger.Warning.Printf("I've already generated %v.\n", filepath.Join(bdirabs, fmt.Sprintf("%v.diff.pdf", commit)))
	}
}

func runCommand(c *exec.Cmd, p string) bool {
	curPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	// change to path to run command
	defer os.Chdir(curPath) // go back to where we started
	if err := os.Chdir(p); err != nil {
		logger.Error.Println(err)
	}
	// below taken from https://stackoverflow.com/questions/10385551/get-exit-code-go
	if err := c.Start(); err != nil {
		logger.Error.Println(err)
	}
	if _, err := os.Getwd(); err != nil {
		logger.Error.Println(err)
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
				logger.Error.Printf("%v failed with error code %v.\n\n", c.Args, err)
				return false
			}
		} else {
			logger.Error.Printf("%v encountered an error: %v.\n\n", c.Args, err)
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

func initRepo(row *parser.Commit) {
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
