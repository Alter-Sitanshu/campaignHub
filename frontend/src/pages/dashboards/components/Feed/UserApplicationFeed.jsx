import { useEffect, useState } from "react";
import { api } from "../../../../api";
import ApplicationCard from "../../../../components/Card/ApplicationCard";

const UserApplicationFeed = () => {
    const [ page, setPage ] = useState(0);
    const [ appls, setAppls ] = useState([]);
    const [ hasNext, setNext ] = useState(false);
    const [loading, setLoading] = useState(false);
    const pageLimit = 20;

    const fetchApplications = async (page) => {
        try {
            setLoading(true);

            const endpoint = `/applications/my-applications?offset=${page * pageLimit}&limit=${pageLimit}`;
            const response = await api.get(endpoint);

            setAppls(response.data.data.applications);
            setNext(response.data.data.has_more);
        } catch (err) {
            console.error(err);
            setAppls([]);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchApplications(page);
    }, [page]);

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

    const formatDate = (dateString) => {
        const date = new Date(dateString);
        return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
    };

    let width = window.innerWidth;
    if (width < 880) {
        return (
            <>
            {loading ? <p>Loading applications...</p> : 
                <div className="campaigns-table-container">
                    <div className="campaigns-table-wrapper">
                        {appls?.length > 0 ? (
                            <div className="appl-card-box">
                                {appls.map((appl) => {
                                    return <ApplicationCard key={appl.id} application={appl}/>
                                })}
                            </div>
                            ) : 
                            <p className="empty-container-text">Start Collaborating today !</p>
                        }
                        {appls?.length === 0 ? null : (
                            <div className="page-buttons-box">
                                <button disabled={loading || page <= 0} onClick={() => {setPage(page - 1)}} className="page-button">Prev</button>
                                <button disabled={loading || !hasNext} onClick={() => {setPage(page + 1)}} className="page-button">Next</button>
                            </div>
                        )}
                    </div>
                </div>
            }
            </>
        )
    }

    return (
        <>
            {loading ? <p>Loading applications...</p> : 
            <div className="campaigns-table-container">
                <div className="campaigns-table-wrapper">
                    {appls?.length > 0 ? <table className="campaigns-table">
                    <thead className="campaigns-table-head">
                        <tr>
                            <th className="campaigns-table-header">S.No</th>
                            <th className="campaigns-table-header">Title</th>
                            <th className="campaigns-table-header">Brand</th>
                            <th className="campaigns-table-header">Status</th>
                            <th className="campaigns-table-header">Created At</th>

                        </tr>
                    </thead>

                    <tbody className="campaigns-table-body">
                        {appls?.map((appl, index) => (
                                <tr key={appl.id}
                                    className="campaigns-table-row"
                                >
        
                                    {/* Id */}
                                    <td className="campaigns-table-cell">
                                    <span className="campaigns-table-text">{(page * 20) + index + 1}</span>
                                    </td>
        
                                    {/* Campaign Title */}
                                    <td className="campaigns-table-cell">
                                    <span className="campaigns-table-text">{appl.campaign_name}</span>
                                    </td>

                                    {/* Brand Name */}
                                    <td className="campaigns-table-cell">
                                    <span className="campaigns-table-text">{appl.brand_name}</span>
                                    </td>
        
                                    {/* Status */}
                                    <td className="campaigns-table-cell">
                                    <span className={`campaign-status-badge campaign-status-${appl.status === 0 ? "draft" : "active"}`}>
                                        {getStatus(appl.status)}
                                    </span>
                                    </td>
        
                                    {/* Created */}
                                    <td className="campaigns-table-cell">
                                    <span className="campaigns-table-text">{formatDate(appl.created_at)}</span>
                                    </td>
        
                                </tr>
                            ))}
                    </tbody>
                    </table> : <p className="empty-container-text">Start Collaborating today !</p> }
                    {appls?.length === 0 ? null : (
                        <div className="page-buttons-box">
                            <button disabled={loading || page <= 0} onClick={() => {setPage(page - 1)}} className="page-button">Prev</button>
                            <button disabled={loading || !hasNext} onClick={() => {setPage(page + 1)}} className="page-button">Next</button>
                        </div>
                    )}
                </div>
            </div> }
        </>
    )
};

export default UserApplicationFeed;