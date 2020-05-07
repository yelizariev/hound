import React from 'react';
import ReactDOM from 'react-dom';
import { App } from './components/HoundApp/App';
import { Model } from './helpers/Model';

Model.LoadConfig();

ReactDOM.render(
  <App />,
  document.getElementById('root')
);

Model.Load();
