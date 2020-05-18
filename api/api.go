package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
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
	FilesOpened  int64
	ReposScanned int64
	Duration     int
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

type searchResponse struct {
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

/**
 * Searches all repos in parallel.
 */
func searchAll(
	query string,
	opts *index.SearchOptions,
	cfg *config.Config,
	repos []string,
	limitRepos int,
	idx map[string]*searcher.Searcher,
	reposScanned *int64,
	filesOpened *int64,
	duration *int) (map[string]*index.SearchResponse, error) {

	startedAt := time.Now()

	n := len(repos)
	searchersNum := min(n, cfg.MaxConcurrentSearchers)

	// use a buffered channel to avoid routine leaks on errs.
	ch := make(chan *searchResponse, searchersNum)
	chRepos := make(chan string)
	for _, repo := range repos {
		chRepos <- repo
	}
	close(chRepos)

	// stop workers when we have enough results
	quit := false

	for i := 0; i < searchersNum; i++ {
		go func() {
			for repo := range chRepos {
				fms, err := idx[repo].Search(query, opts)
				r := &searchResponse{repo, fms, err}
				atomic.AddInt64(filesOpened, int64(r.res.FilesOpened))
				atomic.AddInt64(reposScanned, 1)
				ch <- r
				if quit {
					return
				}
			}
		}()
	}

	hasFindings := map[string]bool{}
	res := map[string]*index.SearchResponse{}
	var lastIndex int
	for i := 0; i < n; i++ {
		r := <-ch
		if r.err != nil {
			quit = true
			return nil, r.err
		}

		if r.res.Matches == nil {
			hasFindings[r.repo] = false
			continue
		}
		hasFindings[r.repo] = true
		res[r.repo] = r.res

		lastIndex = enoughResults(repos, hasFindings, limitRepos)
		if lastIndex != -1 {
			break
		}
	}
	// delete excess results
	for i := lastIndex + 1; i < n; i++ {
		repo := repos[i]
		_, exists := res[repo]
		if exists {
			delete(res, repo)
		}
	}

	*duration = int(time.Now().Sub(startedAt).Seconds() * 1000)

	return res, nil
}
func enoughResults(repos []string, hasFindings map[string]bool, limitRepos int) int {
	resNum := 0
	var i int
	var repo string
	for i, repo = range repos {
		has, processed := hasFindings[repo]
		if !processed {
			// Some of first repos are not processed
			return -1
		}
		if has {
			resNum++
			if resNum >= limitRepos {
				break
			}
		}
	}
	return i
}

// Used for parsing flags from form values.
func parseAsBool(v string) bool {
	v = strings.ToLower(v)
	return v == "true" || v == "1" || v == "fosho"
}

func parseAsRepoList(v string, idx map[string]*searcher.Searcher, offsetRepos int) []string {
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
	// TODO: keep repos order the same as in config
	for repo := range idx {
		if re.MatchString(repo, true, true) < 0 {
			// repo doesn't pass regexp
			continue
		}
		num++
		if num <= offsetRepos {
			continue
		}
		repos = append(repos, repo)
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
		repos := parseAsRepoList(r.FormValue("repos"), idx, offsetRepos)
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

		var filesOpened int64
		var durationMs int
		var reposScanned int64

		results, err := searchAll(query, &opt, cfg, repos, limitRepos, idx, &reposScanned, &filesOpened, &durationMs)
		if err != nil {
			// TODO(knorton): Return ok status because the UI expects it for now.
			writeError(w, err, http.StatusOK)
			return
		}

		var res struct {
			Results map[string]*index.SearchResponse
			Stats   *Stats `json:",omitempty"`
		}

		res.Results = results
		if stats {
			res.Stats = &Stats{
				FilesOpened:  filesOpened,
				ReposScanned: reposScanned,
				Duration:     durationMs,
			}
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

		repos := parseAsRepoList(r.FormValue("repos"), idx, 0)

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
