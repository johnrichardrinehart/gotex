package main

import (
	"fmt"
	"github.com/fuzzybear3965/gotex/internal/parser"
	"os"
	"os/exec"
	"path/filepath"
)

func compile(c chan []*parser.DBRow) {
	rows := <-c
	fmt.Println("Received the rows.")
	for _, row := range rows {
		fmt.Println("%+v", row)
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
		// go ahead and fetch the latest version of the repository (in case we
		// haven't already)
		//runCommand(exec.Command("git reset", "--hard"), repopath) // reset the state of the repository
	}
}

func runCommand(c *exec.Cmd, p string) {
	curPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	os.Chdir(p) // change to path to run command
	fmt.Println("Running", c.Args)
	e := c.Start()
	if e != nil {
		panic(e)
	}
	fmt.Printf("%v finished with error: %v", c.Args, e)
	os.Chdir(curPath) // go back to where we started
}
