import { useAuth } from "../../../../AuthContext";
import "./overview.css";
import CampaignCard from "../../../../components/Card/CampaignCard";
import SubCard from "../../../../components/Card/SubCard";

const Overview = ({ stats, campaigns, isUser = false }) => {
    const { user } = useAuth();
    return (
        <div>
            {/* Stats Grid */}
            <div className="stats-grid">
                <div className="stat-card">
                <div className="stat-card-header">
                    <p className="stat-card-label">Active Campaigns</p>
                    <svg className="stat-card-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
                    </svg>
                </div>
                <p className="stat-card-value">{stats.activeCampaigns}</p>
                <p className="stat-card-description">+2 this month</p>
                </div>

                <div className="stat-card">
                <div className="stat-card-header">
                    <p className="stat-card-label">Total Reach</p>
                    <svg className="stat-card-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                    </svg>
                </div>
                <p className="stat-card-value">{stats.totalReach}</p>
                <p className="stat-card-description">Across all platforms</p>
                </div>

                <div className="stat-card">
                <div className="stat-card-header">
                    <p className="stat-card-label">Engagement Rate</p>
                    <svg className="stat-card-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
                    </svg>
                </div>
                <p className="stat-card-value">{stats.engagement}</p>
                <p className="stat-card-description">Above average</p>
                </div>

                <div className="stat-card">
                <div className="stat-card-header">
                    <p className="stat-card-label">Total Earnings</p>
                    <svg className="stat-card-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                </div>
                <p className="stat-card-value">{stats.earnings}</p>
                <p className="stat-card-description">This year</p>
                </div>
            </div>

            {/* Recent Activity */}
            <div className="campaigns-section">
                <h3 className="campaigns-section-title">Recent Campaigns</h3>
                {!isUser && (
                    <div className="campaigns-list">
                    {campaigns?.map(campaign => (
                        <CampaignCard key={campaign.id} campaign={campaign} isBrand={true} />
                    ))}
                    </div>
                )}
                {isUser && (
                    <div className="campaigns-list">
                        {campaigns?.map(campaign => (
                            <SubCard key={`sub-${campaign.submission_id}`} sub={campaign} />
                        ))}
                    </div>
                )}
            </div>
        </div>
    )
};

export default Overview;