const FeatureCard = ({title, short}) => {
    return (
        <div className="feature-card">
            <div className="feature-icon">
                <svg xmlns="http://www.w3.org/2000/svg" className="icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
            </div>
            <h3 className="feature-title">{title}</h3>
            <p className="feature-description">{short}</p>
        </div>
    )
};

export default FeatureCard;