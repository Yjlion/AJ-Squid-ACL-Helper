package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"aj-squid-acl-helper/internal/aclstore"
)

func main() {
	baseDir := flag.String("base-dir", ".", "repository base directory")
	flag.Parse()

	store := aclstore.New(*baseDir)
	engine := aclstore.NewQueryEngine(store)

	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			fmt.Println("ERR log=invalid-input")
			continue
		}
		group, domain, uri := fields[0], fields[1], fields[2]

		started := time.Now()
		kind, matched, err := engine.Match(group, domain, uri)
		elapsed := time.Since(started).Seconds()
		if err != nil {
			fmt.Printf("ERR log=%v(%.3f)\n", err, elapsed)
			continue
		}
		if matched {
			fmt.Printf("OK log=match:%s/%s(%.3f)\n", group, kind, elapsed)
		} else {
			fmt.Printf("ERR log=(%.3f)\n", elapsed)
		}
	}
	if err := s.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "scan error: %v\n", err)
		os.Exit(1)
	}
}
