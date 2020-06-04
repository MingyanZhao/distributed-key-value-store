import React from 'react';
import './App.css';
import Chatroom from './chatroom.js';

class App extends React.Component {

  constructor(props) {
    super(props);

    const params = new URLSearchParams(window.location.search);
    this.room = params.get('room');
  }

  render() {
    return (
      <div className="App">
        <Chatroom room={this.room}/>
      </div>
    );
  }componentDidMount
}

export default App;
