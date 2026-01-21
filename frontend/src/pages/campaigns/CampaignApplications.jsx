import { useEffect, useState } from "react";
import { api } from "../../api";
import { X } from "lucide-react";
import "./CampaignApplications.css";

const CampaignApplication = ({ isOpen, campaign_id, onClose, accepting = false }) => {
    const [ accept, setAccept ] = useState(accepting);
    const [ loading, setLoading ] = useState(true);
    const [ Message, setMessage ] = useState("Your Applications");
    const [ appls, setAppls ] = useState([]);

    const fetchCampaignApplications = async (campaign_id) => {
        try {
            const response = await api.get(`/applications/campaigns/${campaign_id}`);
            setAppls(response.data.data);
        } catch (err) {
            console.log(err);
            setMessage("Couldn't fetch applications. Try again");
        }
    }

    const getStatus = (status) => {
        switch (status) {
        case 0:
            return 'Rejected';
        case 1:
            return 'Accepted';
        case 2:
            return 'Pending';
        default:
            return status;
        }
    };
    
    const handleApprove = async (applicationId) => {
        try {
            await api.put(`/applications/accept/${applicationId}`);
            setAppls(prev =>
                prev.map(appl =>
                    appl.id === applicationId ? { ...appl, status: 1 } : appl
                )
            );
        } catch (err) {
            console.log(err);
            // Handle error, maybe show message
        }
    };

    const handleStopApplications = async () => {
        try {
            await api.put(`/campaigns/stop_applications/${campaign_id}`);
            setAccept(false);
        } catch (err) {
            console.log(err);
            // Handle error, maybe show message
        }
    }

    const handleReject = async (applicationId) => {
        try {
            await api.put(`/applications/reject/${applicationId}`);
            setAppls(prev =>
                prev.map(appl =>
                    appl.id === applicationId ? { ...appl, status: 0 } : appl
                )
            );
        } catch (err) {
            console.log(err);
            // Handle error
        }
    };
    
    // page loading
    useEffect(() => {
        if (isOpen) {
            fetchCampaignApplications(campaign_id);
            setLoading(false);
        }
    }, [isOpen]);

    // Sync accept state with accepting prop
    useEffect(() => {
        setAccept(accepting);
    }, [accepting]);

    if (!isOpen) return null;
    return (
        <div className="appl-overlay">
            <div className="appl-container">
                <div className="appl-header">
                <div className="appl-header-content">
                    <div>
                        <h2 className="appl-title">{Message}</h2>
                        <p className="appl-subtitle">Campaign Applications Below</p>
                    </div>
                    <div className="appl-button-group">
                        {accept ? 
                            <button 
                                onClick={handleStopApplications} 
                                className="stop-applications-btn"
                                disabled={!accept}
                            >Stop Applications</button> : <p>Applications Closed</p>
                        }
                        <button onClick={onClose} className="appl-close-btn">
                        <X className="appl-close-icon" />
                        </button>
                        
                    </div>
                </div>
                </div>

                <div className="appl-body">
                    {appls?.map((application) => (
                        <div key={application.id} className="application-overlay-card">
                            <div className="application-info">
                                <p className="creator-name">Creator: {application.creator_name || application.creator_id}</p>
                                <p className="application-type">Status: {getStatus(application.status)}</p>
                                <p className="application-date">Applied on: {new Date(application.created_at).toLocaleDateString()}</p>
                            </div>
                            {application.status === 2 ? 
                                <div className="application-actions">
                                    <button className="approve-btn" onClick={() => handleApprove(application.id)}>Approve</button>
                                    <button className="reject-btn" onClick={() => handleReject(application.id)}>Reject</button>
                                </div> : 
                                <div className={`status-${getStatus(application.status).toLowerCase()}`}>
                                    {getStatus(application.status)}
                                </div>
                            }
                        </div>
                    ))}
                </div>
            </div>
        </div>
    )
}

export default CampaignApplication;