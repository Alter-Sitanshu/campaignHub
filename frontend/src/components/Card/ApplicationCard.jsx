import './ApplicationCard.css';

const ApplicationCard = ({ application }) => {
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

    const getStatusClass = (status) => {
        switch (status) {
            case 0:
                return 'rejected';
            case 1:
                return 'accepted';
            case 2:
                return 'pending';
            default:
                return 'unknown';
        }
    };

    const formatDate = (dateString) => {
        const date = new Date(dateString);
        return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
    };

    return (
        <div className="application-card">
            <div className="application-card-header">
                <h3 className="application-campaign-name">{application.campaign_name}</h3>
                <span className={`application-status application-status-${getStatusClass(application.status)}`}>
                    {getStatus(application.status)}
                </span>
            </div>
            <div className="application-card-body">
                <p className="application-brand-name">Brand: {application.brand_name}</p>
                <p className="application-created-at">Applied on: {formatDate(application.created_at)}</p>
            </div>
        </div>
    );
};

export default ApplicationCard;