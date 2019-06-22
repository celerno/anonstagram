import React, { Component } from 'react';
import Header from './components/Header';
import Post from './components/Post'
import { ApolloProvider } from "react-apollo";
import ApolloClient from 'apollo-boost'; 
import './App.css';
  const client = new ApolloClient({
    uri: "http://localhost:4000/graphql"
  });

class App extends Component {
  render() {
    return <div className="App">
    <Header />
    <section className="App-main">
      <Post nickname="Chris" avatar="https://www.laravelnigeria.com/img/chris.jpg" caption="Moving the community!" image="https://pbs.twimg.com/media/DOXI0IEXkAAkokm.jpg" />
      <Post nickname="OG" avatar="https://www.laravelnigeria.com/img/chris.jpg" caption="Holding a mic" image="https://pbs.twimg.com/media/DOXI0IEXkAAkokm.jpg" />
      {/* more posts */}
    </section>
    <ApolloProvider client={client}>
          <div className="App">
            <Header />
            <section className="App-main">
              <Post />
            </section>
          </div>
    </ApolloProvider>
  </div>;
  }
}
export default App;