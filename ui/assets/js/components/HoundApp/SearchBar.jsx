import React, { useEffect, useState, useRef } from 'react';
import { FormatNumber, ParamValueToBool } from '../../utils';
import { Model } from '../../helpers/Model';
import Select from 'react-select';

export const SearchBar = (props) => {

    const { query, ignoreCase, files, excludeFiles, repos, allRepos, reposRE, stats, onSearchRequested } = props;
    const [ showAdvanced, setShowAdvanced] = useState(false);
    const [ searchQuery, setSearchQuery ] = useState(query);
    const [ searchIgnoreCase, setSearchIgnoreCase ] = useState(ignoreCase);
    const [ searchFiles, setSearchFiles ] = useState(files);
    const [ searchExcludeFiles, setSearchExcludeFiles ] = useState(excludeFiles);
    const [ searchReposRE, setSearchReposRE ] = useState(reposRE);
    const [ searchRepos, setSearchRepos ] = useState(repos);
    const queryInput = useRef(null);
    const fileInput = useRef(null);
    //const excludeFileInput = useRef(null);

    const hasAdvancedValues = () => (
        ( searchFiles && searchFiles.trim() !== '' ) ||
        ( searchExcludeFiles && searchExcludeFiles.trim() !== '' ) ||
        ( searchIgnoreCase && searchIgnoreCase.trim() === 'fosho' ) ||
        ( searchReposRE && searchReposRE.trim() !== '' )
    );

    useEffect(() => { setSearchQuery(query) }, [query]);
    useEffect(() => { setSearchIgnoreCase(ignoreCase) }, [ignoreCase]);
    useEffect(() => { setSearchFiles(files) }, [files]);
    useEffect(() => { setSearchExcludeFiles(excludeFiles) }, [excludeFiles]);
    useEffect(() => { setSearchReposRE(reposRE); }, [reposRE]);
    useEffect(() => { setSearchRepos(repos) }, [repos]);

    const repoOptions = allRepos.map(rname => ({
        value: rname,
        label: rname
    }));

    const selectedRepos = repoOptions.filter(o => searchRepos.indexOf(o.value) >= 0);

    const showAdvancedCallback = () => {
        setShowAdvanced(true);
        if (searchQuery.trim() !== '') {
            fileInput.current.focus();
        }
    };

    const hideAdvancedCallback = () => {
        setShowAdvanced(false);
        if (queryInput.current) {
            queryInput.current.focus();
        }
    };

    const elementChanged = (prop, evt) => {
        console.log('elementChanged', prop, evt.currentTarget.value);
        switch (prop) {
            case 'query':
                setSearchQuery(evt.currentTarget.value);
                break;
            case 'files':
                setSearchFiles(evt.currentTarget.value);
                break;
            case 'excludeFiles':
                setSearchExcludeFiles(evt.currentTarget.value);
                break;
            case 'reposRE':
                setSearchReposRE(evt.currentTarget.value);
                // repos RE is changed manually, so remove selection
                setSearchRepos([]);
                break;
            case 'ignoreCase':
                setSearchIgnoreCase(evt.currentTarget.checked && 'fosho' || 'nope');
                break;
        }
    };

    const submitQuery = () => {
        if (searchQuery.trim() !== '') {
            onSearchRequested({
                q: searchQuery,
                i: searchIgnoreCase,
                files: searchFiles,
                excludeFiles: searchExcludeFiles,
                repos: searchReposRE
            });
        }
    };

    const queryGotKeydown = (event) => {
        switch (event.keyCode) {
            case 40:
                showAdvancedCallback();
                fileInput.current.focus();
                break;
            case 38:
                hideAdvancedCallback();
                break;
            case 13:
                submitQuery();
                break;
        }
    };

    const queryGotFocus = () => {
        if ( !hasAdvancedValues() ) {
            hideAdvancedCallback();
        }
    };

    const filesGotKeydown = (event) => {
        switch (event.keyCode) {
            case 38:
                // if advanced is empty, close it up.
                if (searchFiles.trim() === '') {
                    hideAdvancedCallback();
                }
                queryInput.current.focus();
                break;
            case 13:
                submitQuery();
                break;
        }
    };

    const excludeFilesGotKeydown = (event) => {
        switch (event.keyCode) {
            case 38:
                // if advanced is empty, close it up.
                if (searchExcludeFiles.trim() === '') {
                    hideAdvancedCallback();
                }
                queryInput.current.focus();
                break;
            case 13:
                submitQuery();
                break;
        }
    };

    const reposREGotKeydown = (event) => {
        switch (event.keyCode) {
        case 38:
            // if advanced is empty, close it up.
            if (searchReposRE.trim() === '') {
                hideAdvancedCallback();
            }
            queryInput.current.focus();
            break;
        case 13:
            submitQuery();
            break;
        }
    };

    const repoSelected = (selected) => {
        const repos = selected
              ? selected.map(item => item.value)
              : []
        setSearchRepos(repos);
        setSearchReposRE(repos ? repos.map(r => "^" + r + "$").join("|") : ".*")
    };

    const statsView = stats
        ? (
            <div className="stats">
                <div className="stats-left">
                    <a href="excluded_files.html"
                       className="link-gray">
                        Excluded Files
                    </a>
                </div>
                <div className="stats-right">
                    <div className="val">{ FormatNumber(stats.Total) }ms total</div> /
                    <div className="val">{ FormatNumber(stats.Server) }ms server</div> /
                    <div className="val">{ stats.Files } files</div>
                    <div className="val">{ stats.Repos } repos</div>
                </div>
            </div>
        )
        : '';

    return (
        <div id="input">
            <div id="ina">
                <input
                    ref={ queryInput }
                    type="text"
                    placeholder="Search by Regexp"
                    autoComplete="off"
                    autoFocus
                    value={ searchQuery }
                    onFocus={ queryGotFocus }
                    onChange={ elementChanged.bind(this, "query") }
                    onKeyDown={ queryGotKeydown }
                />
                <div className="button-add-on">
                    <button id="dodat" onClick={ submitQuery }></button>
                </div>
            </div>

            <div id="inb" className={ showAdvanced ? 'opened' : 'closed' }>
                <div id="adv">
                    <span className="octicon octicon-chevron-up hide-adv" onClick={ hideAdvancedCallback }></span>
                    <div className="field">
                        <label htmlFor="files">File Path</label>
                        <div className="field-input">
                            <input
                                ref={ fileInput }
                                type="text"
                                placeholder="regexp"
                                value={ searchFiles }
                                onChange={ elementChanged.bind(this, "files") }
                                onKeyDown={ filesGotKeydown }
                            />
                        </div>
                    </div>
                    <div className="field">
                        <label htmlFor="excludeFiles">Exclude File Path</label>
                        <div className="field-input">
                            <input
                                type="text"
                                placeholder="regexp"
                                value={ searchExcludeFiles }
                                onChange={ elementChanged.bind(this, "excludeFiles") }
                                onKeyDown={ excludeFilesGotKeydown }
                            />
                        </div>
                    </div>
                    <div className="field">
                        <label htmlFor="ignore-case">Ignore Case</label>
                        <div className="field-input">
                            <input type="checkbox" onChange={ elementChanged.bind(this, "ignoreCase") } checked={ ParamValueToBool(searchIgnoreCase) } />
                        </div>
                    </div>
                    <div className="field">
                      <label htmlFor="reposRE">Repos</label>
                      <div className="field-input">
                        <input
                          type="text"
                          placeholder="regexp"
                          value={ searchReposRE }
                          onChange={ elementChanged.bind(this, "reposRE") }
                          onKeyDown={ reposREGotKeydown }
                          />
                      </div>
                    </div>
                    <div className="field-repo-select">
                        <label className="multiselect_label" htmlFor="repos">Select Repo</label>
                        <div className="field-input">
                            <Select
                                options={ repoOptions }
                                onChange={ repoSelected }
                                value={ selectedRepos }
                                isMulti
                                closeMenuOnSelect={ false }
                            />
                        </div>
                    </div>
                </div>
                <div className="ban" onClick={ showAdvancedCallback }>
                    <em>Advanced:</em> ignore case, filter by path, stuff like that.
                </div>
            </div>
            { statsView }
        </div>
    );
};
