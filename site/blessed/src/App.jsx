import { BrowserRouter as Router } from "react-router-dom";
import { AppRouter } from "./router/AppRouter";
import './App.scss';
export default function App() {
  return (
    <div className='app'>
      <Router>
        <AppRouter />
      </Router>
    </div>
  );
}
