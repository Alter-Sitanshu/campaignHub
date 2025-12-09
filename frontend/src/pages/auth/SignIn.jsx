import { useState, useEffect } from 'react';
import './signup.css';
import { useNavigate } from 'react-router-dom';
import { signin } from '../../api';
import { useAuth } from "../../AuthContext";

// entity is either users/brands
const SignIn = () => {
    // will use this to navigate to the creator page
    const { user, login } = useAuth();
    const navigate = useNavigate();
    const [signupURL, setSignupURL ] = useState("/auth/users/sign_up");

    const [isValid, setIsValid] = useState(false);
    const [ isLoading, setIsLoading ] = useState(false);
    const [errors, setErrors] = useState({});

    const [formData, setFormData] = useState({
        email: '',
        password: '',
        entity: 'users',
    });

    useEffect(() => {

        if(user !== null) {
            navigate(`/${user.entity}/dashboard/${user.id}`);
        }

        const currentErrors = getErrors();
        setErrors(currentErrors);
        setIsValid(Object.keys(currentErrors).length === 0);
    }, [formData, user]);

    const getErrors = () => {
        const newErrors = {};
        const { email, password } = formData;
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(email)) newErrors.email = 'Enter a valid email';
        if (password.length < 6) newErrors.password = 'Password required';
        return newErrors;
    }

    const handleChange = (e) => {
        if(e.target.name === "entity") {
            // entity is a checkbox
            let newEntity = e.target.checked ? "brands": "users";
            setSignupURL(`/auth/${newEntity}/sign_up`);
            setFormData({...formData, [e.target.name]: newEntity});
        } else {
            setFormData({ ...formData, [e.target.name]: e.target.value });
        }
    };

    const handleSubmit = async () => {
        setIsLoading(true);
        const payload = {
            email: formData.email,
            password: formData.password,
            entity: formData.entity,
        }
        let resp = await signin(payload);
        if (resp.type == "error") {
            navigate(`/errors/${resp.status}`);
        } else{
            login(resp);
            navigate(`/${formData.entity}/dashboard/${resp.id}`);
        }
    };

    return (
        <div className='form-page'>
                
            <div className="auth-page">
                <div className="auth-container">
                    <div className="signup-box">
                        <div className="auth-header">
                            <h1>Account Login</h1>
                            <p>
                                Welcome back to <span>FrogMedia</span>
                            </p>
                        </div>
                        <form action="" className='form'>
                            <div className="form-group">
                                <label htmlFor="email">Email</label>
                                <input
                                    type="email"
                                    id="email"
                                    name="email"
                                    placeholder="john@example.com"
                                    value={formData.email}
                                    onChange={handleChange}
                                    spellCheck={false}
                                />
                                {errors.email && <small className="error-text">{errors.email}</small>}
                            </div>
                            <div className="form-group">
                                <label htmlFor="password">Password</label>
                                <input
                                    type="password"
                                    id="password"
                                    name="password"
                                    placeholder="••••••••"
                                    value={formData.password}
                                    onChange={handleChange}
                                />
                            </div>
                            <div className="form-group checkbox-wrapper">
                                <label htmlFor="entity" className="checkbox-label">
                                    <input 
                                    type="checkbox" 
                                    className="checkbox-input"
                                    name="entity"
                                    onChange={handleChange}
                                    />
                                    <span className="checkbox-custom"></span>
                                    Are you a brand?
                                </label>
                            </div>
                            <div className="button-group">
                                <button 
                                    onClick={handleSubmit}
                                    id={!isValid ? 'submit-disabled' : 'submit'}
                                    type="button"
                                    disabled={!isValid || isLoading}
                                    >
                                    {isLoading ? "Signing In..." : "Sign In"}
                                </button>
                            </div>
                        </form>
                        <p className="auth-footer">
                            Don't have an account? <a href={signupURL}>Create One</a>
                        </p>
                    </div>
                </div>
            </div>
            
        </div>
    )
}


export default SignIn;