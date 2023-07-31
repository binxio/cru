package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	go_git_ssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	"io/ioutil"
	"os"
	"testing"
)

func TestClone(t *testing.T) {

	s := fmt.Sprintf("%s/.ssh/id_rsa", os.Getenv("HOME"))
	sshKey, err := ioutil.ReadFile(s)
	signer, err := ssh.ParsePrivateKey([]byte(sshKey))
	auth := &go_git_ssh.PublicKeys{User: "git", Signer: signer}
	r, err := git.Clone(memory.NewStorage(), memfs.New(), &git.CloneOptions{
		URL:  "git@github.com:binxio/google-pubsub-testbench.git",
		Auth: auth,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = r.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"},
		Auth:     auth,
	})
	if err != nil {
		t.Fatal(err)
	}

	wt, err := r.Worktree()
	if err != nil {
		t.Fatal(err)
	}

	err = wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("cloud_implementation"),
		Force:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
}
