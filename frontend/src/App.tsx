import { useEffect, useState } from "react";
import "./App.css";

function App() {
  const [version, setVersion] = useState("loading...");

  useEffect(() => {
    if (window.go?.app?.App?.GetVersion) {
      window.go.app.App.GetVersion().then(setVersion);
    } else {
      setVersion("dev (no wails runtime)");
    }
  }, []);

  return (
    <div className="app">
      <h1>Hatch</h1>
      <p className="version">v{version}</p>
      <p className="subtitle">Local HTTPS reverse proxy for macOS development</p>
    </div>
  );
}

export default App;
