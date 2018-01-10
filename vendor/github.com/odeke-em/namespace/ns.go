package namespace

// Enables you do define clauses that will be added to separate namespaces.
// By default anything in an undefined namespace is added to the global namespace.
// A namespace is defined by:
// namespace := "[" + token "]"
// token := [^/](ie not "/") + ("/" + token)*
// ie (non "/" character) + (optionally "/" + non "/" character)
//
// For example:
// [global]
// [pull/push]
//
// Empty spaces are skipped over and not

import (
	"fmt"
	"io"
	"strings"

	"github.com/odeke-em/go-utils/fread"
)

const (
	commonCommandDelim = "/"
	lBrace             = "["
	rBrace             = "]"

	GlobalNamespaceKey = "global"
)

type type_ uint

const (
	tUnknown type_ = iota
	tNamespace
	tClause
)

func classify(line string) type_ {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, lBrace) {
		return tNamespace
	}
	return tClause
}

type Namespace map[string][]string

func Parse(r io.Reader) (Namespace, error) {
	// Expecting the form
	// [command]
	// key=value
	lines := fread.Fread(r)

	return ParseCh(lines)
}

func ParseCh(lines <-chan string) (Namespace, error) {
	type nsSetter func(values ...string)
	ns := make(Namespace)
	makeNSSetter := func(nsKeys ...string) nsSetter {
		return func(values ...string) {
			for _, nsKey := range nsKeys {
				if nsKey == "" {
					nsKey = GlobalNamespaceKey
				}
				ns[nsKey] = append(ns[nsKey], values...)
			}
		}
	}

	globalNSSetter := makeNSSetter(GlobalNamespaceKey)
	var lastNSSetter nsSetter = globalNSSetter

	lineno := uint64(0)
	for line := range lines {
		typ := classify(line)
		lineno += 1

		switch typ {
		case tNamespace:
			namespaceKeys, err := parseOutNamespaceHeaders(line)
			if err != nil {
				return nil, err
			}
			lastNSSetter = makeNSSetter(namespaceKeys...)
		case tClause:
			clause, err := parseOutClause(line)
			if err != nil {
				return nil, err
			}
			if clause != "" {
				lastNSSetter(clause)
			}
		}
	}

	return ns, nil
}

func parseOutClause(line string) (string, error) {
	line = strings.TrimSpace(line)
	return line, nil
}

func parseOutNamespaceHeaders(line string) ([]string, error) {
	line = strings.TrimSpace(line)
	lbc, rbc := strings.Count(line, lBrace), strings.Count(line, rBrace)
	if lbc != 1 {
		return nil, fmt.Errorf("expecting exactly 1 %q got %s", lBrace, line)
	}
	if rbc != 1 {
		return nil, fmt.Errorf("expecting exactly 1 %q", rBrace)
	}
	lbi, rbi := strings.Index(line, lBrace), strings.Index(line, rBrace)
	if lbi >= rbi {
		return nil, fmt.Errorf("%q must be before %q", lBrace, rBrace)
	}
	namespaceName := line[lbi+1 : rbi]

	splits := strings.Split(namespaceName, commonCommandDelim)
	return prepareNamespaceKeys(splits), nil
}

func prepareNamespaceKeys(keys []string) []string {
	var cleaned []string
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key != "" {
			cleaned = append(cleaned, key)
		}
	}

	if len(keys) >= 1 && len(cleaned) < 1 {
		// They wanted the global namespace but due to our strict policy
		// of ignoring empty keys, we'll just reset to the globalNamespace
		cleaned = []string{GlobalNamespaceKey}
	}

	return cleaned
}
