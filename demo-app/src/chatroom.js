import styles from './chatroom.module.css';
import React from 'react';
const {FollowerClient} = require('./follower/follower_grpc_web_pb.js')
const {GetRequest, PutRequest} = require('./follower/follower_pb.js')


class Chatroom extends React.Component {

  messagesEndRef = React.createRef()

  constructor(props) {
    super(props);
    this.state = {
        messages: [],
        newValue: '',
    };

    this.handleInputChange = this.handleInputChange.bind(this);
    this.handleSubmit = this.handleSubmit.bind(this);
  }

  componentDidMount() {    
    this.client = new FollowerClient("http://" + process.env.REACT_APP_IP + ":9090", null, null);
    this.refresh();
    this.refreshTimer = setInterval(() => this.refresh(), 5000);
  }

  componentWillUnmount() {
    clearInterval(this.refreshTimer)
  }

  refresh() {
    const request = new GetRequest();
    request.setKey(this.props.room);

    this.client.get(request, {}, (err, response) => {
      if (err) {
        console.log(err);
        return;
      }

      this.setState({messages: response.getValuesList()});
    });
  }

  handleInputChange(event) {
    this.setState({newValue: event.target.value});  
  }

  handleSubmit(event) {
    console.log("handSubmit");
    if (this.state.newValue === '') {
      return
    }
    const newValue = this.state.newValue;
    this.setState({newValue: ''});

    const request = new PutRequest();
    request.setKey(this.props.room);
    request.setValue(newValue);    

    this.client.put(request, {}, (err, response) => {});
    
    
    event.preventDefault();
  }

  scrollToBottom = () => {
    this.messagesEnd.current.scrollIntoView({ behavior: 'smooth' })
  }

  render() {
    var messages = this.state.messages.map((msg) => {
      return <li className={styles.messageItem}>{msg}</li>;
    })
    return(
      <div className={styles.main}>
        <div className={styles.heading}>
          {process.env.REACT_APP_REGION}
        </div>
        <div ref={this.messagesEndRef} className={styles.chat}>
          <ul>
            {messages} 
          </ul>
          
        </div>
        <div className={styles.control}>
          <form onSubmit={this.handleSubmit}>
            <input className={styles.input} type="text" value={this.state.newValue} onChange={this.handleInputChange}></input>
            <input className={styles.send} type="submit" value="Send" />
          </form>
        </div>
      </div>
    )
  }
}

export default Chatroom