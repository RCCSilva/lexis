package library

import (
	"math"
	"sort"
	"strings"
	"unicode"

	"github.com/RCCSilva/lexis/internal/store"
)

const (
	bm25K1 = 1.2
	bm25B  = 0.75
	rrfK   = 60 // reciprocal rank fusion constant
)

// tokenize lowercases s and splits it into words. References such as
// "8-7-5" or "12-2-8(b)" are kept whole (as "8-7-5" / "12-2-8") so they can
// be matched literally; everything else is split on non-alphanumerics.
func tokenize(s string) []string {
	alnum := func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }
	var out []string
	for _, w := range strings.Fields(strings.ToLower(s)) {
		w = strings.TrimFunc(w, func(r rune) bool { return !alnum(r) })
		if w == "" {
			continue
		}
		if strings.Contains(w, "-") && strings.ContainsFunc(w, unicode.IsDigit) {
			end := strings.IndexFunc(w, func(r rune) bool { return !alnum(r) && r != '-' })
			if end < 0 {
				end = len(w)
			}
			out = append(out, strings.TrimRight(w[:end], "-"))
			continue
		}
		out = append(out, strings.FieldsFunc(w, func(r rune) bool { return !alnum(r) })...)
	}
	return out
}

// bm25 ranks docs against query and returns the n best with a positive score.
func bm25(query string, docs []store.Document, n int) []store.Result {
	terms := tokenize(query)
	if len(terms) == 0 || len(docs) == 0 {
		return nil
	}

	type counted struct {
		tf  map[string]int
		len int
	}
	counts := make([]counted, len(docs))
	df := make(map[string]int)
	var total int
	for i, d := range docs {
		toks := tokenize(d.Text)
		tf := make(map[string]int)
		for _, t := range toks {
			tf[t]++
		}
		for t := range tf {
			df[t]++
		}
		counts[i] = counted{tf: tf, len: len(toks)}
		total += len(toks)
	}
	avg := float64(total) / float64(len(docs))
	if avg == 0 {
		return nil
	}

	seen := make(map[string]bool, len(terms))
	var results []store.Result
	for i, d := range docs {
		var score float64
		clear(seen)
		for _, t := range terms {
			if seen[t] {
				continue
			}
			seen[t] = true
			f := float64(counts[i].tf[t])
			if f == 0 {
				continue
			}
			idf := math.Log(1 + (float64(len(docs))-float64(df[t])+0.5)/(float64(df[t])+0.5))
			norm := f * (bm25K1 + 1) / (f + bm25K1*(1-bm25B+bm25B*float64(counts[i].len)/avg))
			score += idf * norm
		}
		if score > 0 {
			results = append(results, store.Result{Document: d, Score: float32(score)})
		}
	}
	sort.SliceStable(results, func(a, b int) bool { return results[a].Score > results[b].Score })
	if len(results) > n {
		results = results[:n]
	}
	return results
}

// fuse merges the two rankings with reciprocal rank fusion and keeps topK.
// Hit.Score is the vector similarity, or 0 for keyword-only hits.
func fuse(semantic, keyword []store.Result, topK int) []Hit {
	type entry struct {
		hit   Hit
		score float64
	}
	entries := make(map[string]*entry)
	var order []string

	add := func(results []store.Result, kind string, keepScore bool) {
		for rank, r := range results {
			e, ok := entries[r.ID]
			if !ok {
				e = &entry{hit: Hit{Result: r, Match: kind}}
				if !keepScore {
					e.hit.Score = 0
				}
				entries[r.ID] = e
				order = append(order, r.ID)
			} else {
				e.hit.Match = "semantic+keyword"
			}
			e.score += 1 / float64(rrfK+rank+1)
		}
	}
	add(semantic, "semantic", true)
	add(keyword, "keyword", false)

	sort.SliceStable(order, func(a, b int) bool { return entries[order[a]].score > entries[order[b]].score })
	if len(order) > topK {
		order = order[:topK]
	}
	hits := make([]Hit, len(order))
	for i, id := range order {
		hits[i] = entries[id].hit
	}
	return hits
}
