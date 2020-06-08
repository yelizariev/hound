import React, { createRef, useState } from 'react';
import { CoalesceMatches, ContentFor } from '../../utils';
import { Model } from '../../helpers/Model';
import { Match } from './Match';

export class File extends React.Component {
    constructor(props) {
        super(props)
        this.state = {showContent: true}
        this.textArea = createRef(null)
        this.toggleContent = this.toggleContent.bind(this)
        this.copyFilepath = this.copyFilepath.bind(this)
    }
    openOrClose(to_open) {
        this.setState({'showContent': to_open})
    }
    toggleContent() {
        // console.log('toggleContent')
        this.setState({'showContent': !this.state.showContent})
    }
    copyFilepath(evt) {
        evt.preventDefault()
        evt.stopPropagation()
        console.log(evt)
        this.textArea.current.select()
        document.execCommand('copy')
    }

    render (){
        const filename = this.props.match.Filename;
        const blocks = CoalesceMatches(this.props.match.Matches);

        const matches = blocks.map((block, index) => (
            <Match
              key={`match-${this.props.repo}-${index}`}
              block={ block }
              repo={ this.props.repo }
              regexp={ this.props.regexp }
              rev={ this.props.rev }
              filename={ filename }
              />
        ));

    return (
        <div className={"file " + (this.state.showContent ? "open" : "closed")}>
            <div className="title" onClick={ this.toggleContent }>
              <a href={ Model.UrlToRepo(this.props.repo, filename, null, this.props.rev) }>
                    { filename }
                </a>
              <a href="#" className="octicon octicon-clippy copy-file-path" title="Copy to clipboard" onClick={ this.copyFilepath }></a>
            </div>
            <div className="file-body">
                { matches }
            </div>
            <textarea className="copy-file-path-textarea" ref={ this.textArea } defaultValue={ filename }></textarea>
        </div>
    )
    }

};
