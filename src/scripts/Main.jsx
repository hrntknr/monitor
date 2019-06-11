import React from 'react';
import {Router, Switch, Route} from 'react-router-dom';
import container from './container';
import {createBrowserHistory} from 'history';
import Index from './pages/Index.jsx';
import Debug from './pages/Debug.jsx';

class Main extends React.Component {
  constructor(props) {
    super(props);
    const history = createBrowserHistory();
    this.history = history;
  }
  render() {
    return (
      <Router history={this.history}>
        <Switch>
          <Route path='/' component={Index} exact />
          <Route path='/debug' component={Debug} exact />
        </Switch>
      </Router>
    );
  }
}

export default container(Main);
