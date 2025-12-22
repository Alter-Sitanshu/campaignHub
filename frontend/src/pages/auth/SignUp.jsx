import { useState, useEffect } from 'react';
import './signup.css';
import { useNavigate } from 'react-router-dom';
import { signup } from '../../api';

export default function SignupPage() {
  // will use this to navigate to the creator page
  const navigate = useNavigate();

  const [isLoading, setIsLoading] = useState(false);
  const [errors, setErrors] = useState({});
  const [step, setStep] = useState(1);
  const [isStepValid, setIsStepValid] = useState(false);
  const [formData, setFormData] = useState({
    first_name: '',
    last_name: '',
    email: '',
    password: '',
    gender: '',
    age: '',
    links: []
  });

  useEffect(() => {
    const currentErrors = getStepErrors();
    setErrors(currentErrors);
    setIsStepValid(Object.keys(currentErrors).length === 0);
  }, [formData, step]);
  
  const getStepErrors = () => {
    const newErrors = {};
    if (step === 1) {
      const { first_name, last_name, email, password } = formData;
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

      if (!first_name.trim()) newErrors.first_name = 'First name is required';
      if (!last_name.trim()) newErrors.last_name = 'Last name is required';
      if (!emailRegex.test(email)) newErrors.email = 'Enter a valid email';
      if (password) {
          if (password.trim().length < 6) {
              newErrors.password = 'Password must be at least 6 characters long';
          } else if (!/[A-Z]/.test(password) || !/[0-9]/.test(password)) {
              newErrors.password = 'Include at least one uppercase letter and one number';
          }
      }
    }

    if (step === 2) {
      const { age } = formData;
      if (!age || age < 13 || age > 100) newErrors.age = 'Age must be between 13 and 100';
    }

    if (step === 3) {
       if (formData.links.length === 0) {
          newErrors.links = 'Add at least one platform link';
        } else {
          formData.links.forEach((link, index) => {
            if (!link.platform || !/^https?:\/\/.+/.test(link.url)) {
              newErrors[`link_${index}`] = 'Enter a valid platform and URL';
            }
          });
        }
    }

    return newErrors;
    
  };

  const validateStep = () => {
    const newErrors = getStepErrors();
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleChange = (e) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  const handleLinkChange = (index, field, value) => {
    const updatedLinks = [...formData.links];
    updatedLinks[index] = { ...updatedLinks[index], [field]: value };
    setFormData({ ...formData, links: updatedLinks });
  };

  const addLink = () => {
    setFormData({ ...formData, links: [...formData.links, { platform: '', url: '' }] });
  };

  const removeLink = (index) => {
    const updatedLinks = formData.links.filter((_, i) => i !== index);
    setFormData({ ...formData, links: updatedLinks });
  };

  const nextStep = () => {
    const valid = validateStep();
    if (valid && step < 3) setStep(step + 1);
  };

  const prevStep = () => {
    if (step > 1) setStep(step - 1);
  };

  const handleSubmit = async () => {
    setIsLoading(true);
    let payload = {
      "first_name": formData.first_name,
      "last_name": formData.last_name,
      "email": formData.email,
      "password": formData.password.trim(),
      "gender": "",
      "age": parseInt(formData.age),
      "links": formData.links.map((link) => {
        return {
          "platform": link.platform,
          "url": link.url,
        }
      })
    };
    if( formData.gender === "male" ){
      payload.gender = "M";
    } else if (formData.gender === "female" ){
      payload.gender = "F";
    } else {
      payload.gender = "O";
    };
    const resp = await signup(payload, "users");
    if (resp.type !== "error") {
      navigate("/auth/sign_in");
    } else {
      navigate(`/errors/${resp.status}/`);
    }
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
                <div className="progressFill" style={{ width: `${(step / 3) * 100}%` }}></div>
              </div>
              <div className="steps-indicator">Step {step} of 3</div>
            </div>

            <form className="form">
              {step === 1 && (
                <>
                  <div className="form-group">
                    <label htmlFor="first_name">First Name</label>
                    <input
                      type="text"
                      id="first_name"
                      name="first_name"
                      placeholder="John"
                      value={formData.first_name}
                      onChange={handleChange}
                      spellCheck={false}
                    />
                    {errors.first_name && <small className="error-text">{errors.first_name}</small>}
                  </div>
                  <div className="form-group">
                    <label htmlFor="last_name">Last Name</label>
                    <input
                      type="text"
                      id="last_name"
                      name="last_name"
                      placeholder="Doe"
                      value={formData.last_name}
                      onChange={handleChange}
                      spellCheck={false}
                    />
                    {errors.last_name && <small className="error-text">{errors.last_name}</small>}
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
                    <label htmlFor="gender">Gender</label>
                    <select
                      id="gender"
                      name="gender"
                      value={formData.gender}
                      onChange={handleChange}
                    >
                      <option value="">Select gender</option>
                      <option value="male">Male</option>
                      <option value="female">Female</option>
                      <option value="other">Other</option>
                      <option value="prefer_not_to_say">Prefer not to say</option>
                    </select>
                    {errors.gender && <small className="error-text">{errors.gender}</small>}
                  </div>
                  <div className="form-group">
                    <label htmlFor="age">Age</label>
                    <input
                      type="number"
                      id="age"
                      name="age"
                      placeholder="25"
                      value={formData.age}
                      onChange={handleChange}
                      min="13"
                      max="120"
                    />
                    {errors.age && <small className="error-text">{errors.age}</small>}
                  </div>
                </>
              )}

              {step === 3 && (
                <div className="form-group">
                  <label>Platform Links</label>
                  {formData.links.map((link, index) => (
                    <div key={index} className="link-group">
                      <select
                        id='platform-select'
                        value={link.platform}
                        onChange={(e) => handleLinkChange(index, 'platform', e.target.value)}
                        className="link-select"
                      >
                        <option value="">Select platform</option>
                        <option value="instagram">Instagram</option>
                        <option value="youtube">YouTube</option>
                        <option value="twitter">Twitter</option>
                        <option value="tiktok">TikTok</option>
                        <option value="linkedin">LinkedIn</option>
                      </select>
                      <input
                        type="url"
                        placeholder="https://..."
                        value={link.url}
                        onChange={(e) => handleLinkChange(index, 'url', e.target.value)}
                        className="link-input"
                      />
                      <button
                        onClick={() => removeLink(index)}
                        className="btn-remove"
                        type="button"
                      >
                        ×
                      </button>
                    </div>
                  ))}
                  <button onClick={addLink} className="btn-add" type="button">
                    + Add Platform
                  </button>
                </div>
              )}

              <div className="button-group">
                {step > 1 && (
                  <button onClick={prevStep} className="btn-secondary" type="button">
                    Back
                  </button>
                )}
                {step < 3 ? (
                  <button 
                    onClick={nextStep}
                    id={!isStepValid ? 'submit-disabled' : 'submit'}
                    type="button"
                    disabled={!isStepValid}
                  >
                    Next
                  </button>
                ) : (
                  <button 
                    onClick={handleSubmit}
                    id={!isStepValid ? 'submit-disabled' : 'submit'}
                    type="button"
                    disabled={!isStepValid || isLoading}
                  >
                    {isLoading ? "Signing up..." : "Create Account"}
                  </button>
                )}
              </div>
            </form>

            <p className="auth-footer">
              Already have an account? <a href="/auth/sign_in">Sign in</a>
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}