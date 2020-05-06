import React from 'react';
import { Model } from '../../helpers/Model';
import { FilesView } from './FilesView';

export class Repo extends React.Component {
    constructor(props) {
        super(props)
        this.state = {repoOpen: true}
        this.filesCollection = {}
        this.toggleFiles = this.toggleFiles.bind(this)
    }

    openOrCloseFiles(to_open) {
        this.setState({
            repoOpen: to_open
        })
        for (let index in this.filesCollection) {
            let f = this.filesCollection[index]
            f.openOrClose(to_open)
        }
    }
    toggleFiles() {
        this.openOrCloseFiles(!this.state.repoOpen)
    }
    openFiles () {
        this.openOrCloseFiles(true)
    }
    closeFiles () {
        this.openOrCloseFiles(false)
    }

    render () {
        return (
        <div className={"repo" + (this.state.repoOpen ? "open" : "closed")}>
          <div className="title" onClick={ this.toggleFiles }>
            <span className="mega-octicon octicon-repo"></span>
            <span className="name"> { Model.NameForRepo(this.props.repo) } </span>
            <span className={"indicator octicon octicon-chevron-"+ (this.state.repoOpen ? "down" : "right" )} onClick={ this.toggleFiles }></span>
          </div>
          <FilesView
            filesCollection={ this.filesCollection }
            matches={ this.props.matches }
            rev={ this.props.rev }
            repo={ this.props.repo }
            regexp={ this.props.regexp }
            totalMatches={ this.props.files }
            filesShowState={ this.filesShowState }
            />
        </div>
        )
    }
}


/*

export const Repo = (props) => {
    const { ref, repo, rev, matches, regexp, files, stateShow } = props;
    const [ status, setStatus] = stateShow

    const openOrCloseAll = (to_open) => {
        setStatus(to_open)
        for (let index in filesShowState) {
            let [state, setState] = reposShowState[index]
            if (to_open) {
                setState(true)
            } else {
                setState(false)
            }
        }
    }

    const toggleStatus = () => {
        openOrCloseAll(!status)
    }

    const filesShowState = {}

    return (
        <div className={"repo" + (status ? "open" : "closed")}>
          <div className="title" onClick={ toggleStatus }>
            <span className="mega-octicon octicon-repo"></span>
            <span className="name">{ Model.NameForRepo(repo) }</span>
            <span className={"indicator octicon octicon-chevron-"+ (status ? "up" : "down" )} onClick={ toggleStatus }></span>
          </div>
          <FilesView
            matches={ matches }
            rev={ rev }
            repo={ repo }
            regexp={ regexp }
            totalMatches={ files }
            filesShowState=filesShowState
            />
        </div>
    )
}

import React from 'react';
import { Model } from '../../helpers/Model';
import { FilesView } from './FilesView';

class Repo extends React.Component {
    constructor(props) {
        super(props);
        this.state = {visible: true};
    }

//    const { ref, repo, rev, matches, regexp, files } = props;
//    const [ showContent, setShowContent] = useState(true);

    toggleContent() {
        this.setState({
            visible: !this.state.visible
        })
    }
    open () {
        this.setState({
            visible: true
        })
    }
    close () {
        this.setState({
            visible: false
        })
    }

    render () {
        return (
        <div className={"repo" + (this.state.visible ? "open" : "closed")}>
          <div className="title" onClick={ this.toggleContent }>
            <span className="mega-octicon octicon-repo"></span>
            <span className="name">{ Model.NameForRepo(repo) }</span>
            <span className={"indicator octicon octicon-chevron-"+ (this.state.visible ? "up" : "down" )} onClick={ this.toggleContent }></span>
          </div>
          <FilesView
            matches={ this.props.matches }
            rev={ this.props.rev }
            repo={ this.props.repo }
            regexp={ this.props.regexp }
            totalMatches={ this.props.files }
            />
        </div>
        )
    }
}


  */
