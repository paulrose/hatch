interface Window {
  go?: {
    app?: {
      App?: {
        GetVersion(): Promise<string>;
      };
    };
  };
}
