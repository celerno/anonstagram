import React, { Component } from 'react';
import Header from './components/Header';
import { ApolloProvider } from "react-apollo";
import ApolloClient from 'apollo-boost'; 
import './App.css';
import Posts from './components/Posts';
  const client = new ApolloClient({
    uri: "http://localhost:4000/graphql"
  });

class App extends Component {
  render() {
    return <div className="App">
  
    <ApolloProvider client={client}>
          <div className="App">
            <Header />
            <section className="App-main">
              <Posts />
            </section>
          </div>
    </ApolloProvider>
  </div>;
  }
}
export default App;