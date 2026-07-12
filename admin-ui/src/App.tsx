import { Routes, Route, NavLink } from 'react-router-dom';
import { Activity, GitMerge, Server, ShieldAlert, Book } from 'lucide-react';
import Dashboard from './pages/Dashboard';
import GatewayRoutes from './pages/Routes';
import Upstreams from './pages/Upstreams';
import CircuitBreakers from './pages/CircuitBreakers';
import Docs from './pages/Docs';

function App() {
  return (
    <div className="app-container">
      <nav className="sidebar">
        <h1>Gateway Admin</h1>
        <div className="nav-links">
          <NavLink to="/" end className={({ isActive }) => isActive ? 'active' : ''}>
            <Activity size={18} /> Dashboard
          </NavLink>
          <NavLink to="/routes" className={({ isActive }) => isActive ? 'active' : ''}>
            <GitMerge size={18} /> Routes
          </NavLink>
          <NavLink to="/upstreams" className={({ isActive }) => isActive ? 'active' : ''}>
            <Server size={18} /> Upstreams
          </NavLink>
          <NavLink to="/circuit-breakers" className={({ isActive }) => isActive ? 'active' : ''}>
            <ShieldAlert size={18} /> Circuit Breakers
          </NavLink>
          <NavLink to="/docs" className={({ isActive }) => isActive ? 'active' : ''}>
            <Book size={18} /> Documentation
          </NavLink>
        </div>
      </nav>
      
      <main className="main-content">
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/routes" element={<GatewayRoutes />} />
          <Route path="/upstreams" element={<Upstreams />} />
          <Route path="/circuit-breakers" element={<CircuitBreakers />} />
          <Route path="/docs" element={<Docs />} />
        </Routes>
      </main>
    </div>
  );
}

export default App;
