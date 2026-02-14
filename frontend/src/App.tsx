import "./App.css";

const version = import.meta.env.VITE_APP_VERSION || "dev";

function App() {
  return (
    <div className="app">
      <h1>Hatch</h1>
      <p className="version">v{version}</p>
      <p className="subtitle">Local HTTPS reverse proxy for macOS development</p>
    </div>
  );
}

export default App;
