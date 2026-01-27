import { useState } from "react";
import { CheckCircle, X } from 'lucide-react';
import "./TicketForm.css";
import { useAuth } from "../../AuthContext";
import { api } from "../../api.js";

const TicketForm = ({ isOpen, onClose }) => {
    const { user } = useAuth();
    const [ isLoading, setIsLoading ] = useState(false);
    const [ formData, setFormData ] = useState({
        subject: "suggestion",
        message: "",
    });
    const [ success, setSuccess ] = useState(false);

    const handleChange = (e) => {
        setFormData(prev => ({
            ...prev,
            [e.target.name]: e.target.value,
        }));
    };

    const handleSubmit = async () => {
        try {
            const payload = {
                customer_id: user?.id,
                type: user?.entity === "users" ? "creator" : "brand",
                subject: formData.subject,
                message: formData.message,
            };
            let _ = await api.post("/tickets", payload);
            setIsLoading(false);
            setSuccess(true);
            setTimeout(handleSuccess, 1500);
        } catch (err) {
            console.log("could not raise ticket", err);
        }

    }

    const handleSuccess = () => {
        setFormData({
            subject: "suggestion",
            message: ""
        });
        setSuccess(false);
        onClose();
    }

    return (
        <>
        {isOpen && !success && (
            <div className="modal-overlay">
            <div className="ticket-container">
                <div className="modal-header">
                    <div className="modal-header-content">
                        <div>
                            <h2 className="modal-title">Raise Ticket</h2>
                            <p className="modal-subtitle">suggestions and bugs are welcome</p>
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
                        <label className="form-label">Subject</label>
                        <select name="subject" 
                            value={formData.subject}
                            onChange={handleChange}
                            className="form-select"
                            >
                            <option value="bug">Bug Report</option>
                            <option value="suggestion">Suggest Improvement</option>
                            <option value="complaint">Complaint</option>
                        </select>
                    </div>
                    <div className="form-group">
                        <label className="form-label">Message</label>
                        <textarea type="text" name="message" className="message"
                            placeholder="Enter your feedback here"
                            onChange={handleChange}
                        />
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
                                "Raise"
                            )}
                        </button>
                    </div>
                </div>
            </div>
        </div>
        )}
        {isOpen && success && (
            <div className="modal-overlay">
            <div className="ticket-container">
                <div className="modal-header">
                    <div className="modal-header-content">
                        <div>
                            <h2 className="modal-title">Ticket Raised</h2>
                            <p className="modal-subtitle">Thank you for your feedback</p>
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
                        <p>Thank you for Feedback</p>
                        <CheckCircle size={50} className="application-icon"/>
                    </div>
                </div>
            </div>
        </div>
        )}
        </>
    )
};

export default TicketForm;