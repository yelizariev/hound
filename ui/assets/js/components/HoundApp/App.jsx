import React, { Fragment, useState, useEffect } from 'react';
import { ParamsFromUrl } from '../../utils';
import { Model } from '../../helpers/Model';
import { SearchBar } from './SearchBar';
import { ResultView } from './ResultView';
import { SelectionTooltip } from "./SelectionTooltip";

export const App = function (props) {

    const [ query, setQuery ] = useState('');
    const [ ignoreCase, setIgnoreCase ] = useState('nope');
    const [ files, setFiles ] = useState('');
    const [ excludeFiles, setExcludeFiles ] = useState('');
    const [ repos, setRepos ] = useState([]);
    const [ reposRE, setReposRE ] = useState('');
    const [ allRepos, setAllRepos ] = useState([]);
    const [ stats, setStats ] = useState('');
    const [ reposPagination, setReposPagination ] = useState(null);
    const [ results, setResults ] = useState(null);
    const [ error, setError ] = useState(null);

    useEffect(() => {

        const config = Model.config;
        const InitSearch = config.InitSearch || {};
        // This must be the same as on server side (see config.go > ToOpenSearchParams)
        const initParams = {
            q: InitSearch.q || '',
            i: InitSearch.i || 'nope',
            files: InitSearch.files || '',
            excludeFiles: InitSearch.excludeFiles || '',
            repos: InitSearch.repos || '.*'
        }
        const urlParams = ParamsFromUrl(initParams);
        setQuery(urlParams.q);
        setIgnoreCase(urlParams.i);
        setFiles(urlParams.files);
        setExcludeFiles(urlParams.excludeFiles);
        setReposRE(urlParams.repos);

        Model.didLoadRepos.tap((model, allRepos) => {
            // If all repos are selected, don't show any selected.
            if (model.ValidRepos(repos).length === model.RepoCount()) {
                setRepos([]);
            }
            setAllRepos(Object.keys(allRepos));
        });

        Model.didSearch.tap((model, results, stats, reposPagination) => {
            setStats(stats);
            setResults(results);
            setReposPagination(reposPagination);
            setError(null);
        });

        Model.didLoadMore.tap((model, repo, results) => {
            setResults([...results]);
            setError(null);
        });

        Model.didLoadOtherRepos.tap((model, results, reposPagination) => {
            setResults(results);
            setReposPagination(reposPagination);
            setError(null);
        });

        Model.didError.tap((model, error) => {
            setResults(null);
            setError(error);
        });

        window.addEventListener('popstate', (e) => {
            const urlParams = ParamsFromUrl();
            if ( urlParams.q !== query ) { setQuery(urlParams.q); }
            if ( urlParams.i !== ignoreCase ) { setIgnoreCase(urlParams.i) }
            if ( urlParams.files !== files ) { setFiles(urlParams.files) }
            if ( urlParams.excludeFiles !== excludeFiles ) { setExcludeFiles(urlParams.excludeFiles) }
            if ( urlParams.repos !== reposRE ) {
                setReposRE(urlParams.repos)
                setRepos([])
            }
            Model.Search(urlParams);
        });

    }, []);

    const updateHistory = (params) => {
        const path = `${ location.pathname }`
            + `?q=${ encodeURIComponent(params.q) }`
            + `&i=${ encodeURIComponent(params.i) }`
            + `&files=${ encodeURIComponent(params.files) }`
            + `&excludeFiles=${ encodeURIComponent(params.excludeFiles) }`
            + `&repos=${ encodeURIComponent(params.repos) }`;
        history.pushState({ path: path }, '', path);
    };

    const onSearchRequested = (params) => {
        updateHistory(params);
        if ( params.q !== query ) { setQuery(params.q); }
        if ( params.i !== ignoreCase ) { setIgnoreCase(params.i) }
        if ( params.files !== files ) { setFiles(params.files) }
        if ( params.excludeFiles !== excludeFiles ) { setExcludeFiles(params.excludeFiles) }
        if ( params.repos !== reposRE ) { setReposRE(params.repos) }
        setResults(null);
        setReposPagination(null);
        Model.Search(params);
    };

    return (
        <Fragment>
            <SearchBar
                query={ query }
                ignoreCase={ ignoreCase }
                files={ files }
                excludeFiles={ excludeFiles }
                repos={ repos }
                reposRE={ reposRE }
                allRepos={ allRepos }
                stats={ stats }
                onSearchRequested={ onSearchRequested }
            />
            <ResultView
                query={ query }
                ignoreCase={ ignoreCase }
                results={ results }
                reposPagination={ reposPagination }
                error={ error }
            />
            <SelectionTooltip delay={ 50 }/>
        </Fragment>
    );
};
