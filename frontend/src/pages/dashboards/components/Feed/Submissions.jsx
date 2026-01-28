import { useEffect, useState } from "react";
import { api } from "../../../../api";
import SubmissionsForm from "../../../submissions/SubmissionsForm";
import SubCard from "../../../../components/Card/SubCard";
import { Plus } from "lucide-react";
import SubmissionsModal from "../../../submissions/SubmissionsModal";
import { useAuth } from "../../../../AuthContext";

const Submissions = () => {
    const [loading, setLoading] = useState(true);
    const [subs, setSubs] = useState([]);
    const [pageNum, setPageNum] = useState(0);
    const [hasMore, setHasMore] = useState(false);
    const [isApplicationModalOpen, setApplicationModalOpen] = useState(false);
    const { user } = useAuth();
    
    const pageLimit = 20;

    const fetchSubmissions = async (pageNum) => {
        // safe fallback
        if (pageNum < 0) return [];

        try {
            let endpoint = `/submissions/my-submissions?offset=${pageNum * pageLimit}&limit=${pageLimit}`;
            let res = await api.get(endpoint);
            setSubs(res.data.data.submissions);
            setHasMore(res.data.data.has_more);
        } catch (err) {
            console.log("error getting user submissions", err);
        }
        setLoading(false);
    };

    

    const handleNewSubmission = () => {
        setApplicationModalOpen(true);
    };

    const handleSubmissionSuccess = () => {
        fetchSubmissions(pageNum); // Refresh submissions list
        setApplicationModalOpen(false); // Close the modal after successful submission
    };


    useEffect(() => {
        fetchSubmissions(pageNum);
    }, [pageNum]);

    return (
        <>
            {loading ? <p>Loading submissions...</p> :
                <div>
                    <div className="campaigns-page-header">
                        <div className="campaigns-page-header-text">
                            <h3 className="campaigns-page-title">All Submissions</h3>
                            <p className="campaigns-page-subtitle">Manage your brand partnerships</p>
                        </div>
                        <button
                            className="button-new-submission"
                            onClick={handleNewSubmission}
                            disabled={!user.account_exists}
                        >
                            <Plus size={16} className="button-icon" />
                            New Submission
                        </button>
                    </div>

                    {/* Application Selection Modal */}
                    <SubmissionsModal 
                        isOpen={isApplicationModalOpen}
                        onClose={() => setApplicationModalOpen(false)}
                        handleSuccess={handleSubmissionSuccess}
                    />

                    <div className="submissions-table-container">
                        <div className="submissions-table-wrapper">
                            {subs.length > 0 ? subs.map((sub, i) => (
                                <SubCard key={`sub-${i}`} sub={sub} />
                            )) : <p className="empty-container-text">Start Collaborating with brands today!</p>}
                            {subs?.length > 0 ?
                                <div className="page-buttons-box">
                                    <button
                                        disabled={loading || pageNum <= 0}
                                        onClick={() => { setPageNum(pageNum - 1) }}
                                        className="page-button"
                                    >
                                        Prev
                                    </button>
                                    <button
                                        disabled={loading || !hasMore}
                                        onClick={() => { setPageNum(pageNum + 1) }}
                                        className="page-button"
                                    >
                                        Next
                                    </button>
                                </div> : null}
                        </div>
                    </div>
                </div>
            }
        </>
    )
};

export default Submissions;