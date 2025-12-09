import { useState } from 'react';
import './dashboard.css';
import Profile from './components/Profile/Profile';
import Analytics from './components/Analytics/Analytics';
import Messages from './components/Messages/Messages';
import Overview from './components/Overview/Overview';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../AuthContext';

const Dashboard = () => {

    const navigate = useNavigate();
    const { user, loading } = useAuth();
    const [activeTab, setActiveTab] = useState('overview');
    const [sidebarOpen, setSidebarOpen] = useState(false);

    if (loading) return <div>Loading...</div>;

    if (!user) {
        navigate("/auth/brands/sign_in");
    }


    // Mock data
    const stats = {
        activeCampaigns: 5,
        totalReach: '2.4M',
        engagement: '8.5%',
        earnings: '$12,450'
    };

    const campaigns = [
        { id: 1, brand: 'EcoWear', status: 'active', deadline: '2025-11-15', budget: '$2,500' },
        { id: 2, brand: 'TechGadgets', status: 'pending', deadline: '2025-11-20', budget: '$3,200' },
        { id: 3, brand: 'HealthPlus', status: 'completed', deadline: '2025-10-30', budget: '$1,800' },
    ];

    const messages = [
        { id: 1, brand: 'EcoWear', preview: 'Can we schedule a call to discuss...', time: '2h ago', unread: true },
        { id: 2, brand: 'TechGadgets', preview: 'The content looks great! Just one...', time: '5h ago', unread: false },
        { id: 3, brand: 'HealthPlus', preview: 'Payment has been processed...', time: '1d ago', unread: true },
        { id: 3, brand: 'HealthPlus', preview: 'Payment has been processed...', time: '1d ago', unread: false },
    ];

    return (
        <div className="dashboard-container">
        {/* Sidebar */}
        <aside className={`sidebar ${sidebarOpen ? 'sidebar-open' : ''}`}>
            <div className="sidebar-inner">
            {/* Logo */}
            <div className="sidebar-header">
                <a className="sidebar-logo">FrogMedia</a>
                <p className="sidebar-subtitle">Brand Dashboard</p>
            </div>

            {/* Navigation */}
            <nav className="sidebar-navigation">
                <button
                onClick={() => setActiveTab('overview')}
                className={`sidebar-nav-button ${activeTab === 'overview' ? 'nav-button-active' : ''}`}
                >
                <svg className="sidebar-nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
                </svg>
                <span className="sidebar-nav-label">Overview</span>
                </button>

                <button
                onClick={() => setActiveTab('campaigns')}
                className={`sidebar-nav-button ${activeTab === 'campaigns' ? 'nav-button-active' : ''}`}
                >
                <svg className="sidebar-nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                </svg>
                <span className="sidebar-nav-label">Campaigns</span>
                </button>

                <button
                onClick={() => setActiveTab('messages')}
                className={`sidebar-nav-button ${activeTab === 'messages' ? 'nav-button-active' : ''}`}
                >
                <svg className="sidebar-nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" />
                </svg>
                <span className="sidebar-nav-label">Messages</span>
                </button>

                <button
                onClick={() => setActiveTab('analytics')}
                className={`sidebar-nav-button ${activeTab === 'analytics' ? 'nav-button-active' : ''}`}
                >
                <svg className="sidebar-nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
                <span className="sidebar-nav-label">Analytics</span>
                </button>

                <button
                onClick={() => setActiveTab('profile')}
                className={`sidebar-nav-button ${activeTab === 'profile' ? 'nav-button-active' : ''}`}
                >
                <svg className="sidebar-nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                </svg>
                <span className="sidebar-nav-label">Profile</span>
                </button>
            </nav>

            {/* User Info */}
            <div className="sidebar-user-info">
                <div className="sidebar-user-wrapper">
                <div className="sidebar-user-avatar">{user.username.charAt(0)}</div>
                <div className="sidebar-user-details">
                    <p className="sidebar-user-name">{user.username}</p>
                    <p className="sidebar-user-handle">{user.email}</p>
                </div>
                </div>
            </div>
            </div>
        </aside>

        {/* Overlay for mobile */}
        <div
            className={`mobile-overlay ${sidebarOpen ? 'overlay-visible' : ''}`}
            onClick={() => setSidebarOpen(false)}
        />

        {/* Main Content */}
        <div className="main-content-wrapper">
            {/* Top Bar */}
            <header className="top-bar">
            <div className="top-bar-inner">
                <div className="top-bar-left">
                <button
                    onClick={() => setSidebarOpen(!sidebarOpen)}
                    className="hamburger-button"
                >
                    <svg className="hamburger-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 6h16M4 12h16m-7 6h7" />
                    </svg>
                </button>
                <h2 className="page-title">
                    {activeTab.charAt(0).toUpperCase() + activeTab.slice(1)}
                </h2>
                </div>

                <div className="top-bar-right">
                <button className="notification-button">
                    <svg className="notification-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
                    </svg>
                    <span className="notification-badge"></span>
                </button>
                </div>
            </div>
            </header>

            {/* Content Area */}
            <main className="content-area">
            {activeTab === 'overview' && (<Overview stats={stats} campaigns={campaigns.slice(0,3)}/>)}

            {activeTab === 'campaigns' && (
                <div>
                <div className="campaigns-page-header">
                    <div className="campaigns-page-header-text">
                    <h3 className="campaigns-page-title">All Campaigns</h3>
                    <p className="campaigns-page-subtitle">Manage your brand partnerships</p>
                    </div>
                    <button className="button-new-campaign">+ New Campaign</button>
                </div>

                <div className="campaigns-table-container">
                    <div className="campaigns-table-wrapper">
                    <table className="campaigns-table">
                        <thead className="campaigns-table-head">
                        <tr>
                            <th className="campaigns-table-header">Brand</th>
                            <th className="campaigns-table-header">Status</th>
                            <th className="campaigns-table-header">Deadline</th>
                            <th className="campaigns-table-header">Budget</th>
                            <th className="campaigns-table-header">Actions</th>
                        </tr>
                        </thead>
                        <tbody className="campaigns-table-body">
                        {campaigns.map(campaign => (
                            <tr key={campaign.id} className="campaigns-table-row">
                            <td className="campaigns-table-cell">
                                <div className="campaigns-table-brand-cell">
                                <div className="campaigns-table-avatar">
                                    {campaign.brand.slice(0, 2)}
                                </div>
                                <span className="campaigns-table-brand-name">{campaign.brand}</span>
                                </div>
                            </td>
                            <td className="campaigns-table-cell">
                                <span className={`campaign-status-badge campaign-status-${campaign.status}`}>
                                {campaign.status}
                                </span>
                            </td>
                            <td className="campaigns-table-cell">
                                <span className="campaigns-table-text">{campaign.deadline}</span>
                            </td>
                            <td className="campaigns-table-cell">
                                <span className="campaigns-table-budget">{campaign.budget}</span>
                            </td>
                            <td className="campaigns-table-cell">
                                <button className="campaigns-table-action-button">View</button>
                            </td>
                            </tr>
                        ))}
                        </tbody>
                    </table>
                    </div>
                </div>
                </div>
            )}

            {activeTab === 'messages' && (<Messages messages={messages}/>)}
            {activeTab === 'analytics' && (<Analytics />)}

            {activeTab === 'profile' && (<Profile entity={"brands"}/>)}
            </main>
        </div>
        </div>
    );
};

export default Dashboard;