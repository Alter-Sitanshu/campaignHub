import { useState, useEffect } from 'react';
import './signup.css';
import { useNavigate } from 'react-router-dom';

// entity is either users/brands
const SignIn = (entity) => {
    // will use this to navigate to the creator page
    const navigate = useNavigate();

    const [isValid, setIsValid] = useState(false);
    const [errors, setErrors] = useState({});
    const [formData, setFormData] = useState({
        email: '',
        password: '',
    });

    useEffect(() => {
        const currentErrors = getErrors();
        setErrors(currentErrors);
        setIsValid(Object.keys(currentErrors).length === 0);
    }, [formData]);

    const getErrors = () => {
        const newErrors = {};
        const { email, password } = formData;
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(email)) newErrors.email = 'Enter a valid email';
        if (password.length < 6) newErrors.password = 'Password required';
        return newErrors;
    }

    const handleChange = (e) => {
        setFormData({ ...formData, [e.target.name]: e.target.value });
    };

    const handleSubmit = () => {
        console.log('Submitting:', formData);
        navigate("/");
        // TODO : Send to backend here
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
                            <div className="button-group">
                                <button 
                                    onClick={handleSubmit}
                                    id={!isValid ? 'submit-disabled' : 'submit'}
                                    type="button"
                                    disabled={!isValid}
                                    >
                                    Log In
                                </button>
                            </div>
                        </form>
                        <p className="auth-footer">
                            Don't have an account? <a href={`/auth/${entity}/sign_up`}>Create One</a>
                        </p>
                    </div>
                </div>
            </div>
            
        </div>
    )
}


export default SignIn;