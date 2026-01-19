import { useEffect, useState } from "react";
import SubCard from "../../components/Card/SubCard";
import Profile from "./components/Profile/Profile";
import Overview from "./components/Overview/Overview";
import MessagesPage from "../messages/MessagePage";
import Analytics from "./components/Analytics/Analytics";
import Feed from "./components/Feed/Feed";
import { useAuth } from "../../AuthContext";
import { useNavigate } from "react-router-dom";
import { api } from "../../api";
import UserApplicationFeed from "./components/Feed/UserApplicationFeed";

const UserDashboard = () => {
    const navigate = useNavigate();
    const { user, loading, logout } = useAuth();
    const [ activeTab, setActiveTab ] = useState("overview");
    const [ stats, setStats ] = useState(null);
    const [ subs, setSubs ] = useState([]);
    const [ isPageLoading, setIsPageLoading ] = useState(true);
    const [ sidebarOpen, setSidebarOpen ] = useState(false);

    const fetchStats = async (id) => {
        if (!id) return null;
        try {
            const endpoint = `/users/stats/${id}`;
            const resp = await api.get(endpoint);
            if (resp.status !== 200) {
                throw new Error(`Failed to fetch stats: ${resp.status}`);
            }
            return resp.data.data;
        } catch (err) {
            console.error("fetchStats error:", err);
            throw err;
        }
    }

    const fetchSubmissions = async () => {
        try {
            // limit is mandatory, offset and time filters are optional
            const endpoint = "/submissions/my-submissions?limit=10";
            const resp = await api.get(endpoint);
            if (resp.status !== 200) {
                throw new Error(`Failed to fetch submissions: ${resp.status}`);
            }
            return resp.data.data || [];
        } catch (err) {
            console.error("fetchSubmissions error:", err);
            return [];
        }
    }
    
    // page loading api calls
    useEffect(() => {
        if (loading) return;
        if (user === null) {
            navigate("/auth/sign_in");
            return;
        }
        let mounted = true;
        const load = async () => {
            setIsPageLoading(true);
            try {
                const [statsResp, submissionsResp] = await Promise.all([
                    fetchStats(user.id),
                    fetchSubmissions()
                ]);
                if (!mounted) return;
                if (statsResp) {
                    let submissionRate = 0.0;
                    if (statsResp.total_submissions) {
                        submissionRate = statsResp.total_submissions / (statsResp.total_applications == 0 ? 1 : statsResp.total_applications);
                    }
                    setStats({...statsResp, "success_rate": submissionRate});
                }
                setSubs(submissionsResp || []);
            } catch (err) {
                console.error("Error loading dashboard data:", err);
            } finally {
                if (mounted) setIsPageLoading(false);
            }
        };

        load();

        return () => { mounted = false; };
    }, [loading, user]);


    if (loading || isPageLoading) return <div>Loading...</div>;

    const handleLogout = () => {
        logout();
    }

    if (!user) return null;
    return (
        <>
            <div className="dashboard-container">
                <aside className={`sidebar ${sidebarOpen ? 'sidebar-open' : ''}`}>
                    <div className="sidebar-inner">
                        <div className="sidebar-header">
                            <a href="/" className="sidebar-logo">FrogMedia</a>
                            <p className="sidebar-subtitle">Creator Dashboard</p>
                        </div>
                        <div className="sidebar-navigation">
                            {/* Overview View Button */}
                            <button
                                onClick={() => setActiveTab('overview')}
                                className={`sidebar-nav-button ${activeTab === 'overview' ? 'nav-button-active' : ''}`}
                                >
                                <svg className="sidebar-nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
                                </svg>
                                <span className="sidebar-nav-label">Overview</span>
                            </button>
                            {/* Campaigns View Button */}
                            <button
                                onClick={() => setActiveTab('submissions')}
                                className={`sidebar-nav-button ${activeTab === 'submissions' ? 'nav-button-active' : ''}`}
                                >
                                <svg className="sidebar-nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                                </svg>
                                <span className="sidebar-nav-label">Submissions</span>
                            </button>
                            {/* Applications of user */}
                            <button
                                onClick={() => setActiveTab('applications')}
                                className={`sidebar-nav-button ${activeTab === 'applications' ? 'nav-button-active' : ''}`}
                                >
                                <svg className="sidebar-nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                                </svg>
                                <span className="sidebar-nav-label">Applications</span>
                            </button>
                            {/* Messages View Button */}
                            <button
                                onClick={() => setActiveTab('messages')}
                                className={`sidebar-nav-button ${activeTab === 'messages' ? 'nav-button-active' : ''}`}
                                >
                                <svg className="sidebar-nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" />
                                </svg>
                                <span className="sidebar-nav-label">Messages</span>
                            </button>
                            {/* Analytics View Button */}
                            <button
                                onClick={() => setActiveTab('analytics')}
                                className={`sidebar-nav-button ${activeTab === 'analytics' ? 'nav-button-active' : ''}`}
                                >
                                <svg className="sidebar-nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                                </svg>
                                <span className="sidebar-nav-label">Analytics</span>
                            </button>
                            {/* Campaign Feed Button */}
                            <button
                                onClick={() => setActiveTab('feed')}
                                className={`sidebar-nav-button ${activeTab === 'feed' ? 'nav-button-active' : ''}`}
                                >
                                <svg className="sidebar-nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 20H5a2 2 0 01-2-2V7a2 2 0 012-2h1m13 0a2 2 0 012 2v11a2 2 0 01-2 2M6 5v2m0 0h12M6 7v11m6-11v11m6-11v11M9 10h2m-2 4h2" />
                                </svg>
                                <span className="sidebar-nav-label">Feed</span>
                            </button>
                            {/* User Profile Button */}
                            <button
                                onClick={() => setActiveTab('profile')}
                                className={`sidebar-nav-button ${activeTab === 'profile' ? 'nav-button-active' : ''}`}
                                >
                                <svg className="sidebar-nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                                </svg>
                                <span className="sidebar-nav-label">Profile</span>
                            </button>
                        </div>
                        {/* User Info */}
                        <div className="sidebar-user-info">
                            <div className="sidebar-user-wrapper">
                            <div className="sidebar-user-avatar">{user.username.charAt(0)}</div>
                            <div className="sidebar-user-details">
                                <div className="name-container">
                                    <p className="sidebar-user-name">{user.username}</p>
                                    <button className="logout-button" onClick={handleLogout}>logout</button>
                                </div>
                                <p className="sidebar-user-handle">{user.email}</p>
                            </div>
                            </div>
                        </div>
                    </div>
                </aside>
                {/* Overlay div for Mobile View */}
                <div
                    className={`mobile-overlay ${sidebarOpen ? 'overlay-visible' : ''}`}
                    onClick={() => setSidebarOpen(false)}
                />
                <div className="main-content-wrapper">
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
                        </div>
                    </header>
                    <main className="content-area">
                        {activeTab === "overview" && ( <Overview stats={stats} campaigns={subs} isUser={true}/>)}
                        {activeTab === "submissions" && (
                            <div>
                                <div className="campaigns-page-header">
                                    <div className="campaigns-page-header-text">
                                    <h3 className="campaigns-page-title">All Campaigns</h3>
                                    <p className="campaigns-page-subtitle">Manage your brand partnerships</p>
                                    </div>
                                </div>
                                <div className="submissions-table-container">
                                    <div className="submissions-table-wrapper">
                                        {subs.length > 0 ? subs.map((sub, i) => (
                                            <SubCard key={`sub-${i}`} sub={sub} />
                                        )): <p className="empty-container-text">Start Collaborating with brands today !</p>}
                                    </div>
                                </div>
                            </div>
                        )}
                        {activeTab === 'applications' && (<UserApplicationFeed />)}
                        {activeTab === 'messages' && navigate(`/userss/dashboard/${user.id}/messages`)}
                        {activeTab === 'analytics' && (<Analytics />)}
                        {activeTab === 'feed' && (<Feed />)}
                        {activeTab === 'profile' && (<Profile entity={"users"}/>)}
                    </main>
                </div>
            </div>
        </>
    )
};

export default UserDashboard;