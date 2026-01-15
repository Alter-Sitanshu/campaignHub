import { createContext, useContext, useState, useEffect } from 'react';
import { createClient } from '@supabase/supabase-js';

const AuthContext = createContext();

let logout = null;

export const getLogoutHandler = () => logout;

const setLogoutHandler = (fn) => {
  logout = fn;
}

export const AuthProvider = ({ children }) => {
  const supabase = createClient(
    import.meta.env.VITE_SUPABASE_URL, 
    import.meta.env.VITE_SUPABASE_PUBLISHABLE_DEFAULT_KEY
  )
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [session, setSession] = useState(null);

  useEffect(() => {
    // On app load, check if user is stored in localStorage
    setLogoutHandler(logout);
    // Check for existing session
    supabase.auth.getSession().then(({ data: { session } }) => {
      setSession(session);
    });
    const storedUser = localStorage.getItem('user');
    
    if (storedUser) {
      setUser(JSON.parse(storedUser));
    }
    setLoading(false);
  }, []);

  const login = (userData) => {
    setUser(userData);
    localStorage.setItem('user', JSON.stringify(userData));
  };

  const loginWith = async (provider, entity = 'users') => {
    // Include entity in redirect so the callback knows which type of account to create/lookup
    const redirectTo = `${window.location.origin}/auth/verify?entity=${encodeURIComponent(entity)}`;
    const response = await supabase.auth.signInWithOAuth({
      provider: provider.toLowerCase(),
      options: { redirectTo },
    })
    // supabase may redirect the browser; if it returns a session (popup flow), persist it
    if (response?.data?.session) setSession(response.data.session);
  }

  const logout = async () => {
    setUser(null);
    localStorage.removeItem('user');
    await supabase.auth.signOut();
    await cookieStore.delete({
      name: 'session',
      domain: 'frogmedia.onrender.com',
      // domain: 'localhost:8080',
      path: '/'
    });
    setSession(null);
  };

  return (
    <AuthContext.Provider value={{ user, loading, login, logout, session, loginWith, supabase }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
};