# Hound

[![Build Status](https://travis-ci.org/itpp-labs/hound.svg?branch=master)](https://travis-ci.org/itpp-labs/hound) 

Hound is an extremely fast source code search engine. The core is based on this article (and code) from Russ Cox:
[Regular Expression Matching with a Trigram Index](http://swtch.com/~rsc/regexp/regexp4.html). Hound itself is a static
[React](http://facebook.github.io/react/) frontend that talks to a [Go](http://golang.org/) backend. The backend keeps an up-to-date index for each repository and answers searches through a minimal API. Here it is in action:

![Hound Screen Capture](screen_capture.gif)

Live demo: https://odoo-source.com/

## Quick Start Guide

### Using Go Tools

1. Use the Go tools to install Hound. The binaries `houndd` (server) and `hound` (cli) will be installed in your $GOPATH.

```
go get github.com/itpp-labs/hound/cmds/...
```

2. Create a [config.json](config-example.json) in a directory with your list of repositories.

3. Run the Hound server with `houndd` and you should see output similar to:
```
2015/03/13 09:07:42 Searcher started for statsd
2015/03/13 09:07:42 Searcher started for Hound
2015/03/13 09:07:42 All indexes built!
2015/03/13 09:07:42 running server at http://localhost:6080
```

### Using Docker (1.4+)

0. Configure access to github registry with [token](https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line):

       docker login docker.pkg.github.com -u GITHUB_USERNAME -p GITHUB_TOKEN 

1. Create a [config.json](config-example.json) in a directory with your list of repositories.

2. Run 

       docker run -d -p 6080:6080 --name hound -v $(pwd):/data docker.pkg.github.com/itpp-labs/hound/production


You should be able to navigate to [http://localhost:6080/](http://localhost:6080/) as usual.


## Running in Production

There are no special flags to run Hound in production. You can use the `--addr=:6880` flag to control the port to which the server binds. Currently, Hound does not support TLS as most users simply run Hound behind either Apache or nginx. Adding TLS support is pretty straight forward though if anyone wants to add it.

## Why Another Code Search Tool?

We've used many similar tools in the past, and most of them are either too slow, too hard to configure, or require too much software to be installed.
Which brings us to...

## Requirements
* Go 1.4+

Yup, that's it. You can proxy requests to the Go service through Apache/nginx/etc., but that's not required.


## Support

Currently Hound is only tested on MacOS and CentOS, but it should work on any *nix system. Hound on Windows is not supported but we've heard it compiles and runs just fine.

Hound supports the following version control systems: 

* Git - This is the default
* Mercurial - use `"vcs" : "hg"` in the config
* SVN - use `"vcs" : "svn"` in the config
* Bazaar - use `"vcs" : "bzr"` in the config

See [config-example.json](config-example.json) for examples of how to use each VCS.

## Private Repositories

There are a couple of ways to get Hound to index private repositories:

* Use the `file://` protocol. This allows you to index a local clone of a repository. The downside here is that the polling to keep the repo up to date will
not work. (This also doesn't work on local folders that are not of a supported repository type.)
* Use SSH style URLs in the config: `"url" : "git@github.com:foo/bar.git"`. As long as you have your 
[SSH keys](https://help.github.com/articles/generating-ssh-keys/) set up on the box where Hound is running this will work.

## Keeping Repos Updated

By default Hound polls the URL in the config for updates every 30 seconds. You can override this value by setting the `ms-between-poll` key on a per repo basis in the config. If you are indexing a large number of repositories, you may also be interested in tweaking the `max-concurrent-indexers` property. You can see how these work in the [example config](config-example.json).

## Search optimization

If you have large num of repositories you may be interested in tweaking following configs:

* `max-concurrent-searchers`
* `max-repos-in-first-result` -- instructs hound to don't show results from more repos than this number
* `max-repos-in-next-result` -- max num repos to show when users clicks "Load from other repos"

Note, that not every repo may have a search results. That means that to find some results in 10 repos, hound may scan 100 repos. Depending on your need and server copacity, you may set different value for `max-concurrent-searchers`. Let's assume that we have 1000 repos to scan and we want to show result from 10 repos (10 is a number either from `max-repos-in-first-result` or `max-repos-in-next-result` property). Let's say that we have 2 search requests: **reqA** -- gets findings in every first 10 repos, **reqB** -- requires to scan first 100 repos to get 10 repos with findings. Then we'll have following pictures depending on `max-concurrent-searchers` property:

* `max-concurrent-searchers` has value 200:

  * **reqA**:

    * 200-1000 repos are concurrently scanned by 200 searchers. The exact number
      of scanned repos depends on time for each repo and distribution of the
      repos between searchers: imagine that first repo requires very long time
      to process -- at this case the rest searchers will not stop working until
      we got 10 repos with findings
    * 190-990 results are ignored

  * **reqB**:

    * 200-1000 repos are concurrently scanned by 200 searchers
    * 100-900 results are ignored

* `max-concurrent-searchers` has value 20:

  * **reqA**:

    * 20-1000 repos are concurrently scanned by 20 searchers
    * 10-990 are ignored

  * **reqB**:

    * 119-1000 repos are concurrently scanned by 20 searchers
    * 19-900 results were ignored


  20 searchers runs concurrently; once some of the searchers has scanned 100th repo it stops, the rest 19 searches finish current job and stop too. Findings from 10 repos are displayed. We skipped 90 empty results. Results from 19 repos were ignored. At least 119 repos were scanned (this number depends on time for each repo and distribution of the repos between searchers).


## Editor Integration

Currently the following editors have plugins that support Hound:

* [Sublime Text](https://github.com/bgreenlee/SublimeHound)
* [Vim](https://github.com/urthbound/hound.vim)
* [Emacs](https://github.com/ryoung786/hound.el)
* [Visual Studio Code](https://github.com/sjzext/vscode-hound)

## Hacking on Hound

### Editing & Building

#### Requirements:
 * make
 * Node.js ([Installation Instructions](https://github.com/joyent/node/wiki/Installing-Node.js-via-package-manager))

Hound includes a `Makefile` to aid in building locally, but it depends on the source being added to a proper Go workspace so that
Go tools work accordingly. See [Setting GOPATH](https://github.com/golang/go/wiki/SettingGOPATH) for further details about setting
up your Go workspace. With a `GOPATH` set, the following commands will build hound locally.

```
git clone https://github.com/itpp-labs/hound.git ${GOPATH}/src/github.com/itpp-labs/hound
cd ${GOPATH}/src/github.com/itpp-labs/hound
make
```

If this is your only Go project, you can set your GOPATH just for Hound:
```
git clone https://github.com/itpp-labs/hound.git src/github.com/itpp-labs/hound
GOPATH=$(pwd) make -C src/github.com/itpp-labs/hound
```

### Testing

There are an increasing number of tests in each of the packages in Hound. Please make sure these pass before uploading your Pull Request. You can run the tests with the following command.
To run the entire test suite, use:

```
make test
```

If you want to just run the JavaScript test suite, use:
```
npm test
```

Any Go files that end in `_test.go` are assumed to be test files.  Similarly, any JavaScript files that ends in `.test.js` are automatically run by Jest, our test runner. Tests should live next to the files that they cover. [Check out Jest's docs](https://jestjs.io/docs/en/getting-started) for more details on writing Jest tests, and [check out Go's testing docs](https://golang.org/pkg/testing/) for more details on testing Go code.

### Working on the web UI

Hound includes a web UI that is composed of several files (html, css, javascript, etc.). To make sure hound works seamlessly with the standard Go tools, these resources are all bundled inside of the `houndd` binary. Note that changes to the UI will result in local changes to the `ui/bindata.go` file. You must include these changes in your Pull Request.

To bundle UI changes in `ui/bindata.go` use:

```
make ui
```

To make development easier, there is a flag that will read the files from the file system (allowing the much-loved edit/refresh cycle).

First you should ensure you have all the dependencies installed that you need by running:

```
make dev
```

Then run the hound server with the --dev option:

```
bin/houndd --dev
```

Note: to make it work, port `9000` should be free.

## Credits

Created at [Etsy](https://www.etsy.com) by:

* [Kelly Norton](https://github.com/kellegous)
* [Jonathan Klein](https://github.com/jklein)

Maintained by [IT Projects Labs](https://itpp.dev/).
