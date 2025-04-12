import { useEffect } from "react";

function App() {
  useEffect(() => {
    fetch(`${import.meta.env.VITE_API_URL}/user/profile`, {
      credentials: "include",
    })
      .then((res) => res.json())
      .then((data) => console.log(data));
  }, []);

  return (
    <div>
      <h1>Resume Generator</h1>
    </div>
  );
}

export default App;
