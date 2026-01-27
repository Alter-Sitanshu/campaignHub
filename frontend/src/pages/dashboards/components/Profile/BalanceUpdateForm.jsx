import { useAuth } from "../../../../AuthContext";
import { useState } from 'react';
import './CreateAccountForm.css';
import { api } from '../../../../api.js';
import { X, CheckCircle, AlertCircle } from "lucide-react";

const BalanceUpdateForm = ({ balance, isOpen, action, onSuccess, onClose }) => {

    const { user } = useAuth();
    const [ formData, setFormData ] = useState({
        amount: 0,
        currency: localStorage.getItem('currency') ?? 'inr',
    });
    const [ error_amount, setErrorAmountMsg ] = useState(null);
    const [ success, setSuccess ] = useState(false);
    const [ isLoading, setIsLoading ] = useState(false);
    
    const handleSubmit = () => {
        switch (action) {
            case 'deposit': 
                handleDeposit();
                break;
            case 'withdraw':
                handleWithdraw();
                break;
            default:
                console.log("error unsupported action for balance:", action);
        }
    }

    const handleChange = (e) => {
        setFormData(prev => ({
            ...prev,
            [e.target.name]: e.target.value,
        }))
    }

    const handleDeposit = async () => {
        try {
            const payload = {
                to_id: user?.id,
                from_id: user?.id,
                amount: parseFloat(formData.amount),
                currency: formData.currency,
            }
            const endpoint = `accounts/deposit`;
            const response = await api.put(endpoint, payload);

            setSuccess(true);
            setIsLoading(false);
            setTimeout(() => {
                onSuccess(payload.amount);
            }, 1200);
            onClose();

        } catch (err) {
            console.log("error depositing balance", err);
        }
    };
    const handleWithdraw = async () => {
        if(balance <= 0) {
            return;
        }
        const amt = parseFloat(formData.amount);
        if (balance < amt) {
            setErrorAmountMsg("cannot exceed balance");
        }
        try {
            const payload = {
                to_id: user?.id,
                from_id: user?.id,
                amount: amt,
                currency: formData.currency
            }
            const endpoint = `accounts/withdraw`;
            const response = await api.put(endpoint, payload);

            setSuccess(true);
            setIsLoading(false);
            setTimeout(() => {
                onSuccess(payload.amount);
            }, 1200);
            onClose();

        } catch (err) {
            console.log(err);
        }
    };

    return (
        <>
        {isOpen && !success && (
            <div className="modal-overlay">
                <div className="account-modal-container">
                    <div className="modal-header">
                        <div className="modal-header-content">
                            <div>
                                <h2 className="modal-title">{
                                    action?.charAt(0).toUpperCase().concat(action.slice(1))
                                } Form</h2>
                                <p className="modal-subtitle">update account balance</p>
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
                                placeholder={formData.amount}
                                name='amount'
                                onChange={handleChange}
                            />
                            {error_amount && (
                            <div className="error-message">
                                <AlertCircle size={16} />
                                {error_amount}
                            </div>
                            )}
                        </div>
                        <div className="form-group">
                            <label className='from-label'>Currency</label>
                            <select name="currency" 
                                value={formData.currency}
                                readOnly={true}
                                disabled={true}
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
                                        <p>{
                                            action?.charAt(0).toUpperCase().concat(action.slice(1))
                                        }ing...</p>
                                    </div>
                                ) : (
                                    action?.charAt(0).toUpperCase().concat(action.slice(1))
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

export default BalanceUpdateForm;
