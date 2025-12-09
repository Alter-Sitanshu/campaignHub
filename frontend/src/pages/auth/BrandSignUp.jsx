import { useEffect, useState } from "react";
import { signup } from "../../api.js";
import { useNavigate } from "react-router-dom";

const BrandSignUp = () => {
    const navigate = useNavigate();
    const [isLoading, setIsLoading] = useState(false);
    const [ formData, setFormData ] = useState(
        {
            "name": "",
            "email": "",
            "sector": "",
            "password": "",
            "address": "",
            "website": "",
        }
    );
    const [ errors, setErrors ] = useState({});
    const [ isValid, setIsValid ] = useState(false);
    const [ step, setStep ] = useState(1);

    useEffect(() => {
        const err = getErrors();
        setErrors(err);
        setIsValid(Object.keys(err).length === 0);
    }, [formData, step])

    const handleSubmit = async () => {
        setIsLoading(true);
        const payload = {
            "name": formData.name,
            "email": formData.email,
            "sector": formData.sector,
            "password": formData.password,
            "website": formData.website,
            "address": formData.address,
        }
        const resp = await signup(payload, "brands");
        if (resp.type !== "error") {
            navigate("/auth/brands/sign_in");
        } else {
            navigate(`/errors/${resp.status}/`);
        }
    }

    const handleChange = (e) => {
        setFormData(
            {...formData, [e.target.name]: e.target.value}
        );
    };

    const getErrors = () => {
        const newErrors = {};

        // --- Common regex patterns ---
        const emailRegex = /^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$/;
        const urlRegex = /^(https?:\/\/)?([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}(:\d+)?(\/\S*)?$/;

        if (step === 1) {
            const { name, email, password } = formData;

            // Name
            if (!name || !name.trim()) {
                newErrors.name = 'Name is required';
            } else if (name.trim().length < 2) {
                newErrors.name = 'Name must be at least 2 characters long';
            }

            // Email
            if (!email || !email.trim()) {
                newErrors.email = 'Email is required';
            } else if (!emailRegex.test(email.trim())) {
                newErrors.email = 'Enter a valid email address';
            }

            // Password
            if (password) {
                if (password.trim().length < 6) {
                    newErrors.password = 'Password must be at least 6 characters long';
                } else if (!/[A-Z]/.test(password) || !/[0-9]/.test(password)) {
                    newErrors.password = 'Include at least one uppercase letter and one number';
                }
            }
        }

        if (step === 2) {
            const { sector, address, website } = formData;

            // Sector
            if (!sector || !sector.trim()) {
                newErrors.sector = 'Sector is required';
            }

            // Address
            if (!address || !address.trim()) {
                newErrors.address = 'Address is required';
            } else if (address.trim().length < 10) {
                newErrors.address = 'Address should be at least 10 characters';
            }

            // Website
            if (!website || !website.trim()) {
                newErrors.website = 'Website link is required';
            } else if (!urlRegex.test(website.trim())) {
                newErrors.website = 'Enter a valid website URL (e.g., https://example.com)';
            }
        }

        return newErrors;
    };

    const validateStep = () => {
        const newErrors = getErrors();
        setErrors(newErrors);
        return Object.keys(newErrors).length === 0;
    };

    const nextStep = () => {
        const valid = validateStep();
        if (valid && step < 2) setStep(step + 1);
    };

    const prevStep = () => {
        if (step > 1) setStep(step - 1);
    };

    return (
        <div className='form-page'>
        <div className="auth-page">
            <div className="auth-container">
            <div className="signup-box">
                <div className="auth-header">
                <h1>Create your account</h1>
                <p>
                    Join <span>FrogMedia</span> and start collaborating instantly
                </p>
                <div className="progressBar">
                    <div className="progressFill" style={{ width: `${(step / 2) * 100}%` }}></div>
                </div>
                <div className="steps-indicator">Step {step} of 2</div>
                </div>

                <form className="form">
                {step === 1 && (
                    <>
                    <div className="form-group">
                        <label htmlFor="name">Name</label>
                        <input
                        type="text"
                        id="name"
                        name="name"
                        placeholder="FrogMedia"
                        value={formData.name}
                        onChange={handleChange}
                        spellCheck={false}
                        />
                        {errors.name && <small className="error-text">{errors.name}</small>}
                    </div>
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
                        {errors.password && <small className="error-text">{errors.password}</small>}
                    </div>
                    </>
                )}

                {step === 2 && (
                    <>
                        <div className="form-group">
                            <label htmlFor="sector">Sector</label>
                            <input
                            type="text"
                            id="sector"
                            name="sector"
                            placeholder="Supplements/Beauty/etc."
                            value={formData.sector}
                            onChange={handleChange}
                            spellCheck={false}
                            />
                            {errors.sector && <small className="error-text">{errors.sector}</small>}
                        </div>
                        <div className="form-group">
                            <label htmlFor="address">Address</label>
                            <input
                                type="text"
                                id="address"
                                name="address"
                                placeholder="10th Street, Amingaon, Assam"
                                value={formData.address}
                                onChange={handleChange}
                                spellCheck={false}
                            />
                            {errors.address && <small className="error-text">{errors.address}</small>}
                        </div>
                        <div className="form-group">
                            <label htmlFor="website">Website Link</label>
                            <input
                                type="text"
                                id="website"
                                name="website"
                                placeholder="https://example.com"
                                value={formData.website}
                                onChange={handleChange}
                                spellCheck={false}
                            />
                            {errors.website && <small className="error-text">{errors.website}</small>}
                        </div>
                    </>
                )}
                <div className="button-group">
                    {step > 1 && (
                    <button onClick={prevStep} className="btn-secondary" type="button">
                        Back
                    </button>
                    )}
                    {step < 2 ? (
                    <button 
                        onClick={nextStep}
                        id={!isValid ? 'submit-disabled' : 'submit'}
                        type="button"
                        disabled={!isValid}
                    >
                        Next
                    </button>
                    ) : (
                    <button 
                        onClick={handleSubmit}
                        id={!isValid ? 'submit-disabled' : 'submit'}
                        type="button"
                        disabled={!isValid}
                    >
                        {isLoading ? "Submitting..." : "Create Account"}
                    </button>
                    )}
                </div>
                </form>

                <p className="auth-footer">
                    Already have an account? <a href="/auth/brands/sign_in">Sign in</a>
                </p>
            </div>
            </div>
        </div>
        </div>
    );
};


export default BrandSignUp;