import { useEffect, useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { oauthCallbackSignIn } from '../../api';
import { useAuth } from '../../AuthContext';

export default function OAuthCallback() {
  const navigate = useNavigate();
  const location = useLocation();
  const { login, supabase } = useAuth();
  const [message, setMessage] = useState('Processing sign in...');

  useEffect(() => {
    (async () => {
      try {
        const params = new URLSearchParams(location.search);
        const entity = params.get('entity') || 'users';

        // Try to read the session. After OAuth redirect supabase should have set session.
        const { data: { session } } = await supabase.auth.getSession();
        const user = session?.user;

        if (!user || !user.email) {
          setMessage('No session found. Please sign in again.');
          setTimeout(() => navigate('/auth/sign_in'), 1200);
          return;
        }

        const email = user.email;
        const fullName = user.user_metadata?.full_name || user.user_metadata?.name || '';
        const parts = fullName.trim().split(' ');
        const first_name = parts.shift() || '';
        const last_name = parts.join(' ') || '';

        setMessage('Finalizing sign in...');

        const resp = await oauthCallbackSignIn({
          email,
          first_name,
          last_name,
          entity,
        });

        if (resp.type === 'error') {
          setMessage('Sign in failed. Please try again.');
          setTimeout(() => navigate('/auth/sign_in'), 1200);
          return;
        }

        // Persist the user locally and navigate to their dashboard
        login({ id: resp.id, username: resp.username, email: resp.email, entity });
        navigate(`/${entity}/dashboard/${resp.id}`);
      } catch (err) {
        console.error(err);
        setMessage('An error occurred during sign in.');
        setTimeout(() => navigate('/auth/sign_in'), 1500);
      }
    })();
  }, [location, navigate, login]);

  return (
    <div className="form-page">
      <div style={{ textAlign: 'center', paddingTop: '6rem' }}>
        <p>{message}</p>
      </div>
    </div>
  );
}
