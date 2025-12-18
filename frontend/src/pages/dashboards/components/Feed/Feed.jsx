import CampaignCard from "../../../../components/Card/CampaignCard";
import { useEffect, useState } from "react";
import { api } from "../../../../api";

import { useInfiniteQuery } from '@tanstack/react-query';
import { useInView } from 'react-intersection-observer';

const Feed = () => {
    // 1. Setup the infinite query
    const { 
        data, 
        fetchNextPage, 
        hasNextPage, 
        isFetchingNextPage 
    } = useInfiniteQuery({
        queryKey: ['campaignFeed'],
        queryFn: async ({ pageParam = "" }) => {
            // Pass the cursor (pageParam) to backend
            const res = await api.get(`/campaigns/feed?cursor=${pageParam}`);
            return res.data.data;
        },
        // Extract the cursor for the NEXT call from the CURRENT response
        getNextPageParam: (lastPage) => lastPage.meta.has_more ? lastPage.meta.cursor : undefined,
        staleTime: 60 * 1000,          // 1 min cache
        cacheTime: 5 * 60 * 1000,      // 5 min keep in memory
        refetchOnWindowFocus: false,
        refetchOnReconnect: false,
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
    }, [inView, hasNextPage, isFetchingNextPage]);

    return (
        <div>
            <div className="submissions-table-container">
                {data?.pages.map((group, i) => (
                    <div key={i}>
                        {group.campaigns?.map(campaign => (
                            <CampaignCard key={campaign.id} 
                                campaign={campaign}
                            />
                        ))}
                    </div>
                ))}
            </div>
            {/* Invisible Trigger Div */}
            <div ref={ref} className="loading-trigger">
                {isFetchingNextPage ? 'Loading more...' : ''}
            </div>
        </div>
    )
};

export default Feed;