import logo from './logo.svg';
import './App.css';
import { useEffect, useState } from 'react';

function App() {

  const [endpoint, setEndpoint] = useState("");
  const [response, setResponse] = useState("");

  const generate = () => {
    fetch(endpoint).then((res) => {
      return res.text();
    })
    .then((responseJson) => {
      console.log(responseJson);
      setResponse(responseJson);
    })
    .catch((err) => {
      console.log(err.message);
    });
  };

  useEffect(() => {
    generate()
  }, [response])

  return (
    <div className="App">
      <header className="App-header">
        <label>
          Endpoint:
          <input
            value={endpoint}
            onChange={e => setEndpoint(e.target.value)}
            type="text"
          />
          <button onClick={generate}>
            Get Response from Endpoint
          </button>
        </label>
        <img src={logo} className="App-logo" alt="logo" />
        <p>
          Response from calling endpoint
        </p>
        <a
          className="App-link"
          href="https://reactjs.org"
          target="_blank"
          rel="noopener noreferrer"
        >
          {response}
        </a>
      </header>
    </div>
  );
}

export default App;
