import { useEffect, useState } from "react";
import { X, CheckCircle } from "lucide-react";
import SubmissionsForm from "./SubmissionsForm";
import { api } from "../../api";
import './SubmissionsModal.css';


const SubmissionsModal = ({ handleSuccess, onClose, isOpen }) => {

    const [applications, setApplications] = useState([]);
    const [applicationsLoading, setApplicationsLoading] = useState(false);
    const [isSubmissionOpen, setSubmissionOpen] = useState(false);
    const [selectedApplication, setSelectedApplication] = useState(null);


    const fetchApplicationsWithoutSubmissions = async () => {
        setApplicationsLoading(true);
        try {
            const res = await api.get("/applications/my-applications/available");
            setApplications(res.data.data);
        } catch (err) {
            console.log("error getting applications without submissions", err);
        }
        setApplicationsLoading(false);
    };

    useEffect(() => {
        if (isOpen && !isSubmissionOpen) {
            fetchApplicationsWithoutSubmissions();
        }
    }, [isOpen, isSubmissionOpen])

    const handleApplicationSelect = (application) => {
        setSelectedApplication(application);
        setSubmissionOpen(true);
    };

    const handleSubmissionClose = () => {
        setSubmissionOpen(false);
        setSelectedApplication(null);
        // When form closes, go back to application selection
        fetchApplicationsWithoutSubmissions();
    };

    const handleSubmissionSuccess = () => {
        handleSuccess(); // Call parent's success handler
        // Close the entire modal after successful submission
        setSubmissionOpen(false);
        setSelectedApplication(null);
        onClose();
    };

    const handleModalClose = () => {
        if (isSubmissionOpen) {
            // If form is open, close it first
            handleSubmissionClose();
        } else {
            // If application selection is showing, close the modal
            onClose();
        }
    };

    if (!isOpen) return null;

    return (
        <>
            {/* Application Selection Modal */}
            {isOpen && !isSubmissionOpen && (
                <div className="modal-overlay">
                    <div className="modal-container submissions-modal">
                        <div className="modal-header">
                            <div className="modal-header-content">
                                <div>
                                    <h2 className="modal-title">Select Campaign</h2>
                                    <p className="modal-subtitle">Choose a campaign to submit content for</p>
                                </div>
                                <button
                                    onClick={handleModalClose}
                                    className="modal-close-btn"
                                >
                                    <X className="modal-close-icon" />
                                </button>
                            </div>
                        </div>

                        <div className="modal-body">
                            {applicationsLoading ? (
                                <div className="loading-container">
                                    <div className="loading-spinner"></div>
                                    <p>Loading available campaigns...</p>
                                </div>
                            ) : applications?.length > 0 ? (
                                <div className="applications-list">
                                    {applications.map((app) => (
                                        <div
                                            key={app.id}
                                            onClick={() => handleApplicationSelect(app)}
                                            className="application-card"
                                        >
                                            <div className="application-content">
                                                <div className="application-info">
                                                    <h4>{app.campaign_name}</h4>
                                                    <p>by {app.brand_name}</p>
                                                </div>
                                                <CheckCircle size={20} className="application-icon" />
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            ) : (
                                <div className="empty-state">
                                    <p>No campaigns available for submission.</p>
                                    <p>Apply to campaigns first, then come back to submit content.</p>
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            )}

            {/* Submission Form */}
            {isSubmissionOpen && (
                <SubmissionsForm
                    isOpen={isSubmissionOpen}
                    onClose={handleSubmissionClose}
                    campaignId={selectedApplication?.campaign_id}
                    onSuccess={handleSubmissionSuccess}
                />
            )}
        </>
    )
}

export default SubmissionsModal;