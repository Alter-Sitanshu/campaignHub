import { useState } from "react"
import { useAuth } from "../../AuthContext";
import google_icon from "../../assets/google.svg";
import github_icon from "../../assets/github-mark.png";

const AccountBarrier = () => {

    const [ isLoading, setIsLoading ] = useState(false);
    const { supabase } = useAuth()
    const params = new URLSearchParams(location.search);
    const entity = params.get('entity') || 'users';

    const handleAccountVerification = async (provider) => {
        setIsLoading(true);
        const redirectTo = `${window.location.origin}/auth/${encodeURIComponent(entity)}/sign_up`;
        const response = await supabase.auth.signInWithOAuth({
            provider: provider.toLowerCase(),
            options: { redirectTo },
        })
        // supabase may redirect the browser; if it returns a session (popup flow), persist it
        if (response?.data?.session) setSession(response.data.session);
    }

    const handleGoogle = () => {
        handleAccountVerification("google");
    }

    const handleGithub = () => {
        handleAccountVerification("github");
    }

    return (
        <div className='form-page'>
            <div className="auth-page">
                <div className="auth-container">
                    <div className="signup-box">
                        <div className="auth-header">
                            <h1>Choose Provider</h1>
                            <p>
                                To continue to <span>FrogMedia</span>
                            </p>
                        </div>
                        <div className="signin-button-group">
                            <button
                                className='signin-provider'
                                onClick={handleGoogle}
                                type="button"
                                disabled={isLoading}
                            >
                                <img src={google_icon} alt="" className='logo'/>
                                {isLoading ? 'Redirecting...' : 'Continue with Google'}
                            </button>
                            <button
                                className='signin-provider'
                                onClick={handleGithub}
                                type="button"
                                disabled={isLoading}
                            >
                                <img src={github_icon} alt="" className='logo'/>
                                {isLoading ? 'Redirecting...' : 'Continue with GitHub'}
                            </button>
                        </div>
                        <p className="auth-footer">
                            Already have an account? <a href="/auth/sign_in">Sign in</a>
                        </p>
                    </div>
                </div>
            </div>
            
        </div>
    )
}

export default AccountBarrier;