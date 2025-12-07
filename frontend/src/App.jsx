import './App.css';
import {BrowserRouter as Router, Routes, Route} from 'react-router-dom';
import SignUp from './pages/auth/SignUp';
import SignIn from './pages/auth/SignIn';
import Landing from './pages/landing/Landing';
import BrandSignUp from './pages/auth/BrandSignUp';
import BrandDashboard from "./pages/dashboards/BrandDashboard";
import UserDashboard from './pages/dashboards/UserDashboard';

function App() {

  return (
    <Router>
      <Routes>
        {/* Landing page */}
        <Route path='/' element={<Landing />}></Route>
        
        {/* auth pages */}
        <Route path="/auth/users/sign_in" element={<SignIn entity="users"/>}></Route>
        <Route path="/auth/brands/sign_in" element={<SignIn entity="brands"/>}></Route>
        <Route path="/auth/users/sign_up" element={<SignUp />}></Route>
        <Route path="/auth/brands/sign_up" element={<BrandSignUp />}></Route>

        {/* Protected routes */}
        <Route path="/brands/dashboard/" element={<BrandDashboard />}></Route>
        <Route path="/users/dashboard/" element={<UserDashboard />}></Route>
      </Routes>
    </Router>
  )
}

export default App
