import { useEffect, useState } from "react";
import { Call } from "@wailsio/runtime";
import "./App.css";

function App() {
  const [version, setVersion] = useState("loading...");

  useEffect(() => {
    Call.ByName("github.com/paulrose/hatch/internal/app.App.GetVersion")
      .then((result) => setVersion(String(result)))
      .catch((err) => {
        console.warn("GetVersion binding call failed:", err);
        setVersion("dev (no wails runtime)");
      });
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
