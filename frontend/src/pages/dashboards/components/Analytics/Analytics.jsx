import "./analytics.css";

const Analytics = () => {
    return (
        <div>
            <div className="analytics-chart-container">
                <h3 className="analytics-section-title">Performance Overview</h3>
                <div className="analytics-chart-placeholder">
                    <p>Analytics charts coming soon...</p>
                </div>
            </div>

            <div className="analytics-grid">
                <div className="analytics-card">
                    <h4 className="analytics-card-title">Top Performing Content</h4>
                    <div className="analytics-list">
                        {['Instagram Reel - EcoWear', 'YouTube Video - TechGadgets', 'TikTok - HealthPlus'].map((item, i) => (
                        <div key={i} className="analytics-list-item">
                            <span className="analytics-list-item-label">{item}</span>
                            <span className="analytics-list-item-value">{(10 - i * 2)}K views</span>
                        </div>
                        ))}
                    </div>
                </div>

                <div className="analytics-card">
                    <h4 className="analytics-card-title">Engagement Breakdown</h4>
                    <div>
                        {[{ label: 'Likes', value: '85%' }, { label: 'Comments', value: '65%' }, { label: 'Shares', value: '45%' }].map((item, i) => (
                        <div key={i} className="analytics-progress-item">
                            <div className="analytics-progress-header">
                            <span className="analytics-progress-label">{item.label}</span>
                            <span className="analytics-progress-percentage">{item.value}</span>
                            </div>
                            <div className="analytics-progress-bar">
                            <div className="analytics-progress-fill" style={{ width: item.value }}></div>
                            </div>
                        </div>
                        ))}
                    </div>
                </div>
            </div>
        </div>
    )
};

export default Analytics;