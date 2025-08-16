import React from 'react';
import './App.css';
import 'antd/dist/antd.css';
import { HashRouter as Router, Route, Switch, Link } from 'react-router-dom';
import RootLayout from './component/layout/index';
import LiveList from './component/live-list/index';
import LiveInfo from './component/live-info/index';
import ConfigInfo from './component/config-info/index';
import FileList from './component/file-list/index';
import RemoteControl from './component/remote';

const App: React.FC = () => {
  return (
    <Router>
      <RootLayout>
        <Switch>
          <Route path="/fileList/:path(.*)?" component={FileList}></Route>
          <Route path="/configInfo" component={ConfigInfo}></Route>
          <Route path="/liveInfo" component={LiveInfo}></Route>
          <Route path="/remote" component={RemoteControl}></Route>
          <Route path="/" component={LiveList}></Route>
        </Switch>
      </RootLayout>
    </Router>
  );
}

export default App;
