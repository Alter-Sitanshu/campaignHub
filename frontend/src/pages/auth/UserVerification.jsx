import { useEffect, useState } from "react";
import { useParams, Link, useNavigate, useSearchParams } from "react-router-dom";
import "./verification.css";
import { api } from "../../api";

const UserVerification = () => {
    const { token } = useSearchParams();
    const navigate = useNavigate();
    const [status, setStatus] = useState('idle');
    const [message, setMessage] = useState('');

    useEffect(() => {
        if (token === "" || token === null) {
            navigate("/");
        }
        handleVerified(token);

        return () => {
            listener?.subscription?.unsubscribe();
        }
    }, [entity]);

    const handleVerified = async (accessToken) => {
        setStatus('accepting');
        setMessage('Verifying and accepting your account...');
        try {
            const resp = await api.get(`/verify/?token=${accessToken}`);
            if (resp.status === 200) {
                setStatus('success');
                setMessage('Account accepted! Redirecting to your login...');
                
                setTimeout(() => navigate('/auth/sign_in'), 900);
                
            } else {
                setStatus('error');
                setMessage(`Failed to accept account (${resp.status}).`);
            }
        } catch (err) {
            setStatus('error');
            setMessage('Failed to accept account. Try again later.');
        }
    }

    let content;
    switch (status) {
        case 'idle':
            content = <p>Checking verification status...</p>;
            break;
        case 'no_session':
        case 'not_verified':
            content = (
                <div>
                    <p>{message}</p>
                    <p>If you have verified your email, try signing in to refresh your session.</p>
                    <button onClick={() => navigate('/auth/sign_in')}>Sign In</button>
                </div>
            );
            break;
        case 'accepting':
            content = <p>{message}</p>;
            break;
        case 'missing_payload':
        case 'error':
            content = (
                <div>
                    <p>{message}</p>
                    <p>Please retry the signup process.</p>
                    <Link to={entity === 'brand' ? '/auth/brands/sign_up' : '/auth/users/sign_up'}>Sign up</Link>
                </div>
            );
            break;
        case 'success':
            content = (
                <div>
                    <p>{message}</p>
                    <Link to="/auth/sign_in">Proceed to Sign In</Link>
                </div>
            );
            break;
        default:
            content = <p>Unexpected state.</p>;
    }

    return (
        <div className="verification-page">
            <div className="verification-box">
                <h1>Account Verification</h1>
                {content}
            </div>
        </div>
    );
};

export default UserVerification;