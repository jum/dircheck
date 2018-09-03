// dircheck.go - a simple command line tool to check for changes in a
// set of directories.
//
// jum@anubis.han.de

package main

import (
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

var (
	freeze = flag.String("f", "", "name of freeze file (required)")
)

type node struct {
	Name   string
	Size   int64
	Mode   string
	MTime  time.Time
	Childs []*node
	Hash   []byte
	seen   bool
}

func (n *node) String() string {
	return fmt.Sprintf("%s (size %v, mode %s, mtime %v, hash %v)", n.Name, n.Size, n.Mode, n.MTime, hex.EncodeToString(n.Hash))
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] dirs...\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 || len(*freeze) == 0 {
		flag.Usage()
	}
	tree := make(map[string]*node)
	for _, dirRoot := range flag.Args() {
		tree[dirRoot] = loadTree(dirRoot)
	}
	//spew.Dump(tree)
	fd, err := os.Open(*freeze)
	if err == nil {
		dec := gob.NewDecoder(fd)
		treeOld := make(map[string]*node)
		err = dec.Decode(&treeOld)
		fd.Close()
		if err != nil {
			panic(err)
		}
		//diff := deep.Equal(treeOld, tree)
		var diff []string
		for rootName, rootNode := range treeOld {
			newNode := tree[rootName]
			if newNode != nil {
				diff = append(diff, compareTree(rootNode, newNode)...)
			} else {
				// old subtrees
				diff = append(diff, fmt.Sprintf("directory %s not mentioned on command line", rootName))
			}
		}
		for rootName := range tree {
			oldNode := treeOld[rootName]
			if oldNode == nil {
				// new subtrees
				diff = append(diff, fmt.Sprintf("directory %s newly mentioned on command line", rootName))
			}
		}
		if len(diff) > 0 {
			for _, l := range diff {
				fmt.Printf("%v\n", l)
			}
		}
	}
	fd, err = os.Create(*freeze)
	if err != nil {
		panic(err)
	}
	enc := gob.NewEncoder(fd)
	err = enc.Encode(tree)
	fd.Close()
	if err != nil {
		panic(err)
	}
}

func loadTree(dirRoot string) *node {
	root := &node{}
	fd, err := os.Open(dirRoot)
	if err != nil {
		panic(err)
	}
	info, err := fd.Stat()
	if err != nil {
		panic(err)
	}
	if !info.IsDir() {
		panic(dirRoot + " is not a directory")
	}
	root.Name = info.Name()
	root.Size = info.Size()
	root.Mode = info.Mode().String()
	root.MTime = info.ModTime()
	defer fd.Close()
	entries, err := fd.Readdir(0)
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		n := &node{Name: entry.Name(), Size: entry.Size(), Mode: entry.Mode().String(), MTime: entry.ModTime()}
		if entry.IsDir() {
			n = loadTree(filepath.Join(dirRoot, entry.Name()))
		}
		if entry.Mode().IsRegular() {
			fd1, err := os.Open(filepath.Join(dirRoot, entry.Name()))
			if err != nil {
				panic(err)
			}
			hash := sha256.New()
			_, err = io.Copy(hash, fd1)
			if err != nil {
				panic(err.Error())
			}
			fd1.Close()
			n.Hash = hash.Sum(nil)
		}
		root.Childs = append(root.Childs, n)
	}
	return root
}

func compareTree(old, new *node) []string {
	var res []string
	for _, o := range old.Childs {
		n := findNode(new, o.Name)
		if n != nil {
			o.seen = true
			n.seen = true
			if len(o.Childs) > 0 || len(n.Childs) > 0 {
				res = append(res, compareTree(o, n)...)
			} else {
				os := o.String()
				ns := n.String()
				if os != ns {
					res = append(res, os+" != "+ns)
				}
			}
		}
	}
	for _, o := range old.Childs {
		if !o.seen {
			res = append(res, "removed "+o.String())
		}
	}
	for _, n := range new.Childs {
		if !n.seen {
			res = append(res, "new "+n.String())
		}
	}
	return res
}

func findNode(root *node, name string) *node {
	for _, n := range root.Childs {
		if name == n.Name {
			return n
		}
	}
	return nil
}
