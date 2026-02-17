package aclstore

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidType = errors.New("invalid ACL type")
	ErrNotFound    = errors.New("entry not found")
)

const (
	TypeURLs        = "urls"
	TypeDomains     = "domains"
	TypeExpressions = "expressions"
)

var validTypes = map[string]string{
	"url":         TypeURLs,
	"urls":        TypeURLs,
	"domain":      TypeDomains,
	"domains":     TypeDomains,
	"er":          TypeExpressions,
	"expression":  TypeExpressions,
	"expressions": TypeExpressions,
}

func NormalizeType(v string) (string, error) {
	n, ok := validTypes[strings.ToLower(strings.TrimSpace(v))]
	if !ok {
		return "", ErrInvalidType
	}
	return n, nil
}

type Store struct {
	BaseDir string
}

func New(baseDir string) *Store {
	return &Store{BaseDir: baseDir}
}

func (s *Store) groupsDir() string {
	return filepath.Join(s.BaseDir, "lists")
}

func (s *Store) aclPath(group, aclType string) string {
	return filepath.Join(s.groupsDir(), group, aclType)
}

func (s *Store) Add(group, aclType, value string) error {
	entries, err := s.readSet(group, aclType)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		entries = map[string]struct{}{}
	}
	entries[value] = struct{}{}
	return s.writeSet(group, aclType, entries)
}

func (s *Store) Delete(group, aclType, value string) error {
	entries, err := s.readSet(group, aclType)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrNotFound
		}
		return err
	}
	if _, ok := entries[value]; !ok {
		return ErrNotFound
	}
	delete(entries, value)
	return s.writeSet(group, aclType, entries)
}

func (s *Store) Purge(group string, aclType string) error {
	if aclType != "" {
		p := s.aclPath(group, aclType)
		if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	}

	gdir := filepath.Join(s.groupsDir(), group)
	if err := os.RemoveAll(gdir); err != nil {
		return err
	}
	return nil
}

func (s *Store) List(group string, aclType string) (map[string][]string, error) {
	result := map[string][]string{}
	groups, err := s.listGroups(group)
	if err != nil {
		return nil, err
	}
	for _, g := range groups {
		types := []string{TypeDomains, TypeURLs, TypeExpressions}
		if aclType != "" {
			types = []string{aclType}
		}
		for _, t := range types {
			set, err := s.readSet(g, t)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					continue
				}
				return nil, err
			}
			key := fmt.Sprintf("%s/%s", g, t)
			for v := range set {
				result[key] = append(result[key], v)
			}
			sort.Strings(result[key])
		}
	}
	return result, nil
}

func (s *Store) listGroups(group string) ([]string, error) {
	if group != "" {
		return []string{group}, nil
	}
	entries, err := os.ReadDir(s.groupsDir())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var groups []string
	for _, e := range entries {
		if e.IsDir() {
			groups = append(groups, e.Name())
		}
	}
	sort.Strings(groups)
	return groups, nil
}

func (s *Store) readSet(group, aclType string) (map[string]struct{}, error) {
	set := map[string]struct{}{}
	f, err := os.Open(s.aclPath(group, aclType))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		set[line] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return set, nil
}

func (s *Store) writeSet(group, aclType string, values map[string]struct{}) error {
	dir := filepath.Join(s.groupsDir(), group)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	var lines []string
	for v := range values {
		lines = append(lines, v)
	}
	sort.Strings(lines)

	tmp := s.aclPath(group, aclType) + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	for _, l := range lines {
		if _, err := w.WriteString(l + "\n"); err != nil {
			f.Close()
			return err
		}
	}
	if err := w.Flush(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, s.aclPath(group, aclType))
}

type QueryEngine struct {
	store *Store
	mu    sync.RWMutex
	cache map[string]cachedSet
}

type cachedSet struct {
	modTime time.Time
	values  []string
}

func NewQueryEngine(store *Store) *QueryEngine {
	return &QueryEngine{store: store, cache: map[string]cachedSet{}}
}

func (q *QueryEngine) Match(group, domain, uri string) (string, bool, error) {
	if ok, err := q.matchDomain(group, domain); err != nil {
		return "", false, err
	} else if ok {
		return "domain", true, nil
	}
	if ok, err := q.matchContains(group, TypeURLs, uri); err != nil {
		return "", false, err
	} else if ok {
		return "url", true, nil
	}
	if ok, err := q.matchExpression(group, uri); err != nil {
		return "", false, err
	} else if ok {
		return "expression", true, nil
	}
	return "", false, nil
}

func (q *QueryEngine) matchDomain(group, domain string) (bool, error) {
	entries, err := q.load(group, TypeDomains)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	domain = strings.ToLower(strings.TrimSpace(domain))
	for _, k := range entries {
		k = strings.ToLower(k)
		if domain == k || strings.HasSuffix(domain, "."+k) {
			return true, nil
		}
	}
	return false, nil
}

func (q *QueryEngine) matchContains(group, aclType, target string) (bool, error) {
	entries, err := q.load(group, aclType)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	for _, e := range entries {
		if strings.Contains(target, e) {
			return true, nil
		}
	}
	return false, nil
}

func (q *QueryEngine) matchExpression(group, uri string) (bool, error) {
	entries, err := q.load(group, TypeExpressions)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	for _, pattern := range entries {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		if re.MatchString(uri) {
			return true, nil
		}
	}
	return false, nil
}

func (q *QueryEngine) load(group, aclType string) ([]string, error) {
	path := q.store.aclPath(group, aclType)
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	q.mu.RLock()
	cached, ok := q.cache[path]
	q.mu.RUnlock()
	if ok && st.ModTime().Equal(cached.modTime) {
		return cached.values, nil
	}

	set, err := q.store.readSet(group, aclType)
	if err != nil {
		return nil, err
	}
	vals := make([]string, 0, len(set))
	for v := range set {
		vals = append(vals, v)
	}
	sort.Strings(vals)

	q.mu.Lock()
	q.cache[path] = cachedSet{modTime: st.ModTime(), values: vals}
	q.mu.Unlock()

	return vals, nil
}
