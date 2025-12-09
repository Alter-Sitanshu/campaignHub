import './App.css';
import {BrowserRouter as Router, Routes, Route} from 'react-router-dom';
import SignUp from './pages/auth/SignUp';
import SignIn from './pages/auth/SignIn';
import Landing from './pages/landing/Landing';
import BrandSignUp from './pages/auth/BrandSignUp';
import BrandDashboard from "./pages/dashboards/BrandDashboard";
import UserDashboard from './pages/dashboards/UserDashboard';
import ErrorPage from './pages/errors/ErrorPage';

import { AuthProvider } from './AuthContext';



function App() {

  return (
    <AuthProvider>
      <Router>
        <Routes>
          {/* Landing page */}
          <Route path='/' element={<Landing />}></Route>
          
          {/* auth pages */}
          <Route path="/auth/sign_in" element={<SignIn />}></Route>
          <Route path="/auth/users/sign_up" element={<SignUp />}></Route>
          <Route path="/auth/brands/sign_up" element={<BrandSignUp />}></Route>

          {/* Protected routes */}
          <Route path="/brands/dashboard/:id" element={<BrandDashboard />}></Route>
          <Route path="/users/dashboard/:id" element={<UserDashboard />}></Route>

          <Route path="/errors/:code" element={<ErrorPage />}></Route>
        </Routes>
      </Router>
    </AuthProvider>
  )
}

export default App
