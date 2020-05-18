package main

import (
	"flag"
	"log"
	//"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	//"github.com/itpp-labs/hound/api"
	"github.com/itpp-labs/hound/config"
	"github.com/itpp-labs/hound/searcher"
	//"github.com/itpp-labs/hound/ui"
	"github.com/itpp-labs/hound/web"
)

const gracefulShutdownSignal = syscall.SIGTERM

var (
	info_log   *log.Logger
	error_log  *log.Logger
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

func makeSearchers(cfg *config.Config) (map[string]*searcher.Searcher, bool, error) {
	// Ensure we have a dbpath
	if _, err := os.Stat(cfg.DbPath); err != nil {
		if err := os.MkdirAll(cfg.DbPath, os.ModePerm); err != nil {
			return nil, false, err
		}
	}

	searchers, errs, err := searcher.MakeAll(cfg)
	if err != nil {
		return nil, false, err
	}

	if len(errs) > 0 {
		// NOTE: This mutates the original config so the repos
		// are not even seen by other code paths.
		for name, _ := range errs {
			delete(cfg.Repos, name)
		}

		return searchers, false, nil
	}

	return searchers, true, nil
}

func handleShutdown(shutdownCh <-chan os.Signal, searchers map[string]*searcher.Searcher) {
	go func() {
		<-shutdownCh
		info_log.Printf("Graceful shutdown requested...")
		for _, s := range searchers {
			s.Stop()
		}

		for _, s := range searchers {
			s.Wait()
		}

		os.Exit(0)
	}()
}

func registerShutdownSignal() <-chan os.Signal {
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, gracefulShutdownSignal)
	return shutdownCh
}

/*
// TODO: if we need this method, it should be update according to new method specs
func runHttp(
	addr string,
	dev bool,
	cfg *config.Config,
	idx map[string]*searcher.Searcher) error {
	m := http.DefaultServeMux

	h, err := ui.Content(dev, cfg)
	if err != nil {
		return err
	}

	m.Handle("/", h)
	api.Setup(m, idx)
	return http.ListenAndServe(addr, m)
}
*/

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	info_log = log.New(os.Stdout, "", log.LstdFlags)
	error_log = log.New(os.Stderr, "", log.LstdFlags)

	flagConf := flag.String("conf", "config.json", "")
	flagAddr := flag.String("addr", ":6080", "")
	flagDev := flag.Bool("dev", false, "")

	flag.Parse()

	var cfg config.Config
	if err := cfg.LoadFromFile(*flagConf); err != nil {
		panic(err)
	}

	// Start the web server on a background routine.
	ws := web.Start(&cfg, *flagAddr, *flagDev)

	// It's not safe to be killed during makeSearchers, so register the
	// shutdown signal here and defer processing it until we are ready.
	shutdownCh := registerShutdownSignal()
	idx, ok, err := makeSearchers(&cfg)
	if err != nil {
		log.Panic(err)
	}
	if !ok {
		info_log.Println("Some repos failed to index, see output above")
	} else {
		info_log.Println("All indexes built!")
	}

	handleShutdown(shutdownCh, idx)

	host := *flagAddr
	if strings.HasPrefix(host, ":") {
		host = "localhost" + host
	}

	if *flagDev {
		info_log.Printf("[DEV] starting webpack-dev-server at localhost:9000...")
		webpack := exec.Command("./node_modules/.bin/webpack-dev-server", "--mode", "development")
		webpack.Dir = basepath + "/../../"
		webpack.Stdout = os.Stdout
		webpack.Stderr = os.Stderr
		err = webpack.Start()
		if err != nil {
			error_log.Println(err)
		}
	}

	info_log.Printf("running server at http://%s\n", host)

	// Fully enable the web server now that we have indexes
	panic(ws.ServeWithIndex(idx))
}
