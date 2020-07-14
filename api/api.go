package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/itpp-labs/hound/codesearch/regexp"
	"github.com/itpp-labs/hound/config"
	"github.com/itpp-labs/hound/index"
	"github.com/itpp-labs/hound/searcher"
)

const (
	defaultLinesOfContext uint = 2
	maxLinesOfContext     uint = 20
)

type Stats struct {
	FilesOpened  int
	ReposScanned int
	Duration     int
}
type ReposPagination struct {
	NextOffset int
	OtherRepos int
	NextLimit  int
}

func writeJson(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Panicf("Failed to encode JSON: %v\n", err)
	}
}

func writeResp(w http.ResponseWriter, data interface{}) {
	writeJson(w, data, http.StatusOK)
}

func writeError(w http.ResponseWriter, err error, status int) {
	writeJson(w, map[string]string{
		"Error": err.Error(),
	}, status)
}

type preSearchResult struct {
	repo    string
	res     *index.PreSearchResponse
	err     error
	cleanup waiter
}
type searchFilesResult struct {
	i    int
	repo string
	res  *index.SearchResponse
	err  error
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type waiter struct {
	ch     chan bool
	active bool
}

func makeWaiter() waiter {
	return waiter{make(chan bool, 1), true}
}
func (w *waiter) Wait() {
	<-w.ch
}
func (w *waiter) Do() {
	if w.active {
		w.active = false
		w.ch <- true
	}
}

type limiter struct {
	ch     chan bool
	active bool
}

func makeLimiter(n int) limiter {
	return limiter{make(chan bool, n), true}
}

func (l *limiter) Acquire() bool {
	wasActive := l.active
	if wasActive {
		// wait
		l.ch <- true
	}
	if l.active {
		return true
	} else if wasActive {
		l.Release()
	}
	return false
}

func (l *limiter) Release() {
	<-l.ch
}
func (l *limiter) Close() {
	l.active = false
	// no need to close channel
	// see https://stackoverflow.com/a/8593986/222675
}

/**
 * Searches all repos in parallel with respecting to max-concurrent-searchers param
 */
func searchAll(
	query string,
	opts *index.SearchOptions,
	cfg *config.Config,
	repos []string,
	limitRepos int,
	idx map[string]*searcher.Searcher,
	reposScanned *int,
	filesOpened *int,
	duration *int) ([]*index.SearchResponse, int, error) {
	startedAt := time.Now()

	n := len(repos)
	searchersNum := min(n, cfg.MaxConcurrentSearchers)
	limiter := makeLimiter(searchersNum)

	// use a buffered channel to avoid routine leaks on errs.
	ch := make(chan *preSearchResult, n)
	enoughResults := makeWaiter()
	var err error
	firstUndone := 0
	results := map[string]*preSearchResult{}
	go func() {
		defer limiter.Close()
		defer enoughResults.Do()
		resNum := 0
		for i := 0; i < n; i++ {
			r := <-ch
			results[r.repo] = r
			if r.err != nil {
				err = r.err
				return
			}
			if !r.res.Found {
				r.cleanup.Do()
				continue
			}
			firstUndone, resNum = checkResults(repos, results, firstUndone)
			if resNum >= limitRepos {
				break
			}
		}
		if resNum < limitRepos {
			firstUndone, resNum = checkResults(repos, results, firstUndone)
		}
	}()

	var wg sync.WaitGroup
	for _, repo := range repos {
		if !limiter.Acquire() {
			break
		}
		wg.Add(1)
		go func(repo string) {
			defer idx[repo].SearchCleanUp()
			fms, err := idx[repo].PreSearch(query, opts)
			cleanup := makeWaiter()
			r := &preSearchResult{repo, fms, err, cleanup}
			// send result
			ch <- r
			wg.Done()
			// next worker can make PreSearch
			limiter.Release()
			// wait before calling SearchCleanUp
			cleanup.Wait()
		}(repo)
	}
	enoughResults.Wait()
	if err != nil {
		return nil, 0, err
	}

	// cleanup excess repos right now
	for i := firstUndone; i < n; i++ {
		repo := repos[i]
		r, processed := results[repo]
		if !processed {
			continue
		}
		r.cleanup.Do()
	}
	// cleanup other repos at the end
	defer func() {
		wg.Wait()
		close(ch)
		for r := range ch {
			results[r.repo] = r
		}
		for _, res := range results {
			*reposScanned += 1
			res.cleanup.Do()
		}
	}()

	// Grep files
	chFiles := make(chan *searchFilesResult, firstUndone)
	foundNum := 0
	var NextOffsetRepos int
	var i int
	for i = 0; i < firstUndone; i++ {
		repo := repos[i]
		r := results[repo]
		if !r.res.Found {
			continue
		}
		if foundNum >= limitRepos {
			break
		}
		go func(repo string, r *preSearchResult, i int) {
			searchRes, err := idx[repo].Search(r.res, opts)
			chFiles <- &searchFilesResult{i, repo, searchRes, err}
		}(repo, r, foundNum)
		foundNum++
	}
	NextOffsetRepos = i

	finalMap := map[int]*index.SearchResponse{}
	for i := 0; i < foundNum; i++ {
		res := <-chFiles
		if res.err != nil {
			return nil, 0, res.err
		}
		repo := res.repo
		results[repo].cleanup.Do()
		searchRes := res.res
		if searchRes.Matches == nil {
			continue
		}
		*filesOpened += searchRes.FilesOpened
		finalMap[res.i] = searchRes
	}

	final := []*index.SearchResponse{}
	for i := 0; i < foundNum; i++ {
		final = append(final, finalMap[i])
	}

	*duration = int(time.Now().Sub(startedAt).Seconds() * 1000)

	return final, NextOffsetRepos, nil
}
func checkResults(repos []string, results map[string]*preSearchResult, firstUndone int) (int, int) {
	resNum := 0
	var i int
	n := len(repos)
	for i = firstUndone; i < n; i++ {
		r, processed := results[repos[i]]
		if !processed {
			// Some of first repos are not processed
			return i, resNum
		}
		if r.res.Found {
			resNum++
		}
	}
	return i, resNum
}

// Used for parsing flags from form values.
func parseAsBool(v string) bool {
	v = strings.ToLower(v)
	return v == "true" || v == "1" || v == "fosho"
}

func parseAsRepoList(v string, idx map[string]*searcher.Searcher, offsetRepos int, orderedRepos []*config.Repo) []string {
	v = strings.TrimSpace(v)
	var repos []string
	if v == "" {
		v = ".*"
	}
	if v == "*" {
		// Backward compatibility
		v = ".*"
	}
	if strings.Contains(v, ",") {
		// Backward compatibility
		// This also means, that repo name in config cannot have commas
		var new_v []string
		for _, repo := range strings.Split(v, ",") {
			repo_regexp := "^" + repo + "$"
			new_v = append(new_v, repo_regexp)
		}
		v = strings.Join(new_v, "|")
	}

	re, _ := regexp.Compile(v)
	num := 0

	for _, repo := range orderedRepos {
		if re.MatchString(repo.Name, true, true) < 0 {
			// repo doesn't pass regexp
			continue
		}
		num++
		if num <= offsetRepos {
			continue
		}
		repos = append(repos, repo.Name)
	}
	return repos
}

func parseAsUintValue(sv string, min, max, def uint) uint {
	iv, err := strconv.ParseUint(sv, 10, 54)
	if err != nil {
		return def
	}
	if max != 0 && uint(iv) > max {
		return max
	}
	if min != 0 && uint(iv) < min {
		return max
	}
	return uint(iv)
}

func parseRangeInt(v string, i *int) {
	*i = 0
	if v == "" {
		return
	}

	vi, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return
	}

	*i = int(vi)
}

func parseRangeValue(rv string) (int, int) {
	ix := strings.Index(rv, ":")
	if ix < 0 {
		return 0, 0
	}

	var b, e int
	parseRangeInt(rv[:ix], &b)
	parseRangeInt(rv[ix+1:], &e)
	return b, e
}

func Setup(m *http.ServeMux, idx map[string]*searcher.Searcher, cfg *config.Config) {

	m.HandleFunc("/api/v1/repos", func(w http.ResponseWriter, r *http.Request) {
		res := map[string]*config.Repo{}
		for name, srch := range idx {
			res[name] = srch.Repo
		}

		writeResp(w, res)
	})

	m.HandleFunc("/api/v1/search", func(w http.ResponseWriter, r *http.Request) {
		var opt index.SearchOptions

		stats := parseAsBool(r.FormValue("stats"))
		offsetRepos, limitRepos := parseRangeValue(r.FormValue("rngRepos"))
		if limitRepos == 0 {
			limitRepos = cfg.MaxReposInFirstResult
		}
		repos := parseAsRepoList(r.FormValue("repos"), idx, offsetRepos, cfg.Repos)
		query := r.FormValue("q")
		opt.Offset, opt.Limit = parseRangeValue(r.FormValue("rng"))
		opt.FileRegexp = r.FormValue("files")
		opt.ExcludeFileRegexp = r.FormValue("excludeFiles")
		opt.IgnoreCase = parseAsBool(r.FormValue("i"))
		opt.LinesOfContext = parseAsUintValue(
			r.FormValue("ctx"),
			0,
			maxLinesOfContext,
			defaultLinesOfContext)

		var filesOpened int
		var durationMs int
		var reposScanned int

		results, nextOffsetRepos, err := searchAll(query, &opt, cfg, repos, limitRepos, idx, &reposScanned, &filesOpened, &durationMs)
		if err != nil {
			// TODO(knorton): Return ok status because the UI expects it for now.
			writeError(w, err, http.StatusOK)
			return
		}
		var res struct {
			Results         []*index.SearchResponse
			Stats           *Stats `json:",omitempty"`
			ReposPagination *ReposPagination
		}

		res.Results = results
		if stats {
			res.Stats = &Stats{
				FilesOpened:  filesOpened,
				ReposScanned: reposScanned,
				Duration:     durationMs,
			}
		}
		res.ReposPagination = &ReposPagination{
			NextOffset: offsetRepos + nextOffsetRepos,
			NextLimit:  cfg.MaxReposInNextResult,
			OtherRepos: len(repos) - nextOffsetRepos,
		}

		writeResp(w, &res)
	})

	m.HandleFunc("/api/v1/excludes", func(w http.ResponseWriter, r *http.Request) {
		repo := r.FormValue("repo")
		res := idx[repo].GetExcludedFiles()
		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		w.Header().Set("Access-Control-Allow", "*")
		fmt.Fprint(w, res)
	})

	m.HandleFunc("/api/v1/update", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			writeError(w,
				errors.New(http.StatusText(http.StatusMethodNotAllowed)),
				http.StatusMethodNotAllowed)
			return
		}

		repos := parseAsRepoList(r.FormValue("repos"), idx, 0, cfg.Repos)

		for _, repo := range repos {
			searcher := idx[repo]
			if searcher == nil {
				writeError(w,
					fmt.Errorf("No such repository: %s", repo),
					http.StatusNotFound)
				return
			}

			if !searcher.Update() {
				writeError(w,
					fmt.Errorf("Push updates are not enabled for repository %s", repo),
					http.StatusForbidden)
				return

			}
		}

		writeResp(w, "ok")
	})
}
