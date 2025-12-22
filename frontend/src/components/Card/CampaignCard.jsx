import "./CampaignCard.css";
import { api } from "../../api";
import { useNavigate } from "react-router-dom";
import { useState } from "react";


const CampaignCard = ({ campaign, isBrand }) => {
    const navigate = useNavigate();
    const [ Applied, setApplied ] = useState(false);
    const [ Applying, setApplying ] = useState(false);

    async function handleApply() {
        setApplying(true);
        const response = await api.post(`/applications/${campaign.id}`);
        if (response.status != 201) {
            setApplying(false);
            navigate(`/errors/${response.status}`);
        }
        setApplying(false);
        setApplied(true);
    }

    const getStatusConfig = (status) => {
        switch (status) {
        case 0:
            return 'Draft';
        case 1:
            return 'Active';
        case 3:
            return 'Completed';
        default:
            return 'Unknown';
        }
    };

    const getPlatformIcon = (platform) => {
        if (platform === 'instagram') {
            return (
                <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M12 2.163c3.204 0 3.584.012 4.85.07 3.252.148 4.771 1.691 4.919 4.919.058 1.265.069 1.645.069 4.849 0 3.205-.012 3.584-.069 4.849-.149 3.225-1.664 4.771-4.919 4.919-1.266.058-1.644.07-4.85.07-3.204 0-3.584-.012-4.849-.07-3.26-.149-4.771-1.699-4.919-4.92-.058-1.265-.07-1.644-.07-4.849 0-3.204.013-3.583.07-4.849.149-3.227 1.664-4.771 4.919-4.919 1.266-.057 1.645-.069 4.849-.069zm0-2.163c-3.259 0-3.667.014-4.947.072-4.358.2-6.78 2.618-6.98 6.98-.059 1.281-.073 1.689-.073 4.948 0 3.259.014 3.668.072 4.948.2 4.358 2.618 6.78 6.98 6.98 1.281.058 1.689.072 4.948.072 3.259 0 3.668-.014 4.948-.072 4.354-.2 6.782-2.618 6.979-6.98.059-1.28.073-1.689.073-4.948 0-3.259-.014-3.667-.072-4.947-.196-4.354-2.617-6.78-6.979-6.98-1.281-.059-1.69-.073-4.949-.073zm0 5.838c-3.403 0-6.162 2.759-6.162 6.162s2.759 6.163 6.162 6.163 6.162-2.759 6.162-6.163c0-3.403-2.759-6.162-6.162-6.162zm0 10.162c-2.209 0-4-1.79-4-4 0-2.209 1.791-4 4-4s4 1.791 4 4c0 2.21-1.791 4-4 4zm6.406-11.845c-.796 0-1.441.645-1.441 1.44s.645 1.44 1.441 1.44c.795 0 1.439-.645 1.439-1.44s-.644-1.44-1.439-1.44z"/>
                </svg>
            );
        }
        return (
            <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                <path d="M23.498 6.186a3.016 3.016 0 0 0-2.122-2.136C19.505 3.545 12 3.545 12 3.545s-7.505 0-9.377.505A3.017 3.017 0 0 0 .502 6.186C0 8.07 0 12 0 12s0 3.93.502 5.814a3.016 3.016 0 0 0 2.122 2.136c1.871.505 9.376.505 9.376.505s7.505 0 9.377-.505a3.015 3.015 0 0 0 2.122-2.136C24 15.93 24 12 24 12s0-3.93-.502-5.814zM9.545 15.568V8.432L15.818 12l-6.273 3.568z"/>
            </svg>
        );
    };

    const formatDate = (dateString) => {
        const date = new Date(dateString);
        return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
    };

    const formatCurrency = (amount) => {
        return new Intl.NumberFormat('en-IN', {
            style: 'currency',
            currency: 'INR',
            maximumFractionDigits: 0
        }).format(amount);
    };

    return (
        <div className="campaign-card-wrapper">
            <div className="campaign-card-header">
                <div className="campaign-header-left">
                    <div>
                        {getPlatformIcon( campaign.platform )}
                    </div>
                    <div className="campaign-name-wrapper">
                        <h4>{campaign.brand}</h4>
                        <p>{formatDate(campaign.created_at)}</p>
                    </div>
                </div>
                <button className={`campaign-status-${getStatusConfig(campaign.status).toLowerCase()}`}>
                    { getStatusConfig(campaign.status) }
                </button>
            </div>
            <div className="campaign-card-body">
                <h2>{campaign.title}</h2>
                <div className="campaign-req">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                        <path d="M14.5 2h-9A2.5 2.5 0 0 0 3 4.5v15A2.5 2.5 0 0 0 5.5 22h13A2.5 2.5 0 0 0 21 19.5v-11L14.5 2zm0 1.5 6 6h-4.5A1.5 1.5 0 0 1 14.5 8V3.5zM6 12h12v1.5H6V12zm0 4h12v1.5H6V16z"/>
                    </svg>
                    <p>{`${campaign.requirements.slice(0, 150)}...`}</p>
                </div>
            </div>
            <div className="campaign-card-money">
                <div className="budget">
                    <div>
                        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                            <path d="M13.66 7C13.1 5.82 11.9 5 10.5 5H6V3h12v2h-3.26c.48.58.84 1.26 1.05 2H18v2h-2.02c-.25 2.8-2.61 5-5.48 5h-.73l6.73 7h-2.77L7 14v-2h3.5c1.76 0 3.22-1.3 3.46-3H6V7h7.66z"/>
                        </svg>
                        <span>Budget</span>
                    </div>
                    <p>{campaign.budget.toLocaleString()}</p>
                </div>
                <div className="cpm">
                    <div>
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                            <path d="M12 12c2.76 0 5-2.24 5-5s-2.24-5-5-5-5 2.24-5 5 2.24 5 5 5zm0 2c-4 0-8 2-8 5v2h16v-2c0-3-4-5-8-5z"/>
                        </svg>
                        <span>CPM</span>
                    </div>
                    <p>{campaign.cpm.toLocaleString()}</p>
                </div>
            </div>
            <div className="campaign-card-footer">
                {!isBrand && (<button onClick={handleApply}
                    disabled={Applying || Applied}
                >{Applied ? "Applied" : Applying ? "Applying..." : "Apply Now"}</button>
                )}
                <a href="">Open Details</a>
            </div>
        </div>
    )
};

export default CampaignCard;