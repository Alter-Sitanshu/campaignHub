import { useInfiniteQuery, useQueryClient } from "@tanstack/react-query";
import { useInView } from "react-intersection-observer";
import { useAuth } from "../../../../AuthContext";
import CampaignLaunchModal from "../../../campaigns/CampaignLaunchModal";
import { useState, useEffect } from "react";
import { api } from "../../../../api";
import CampaignCard from "../../../../components/Card/CampaignCard";
import CampaignApplication from "../../../campaigns/CampaignApplications";


const BrandCampaignFeed = () => {
    const [ LaunchCampaign, setLaunchCampaign ] = useState(false);
    const [ fillForm, setfillForm ] = useState(null);
    const [ openApplication, setApplications ] = useState(false);
    const [ campaign_id , setCampaign ] = useState("");
    const [ accept_appl, setAcceptApplication ] = useState(true);
    const [width, setWidth] = useState(window.outerWidth);
    const { user } = useAuth();
    const queryClient = useQueryClient();

    const { 
        data, 
        fetchNextPage, 
        hasNextPage, 
        isFetchingNextPage 
    } = useInfiniteQuery({
        queryKey: ['brandCampaignFeed', user?.id],
        initialPageParam: "",
        queryFn: async ({ pageParam = "" }) => {
            // Pass the cursor (pageParam) to backend
            const res = await api.get(`/campaigns/brand/${user?.id}?cursor=${pageParam}`);
            return res.data.data;
        },
        // Extract the cursor for the NEXT call from the CURRENT response
        getNextPageParam: (lastPage) => lastPage.meta.has_more ? lastPage.meta.cursor : undefined,
        staleTime: 60 * 1000,          // 1 min cache
        cacheTime: 5 * 60 * 1000,      // 5 min keep in memory
        refetchOnWindowFocus: false,
        refetchOnReconnect: false,
        enabled: !!user?.id,
    });

    // 2. Setup the scroll trigger (invisible div at bottom)
    const { ref, inView } = useInView({
        threshold: 0,
        rootMargin: '200px', // preload before reaching bottom
    });

    // 3. Auto-load when trigger comes into view
    useEffect(() => {
        if (inView && hasNextPage && !isFetchingNextPage) {
            fetchNextPage();
        }
    }, [inView, hasNextPage, isFetchingNextPage, fetchNextPage]);

    // Handle window resize for responsive design
    useEffect(() => {
        const handleResize = () => setWidth(window.outerWidth);
        window.addEventListener('resize', handleResize);
        return () => window.removeEventListener('resize', handleResize);
    }, []);

    const getStatus = (status) => {
        switch (status) {
        case 0:
            return 'Draft';
        case 1:
            return 'Active';
        case 2:
            return 'Pending';
        case 3:
            return 'Completed';
        default:
            return 'Unknown';
        }
    };


    const handleStatus = async (campaignId, currentStatus) => {
        // Calculate new status immediately
        const newStatus = currentStatus === 0 ? 1 : 3;
        const endpoint = newStatus === 3 ? `/campaigns/stop/${campaignId}` : `/campaigns/activate/${campaignId}`;
        const expectedStatus = 204;

        const response = await api.put(endpoint);
        
        if (response.status === expectedStatus) {
            queryClient.invalidateQueries({ queryKey: ['brandCampaignFeed'] });
        } else {
            alert("Error updating campaign status");
        }
    };

    const formatDate = (dateString) => {
        const date = new Date(dateString);
        return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
    };

    if (width < 880) {
        return (
            <>
            <div className="campaigns-page-header">
                <div className="campaigns-page-header-text">
                <h3 className="campaigns-page-title">All Campaigns</h3>
                <p className="campaigns-page-subtitle">Manage your brand partnerships</p>
                </div>
                <button className="button-new-campaign" onClick={() => {
                        setfillForm(null);
                        setLaunchCampaign(true);
                    }}
                >+ New Campaign</button>
            </div>
            < CampaignLaunchModal
                key={fillForm ? fillForm.id : "modal-key"}
                isOpen={LaunchCampaign}
                onClose={() => {setLaunchCampaign(false)}}
                brandId={user.id}
                fillForm={fillForm}
            />
            <div className="campaigns-section">
                <h3 className="campaigns-section-title">Recent Campaigns</h3>
                <div className="campaigns-list">
                    {
                        data ? data.pages.map((group) => (
                            group.campaigns?.map((campaign) => (
                                <CampaignCard
                                    onClick={() => {
                                        setApplications(true)
                                        setCampaign(campaign.id)
                                    }}
                                    key={campaign.id} 
                                    campaign={campaign} 
                                    isBrand={true} 
                                />
                            ))
                        )) : <p className="empty-container-text">Start Collaborating today !</p>
                    }
                </div>
                <CampaignApplication 
                    isOpen={openApplication} 
                    onClose={() => setApplications(false)}
                    campaign_id={campaign_id}
                    accepting={accept_appl}
                />
            </div>
            </>
        )
    }

    return (
        <>
            <div className="campaigns-page-header">
                <div className="campaigns-page-header-text">
                <h3 className="campaigns-page-title">All Campaigns</h3>
                <p className="campaigns-page-subtitle">Manage your brand partnerships</p>
                </div>
                <button className="button-new-campaign" onClick={() => {
                        setfillForm(null);
                        setLaunchCampaign(true);
                    }}
                >+ New Campaign</button>
            </div>
            < CampaignLaunchModal
                key={fillForm ? fillForm.id : "modal-key"}
                isOpen={LaunchCampaign}
                onClose={() => {setLaunchCampaign(false)}}
                brandId={user.id}
                fillForm={fillForm}
            />
            <div className="campaigns-table-container">
                <div className="campaigns-table-wrapper">
                    <table className="campaigns-table">
                    <thead className="campaigns-table-head">
                        <tr>
                            <th className="campaigns-table-header">Id</th>
                            <th className="campaigns-table-header">Title</th>
                            <th className="campaigns-table-header">Status</th>
                            <th className="campaigns-table-header">Created</th>
                            <th className="campaigns-table-header">Budget</th>
                            <th className="campaigns-table-header">CPM</th>
                            <th className="campaigns-table-header">Action</th>
                            <th className="campaigns-table-header">Platform</th>
                            <th className="campaigns-table-header">Doc</th>
                        </tr>
                    </thead>

                    <tbody className="campaigns-table-body">
                        {data?.pages.map((group, pageIndex) => (
                            group.campaigns?.map((campaign, campaignIndex) => (
                                <tr key={campaign.id}
                                    className="campaigns-table-row"
                                    id={campaign.status === 0 ? "tr-draft" : "tr-active"}
                                    onClick={campaign.status === 0 ? () => {
                                        setfillForm(campaign);
                                        setLaunchCampaign(true);
                                    } : () => {
                                        setApplications(true);
                                        setCampaign(campaign.id);
                                        setAcceptApplication(campaign.accepting_applications)
                                    }}
                                >
        
                                    {/* Id */}
                                    <td className="campaigns-table-cell">
                                    <span className="campaigns-table-text">{(pageIndex * 10) + campaignIndex + 1}</span>
                                    </td>
        
                                    {/* Title */}
                                    <td className="campaigns-table-cell">
                                    <span className="campaigns-table-text">{campaign.title}</span>
                                    </td>
        
                                    {/* Status */}
                                    <td className="campaigns-table-cell">
                                    <span className={`campaign-status-badge campaign-status-${campaign.status === 0 ? "draft" : "active"}`}>
                                        {getStatus(campaign.status)}
                                    </span>
                                    </td>
        
                                    {/* Created */}
                                    <td className="campaigns-table-cell">
                                    <span className="campaigns-table-text">{formatDate(campaign.created_at)}</span>
                                    </td>
        
                                    {/* Budget */}
                                    <td className="campaigns-table-cell">
                                    <span className="campaigns-table-budget">{campaign.budget.toLocaleString()}</span>
                                    </td>
        
                                    {/* CPM */}
                                    <td className="campaigns-table-cell">
                                    <span className="campaigns-table-text">{campaign.cpm.toLocaleString()}</span>
                                    </td>
        
                                    {/* Action */}
                                    <td className="campaigns-table-cell">
                                    <span className="campaigns-table-action"
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            handleStatus(campaign.id, campaign.status)
                                        }
                                    }>{
                                        campaign.status === 0 ? "Go Live" : campaign.status === 1 ? "End" : getStatus(campaign.status)    
                                    }</span>
                                    </td>
        
                                    {/* Platform */}
                                    <td className="campaigns-table-cell">
                                    <span className="campaigns-table-text">{campaign.platform.charAt(0).toUpperCase() + campaign.platform.slice(1)}</span>
                                    </td>
        
                                    {/* DocLink Button */}
                                    <td className="campaigns-table-cell">
                                    <button
                                        className="campaigns-table-action-button"
                                        onClick={(e) => {
                                                e.stopPropagation();
                                                window.open(campaign.doc_link, "_blank");
                                            }
                                        }
                                    >
                                        Doc
                                    </button>
                                    </td>
        
                                </tr>
                            ))))}
                    </tbody>
                    </table>
                    <CampaignApplication 
                        isOpen={openApplication} 
                        onClose={() => setApplications(false)}
                        campaign_id={campaign_id}
                        accepting={accept_appl}
                    />
                    {/* Invisible Trigger Div */}
                    <div ref={ref} className="loading-trigger">
                        {isFetchingNextPage ? 'Loading more...' : ''}
                    </div>
                </div>
            </div>
        </>
    )
};

export default BrandCampaignFeed;