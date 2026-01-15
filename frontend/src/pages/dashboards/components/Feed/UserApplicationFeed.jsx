import { useEffect, useState } from "react";
import { api } from "../../../../api";

const UserApplicationFeed = () => {
    const [ page, setPage ] = useState(0);
    const [ subs, setSubs ] = useState([{
    id: "app_001",
    campaign_id: "camp_101",
    creator_id: "user_001",
    status: "pending",
    created_at: "2025-01-02T10:15:30Z"
  },
  {
    id: "app_002",
    campaign_id: "camp_102",
    creator_id: "user_002",
    status: "accepted",
    created_at: "2025-01-03T11:20:10Z"
  },
  {
    id: "app_003",
    campaign_id: "camp_103",
    creator_id: "user_003",
    status: "rejected",
    created_at: "2025-01-04T09:05:45Z"
  },
  {
    id: "app_004",
    campaign_id: "camp_104",
    creator_id: "user_001",
    status: "pending",
    created_at: "2025-01-05T14:40:00Z"
  },
  {
    id: "app_005",
    campaign_id: "camp_105",
    creator_id: "user_004",
    status: "accepted",
    created_at: "2025-01-06T08:30:15Z"
  },
  {
    id: "app_006",
    campaign_id: "camp_106",
    creator_id: "user_002",
    status: "rejected",
    created_at: "2025-01-07T16:55:50Z"
  },
  {
    id: "app_007",
    campaign_id: "camp_107",
    creator_id: "user_005",
    status: "pending",
    created_at: "2025-01-08T12:10:05Z"
  },
  {
    id: "app_008",
    campaign_id: "camp_108",
    creator_id: "user_003",
    status: "accepted",
    created_at: "2025-01-09T18:25:35Z"
  },
  {
    id: "app_009",
    campaign_id: "camp_109",
    creator_id: "user_006",
    status: "rejected",
    created_at: "2025-01-10T07:45:20Z"
  },
  {
    id: "app_010",
    campaign_id: "camp_110",
    creator_id: "user_001",
    status: "pending",
    created_at: "2025-01-11T21:00:00Z"
  }]);
    const [ hasNext, setNext ] = useState(false);
    const [loading, setLoading] = useState(false);
    const pageLimit = 20;

    const fetchApplications = async (page) => {
        try {
            setLoading(true);

            const endpoint = `/applications/my-applications?offset=${page * pageLimit}&limit=${pageLimit}`;
            const response = await api.get(endpoint);

            setSubs(response.data.data.applications);
            setNext(response.data.data.has_more);
        } catch (err) {
            console.error(err);
            setSubs([]);
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

    return (
        <>
            {loading ? <p>Loading applications...</p> : 
            <div className="campaigns-table-container">
                <div className="campaigns-table-wrapper">
                    {subs?.length > 0 ? <table className="campaigns-table">
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
                        {subs?.map((sub, index) => (
                                <tr key={sub.id}
                                    className="campaigns-table-row"
                                >
        
                                    {/* Id */}
                                    <td className="campaigns-table-cell">
                                    <span className="campaigns-table-text">{(page * 20) + index + 1}</span>
                                    </td>
        
                                    {/* Campaign Title */}
                                    <td className="campaigns-table-cell">
                                    <span className="campaigns-table-text">{sub.campaign_name}</span>
                                    </td>

                                    {/* Brand Name */}
                                    <td className="campaigns-table-cell">
                                    <span className="campaigns-table-text">{sub.brand_name}</span>
                                    </td>
        
                                    {/* Status */}
                                    <td className="campaigns-table-cell">
                                    <span className={`campaign-status-badge campaign-status-${sub.status === 0 ? "draft" : "active"}`}>
                                        {getStatus(sub.status)}
                                    </span>
                                    </td>
        
                                    {/* Created */}
                                    <td className="campaigns-table-cell">
                                    <span className="campaigns-table-text">{formatDate(sub.created_at)}</span>
                                    </td>
        
                                </tr>
                            ))}
                    </tbody>
                    </table> : <p className="empty-container-text">Start Collaborating today !</p> }
                    {subs?.length === 0 ? null : (
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