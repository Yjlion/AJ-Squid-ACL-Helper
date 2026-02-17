package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"aj-squid-acl-helper/internal/aclstore"
)

func main() {
	baseDir := flag.String("base-dir", ".", "repository base directory")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		help()
		os.Exit(1)
	}

	store := aclstore.New(*baseDir)
	action := args[0]

	switch action {
	case "add":
		if len(args) != 4 {
			help()
			os.Exit(1)
		}
		runAdd(store, args[1], args[2], args[3])
	case "del":
		if len(args) != 4 {
			help()
			os.Exit(1)
		}
		runDelete(store, args[1], args[2], args[3])
	case "list":
		group, aclType := optionalGroupType(args[1:])
		runList(store, group, aclType)
	case "purge":
		if len(args) < 2 || len(args) > 3 {
			help()
			os.Exit(1)
		}
		group := args[1]
		aclType := ""
		if len(args) == 3 {
			var err error
			aclType, err = aclstore.NormalizeType(args[2])
			if err != nil {
				die(err)
			}
		}
		if err := store.Purge(group, aclType); err != nil {
			die(err)
		}
		if aclType == "" {
			fmt.Printf("Purged group %q\n", group)
		} else {
			fmt.Printf("Purged %q type %q\n", group, aclType)
		}
	case "make":
		fmt.Println("No-op in Go version: ACL data is plain-text and updated in place.")
	case "help":
		help()
	default:
		help()
		os.Exit(1)
	}
}

func runAdd(store *aclstore.Store, group, typ, value string) {
	aclType, err := aclstore.NormalizeType(typ)
	if err != nil {
		die(err)
	}
	if err := store.Add(group, aclType, value); err != nil {
		die(err)
	}
	fmt.Printf("Added: %s -> %s\n", value, filepath.Join("lists", group, aclType))
}

func runDelete(store *aclstore.Store, group, typ, value string) {
	aclType, err := aclstore.NormalizeType(typ)
	if err != nil {
		die(err)
	}
	err = store.Delete(group, aclType, value)
	if err != nil {
		if errors.Is(err, aclstore.ErrNotFound) {
			die(fmt.Errorf("entry %q not found", value))
		}
		die(err)
	}
	fmt.Printf("Deleted: %s\n", value)
}

func runList(store *aclstore.Store, group, typ string) {
	aclType := ""
	if typ != "" {
		var err error
		aclType, err = aclstore.NormalizeType(typ)
		if err != nil {
			die(err)
		}
	}
	rows, err := store.List(group, aclType)
	if err != nil {
		die(err)
	}
	keys := make([]string, 0, len(rows))
	for k := range rows {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("-> %s\n", k)
		for _, v := range rows[k] {
			fmt.Printf("   - %s\n", v)
		}
	}
}

func optionalGroupType(args []string) (group string, typ string) {
	if len(args) > 0 {
		group = args[0]
	}
	if len(args) > 1 {
		typ = args[1]
	}
	return group, typ
}

func help() {
	fmt.Println("Usage:")
	fmt.Println("  aclctl [--base-dir .] <list|purge|make> [group [type]]")
	fmt.Println("  aclctl [--base-dir .] <add|del> <group> <type> <value>")
	fmt.Println("Types: url, domain, er")
}

func die(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
