import { useState } from 'react';
import './CreateAccountForm.css';
import { useAuth } from '../../../../AuthContext';
import { api } from '../../../../api.js';
import { X, CheckCircle } from "lucide-react";

const CreateAccountForm = ({ isOpen, onClose}) => {
    const { user, login } = useAuth();
    const [ formData, setFormData ] = useState({
        amount: 0,
        currency: "inr",
    });
    const [ success, setSuccess ] = useState(false);
    const [ isLoading, setIsLoading ] = useState(false);
    
    const handleSubmit = async () => {
        setIsLoading(true);
        try {
            const type_ = user?.entity === "users" ? "user" : "brand";
            const payload = {
                holder_id: user?.id,
                type: type_,
                amount: formData.amount,
                currency: formData.currency,
            }
            const response = await api.post("/accounts", payload);
            if (response.status === 201) {
                setSuccess(true);
                setIsLoading(false);
                setTimeout(handleSuccess, 1200);
                localStorage.setItem("currency", payload.currency);
            }
        } catch (err) {
            console.log("could not create account", err);
        }
    }

    const handleCurrency = (e) => {
        setFormData(prev => ({...prev, ['currency']: e.value}));
    }

    const handleSuccess = () => {
        const updated_user_object = {
            ...user,
            ["account_exists"]: true,
        }
        login(updated_user_object);
    }

    return (
        <>
            {isOpen && !success && (
                <div className="modal-overlay">
                    <div className="account-modal-container">
                        <div className="modal-header">
                            <div className="modal-header-content">
                                <div>
                                    <h2 className="modal-title">New Account Form</h2>
                                    <p className="modal-subtitle">creates a new app wallet for you</p>
                                </div>
                                <button
                                    onClick={onClose}
                                    className="modal-close-btn"
                                >
                                    <X className="modal-close-icon" />
                                </button>
                            </div>
                        </div>

                        <div className="modal-body">
                            <div className="form-group">
                                <label className='form-label'>Amount</label>
                                <input type="number" 
                                    value={formData.amount}
                                    readOnly={true}
                                    name='amount'
                                />
                            </div>
                            <div className="form-group">
                                <label className='from-label'>Currency</label>
                                <select name="currency" 
                                    value={formData.currency}
                                    onChange={handleCurrency}
                                    className='form-select'
                                >
                                    <option value="inr">INR</option>
                                    <option value="yen">YEN</option>
                                    <option value="usd">USD</option>
                                </select>
                            </div>
                            <div className="modal-footer">
                                <button type="button" 
                                    disabled={isLoading} 
                                    onClick={onClose} 
                                    className="btn btn-cancel"
                                >
                                    Cancel
                                </button>
                                <button
                                    type="button"
                                    onClick={handleSubmit}
                                    className="btn btn-submit"
                                    disabled={isLoading}
                                    >
                                    {isLoading ? (
                                        <div className="loading-container">
                                            <div className="loading-spinner"></div>
                                            <p>Submitting...</p>
                                        </div>
                                    ) : (
                                        "Create"
                                    )}
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            )}
            {isOpen && success && (
                <div className="modal-overlay">
                    <div className="account-modal-container">
                        <div className="modal-header">
                            <div className="modal-header-content">
                                <div>
                                    <h2 className="modal-title">New Account Created</h2>
                                    <p className="modal-subtitle">start collaborating today !</p>
                                </div>
                                <button
                                    onClick={onClose}
                                    className="modal-close-btn"
                                >
                                    <X className="modal-close-icon" />
                                </button>
                            </div>
                        </div>

                        <div className="modal-body">
                            <div className='success-message'>
                                <p>Congratulations</p>
                                <CheckCircle size={50} className="application-icon"/>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </>
    )

};

export default CreateAccountForm;