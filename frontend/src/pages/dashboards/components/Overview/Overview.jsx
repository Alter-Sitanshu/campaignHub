import { useAuth } from "../../../../AuthContext";
import "./overview.css";
import CampaignCard from "../../../../components/Card/CampaignCard";
import SubCard from "../../../../components/Card/SubCard";
import { IndianRupeeIcon } from "lucide-react";
import { useState } from "react";
import TicketForm from "../../../../components/Tickets/TicketForm";

const Overview = ({ stats, campaigns, isUser = false }) => {
    const { user } = useAuth();
    const [isFeedbackOpen, setFeedbackOpen] = useState(false);

    if (!stats || !user) {
        return (
            <div className="empty-state">
                <div className="empty-message">
                    <h3>Loading stats...</h3>
                </div>
            </div>
        )
    }
    return (
        <div>
            {/* Stats Grid */}
            <div className="stats-grid">
                <div className="stat-card">
                <div className="stat-card-header">
                    <p className="stat-card-label">Total Campaigns</p>
                    <svg className="stat-card-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
                    </svg>
                </div>
                <p className="stat-card-value">{isUser ? stats?.total_submissions?? 0 : stats?.total_campaigns ?? 0}</p>
                <p className="stat-card-description">Lifetime</p>
                </div>

                <div className="stat-card">
                <div className="stat-card-header">
                    <p className="stat-card-label">Total Applications</p>
                    <svg className="stat-card-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                    </svg>
                </div>
                <p className="stat-card-value">{stats?.total_applications ?? 0}</p>
                <p className="stat-card-description">Across all platforms . Lifetime</p>
                </div>
                {isUser && (

                    <div className="stat-card">
                    <div className="stat-card-header">
                        <p className="stat-card-label">{"Success Ratio"}</p>
                        <svg className="stat-card-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                        </svg>
                    </div>
                    <p className="stat-card-value">{stats.success_rate}</p>
                    <p className="stat-card-description">Across all platforms . Lifetime</p>
                    </div>
                )}

                <div className="stat-card">
                <div className="stat-card-header">
                    <p className="stat-card-label">Total Transactions</p>
                    <svg className="stat-card-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
                    </svg>
                </div>
                <p className="stat-card-value">{stats?.total_transactions ?? 0}</p>
                <p className="stat-card-description">Lifetime</p>
                </div>

                <div className="stat-card">
                <div className="stat-card-header">
                    <p className="stat-card-label">Total Earnings</p>
                    <IndianRupeeIcon className="stat-card-icon"/>
                </div>
                <p className="stat-card-value">&#8377;{stats?.total_spent ?? stats?.total_earning ?? 0.0}</p>
                <p className="stat-card-description">Net</p>
                </div>
            </div>

            {/* Recent Activity */}
            <div className="campaigns-section">
                <h3 className="campaigns-section-title">Recent Campaigns</h3>
                {!isUser && (
                    <div className="campaigns-list">
                    {campaigns?.length > 0 ? campaigns?.map(campaign => (
                        <CampaignCard key={campaign.id} campaign={campaign} isBrand={true} />
                    )) : <p className="empty-container-text">Start Collaborating today !</p>}
                    </div>
                )}
                {isUser && (
                    <div className="campaigns-list">
                        {campaigns?.submissions?.length > 0 ? campaigns?.submissions.map(campaign => (
                            <SubCard key={`sub-${campaign.submission_id}`} sub={campaign} />
                        )) : <p className="empty-container-text">Start Collaborating today !</p>}
                    </div>
                )}
            </div>
            <button onClick={() => setFeedbackOpen(true)} className="feedback-btn">Feedback</button>
            <TicketForm isOpen={isFeedbackOpen} onClose={() => setFeedbackOpen(false)}/>
        </div>
    )
};

export default Overview;