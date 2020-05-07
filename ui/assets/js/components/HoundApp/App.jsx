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
    const [ allRepos, setAllRepos ] = useState([]);
    const [ stats, setStats ] = useState('');
    const [ results, setResults ] = useState(null);
    const [ error, setError ] = useState(null);

    useEffect(() => {

        const config = Model.config;
        const InitSearch = config.InitSearch || {};
        const initParams = {
            q: InitSearch.q || '',
            i: InitSearch.i || 'nope',
            files: InitSearch.files || '',
            excludeFiles: InitSearch.excludeFiles || '',
            repos: InitSearch.repos || '*'
        }
        const urlParams = ParamsFromUrl(initParams);
        setQuery(urlParams.q);
        setIgnoreCase(urlParams.i);
        setFiles(urlParams.files);
        setExcludeFiles(urlParams.excludeFiles);
        setRepos(
            (urlParams.repos === '') ? [] : urlParams.repos.split(',')
        );

        Model.didLoadRepos.tap((model, allRepos) => {
            // If all repos are selected, don't show any selected.
            if (model.ValidRepos(repos).length === model.RepoCount()) {
                setRepos([]);
            }
            setAllRepos(Object.keys(allRepos));
        });

        Model.didSearch.tap((model, results, stats) => {
            setStats(stats);
            setResults(results);
            setError(null);
        });

        Model.didLoadMore.tap((model, repo, results) => {
            setResults([...results]);
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
            setRepos( (urlParams.repos === '') ? [] : urlParams.repos.split(',') );
            Model.Search(urlParams);
        });

    }, []);

    const updateHistory = (params) => {
        const path = `${ location.pathname }`
            + `?q=${ encodeURIComponent(params.q) }`
            + `&i=${ encodeURIComponent(params.i) }`
            + `&files=${ encodeURIComponent(params.files) }`
            + `&excludeFiles=${ encodeURIComponent(params.excludeFiles) }`
            + `&repos=${ params.repos }`;
        history.pushState({ path: path }, '', path);
    };

    const onSearchRequested = (params) => {
        updateHistory(params);
        if ( params.q !== query ) { setQuery(params.q); }
        if ( params.i !== ignoreCase ) { setIgnoreCase(params.i) }
        if ( params.files !== files ) { setFiles(params.files) }
        if ( params.excludeFiles !== excludeFiles ) { setExcludeFiles(params.excludeFiles) }
        setResults(null);
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
                allRepos={ allRepos }
                stats={ stats }
                onSearchRequested={ onSearchRequested }
            />
            <ResultView
                query={ query }
                ignoreCase={ ignoreCase }
                results={ results }
                error={ error }
            />
            <SelectionTooltip delay={ 50 }/>
        </Fragment>
    );
};
