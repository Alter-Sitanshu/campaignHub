import "./feed.css";
import sampleThumb from "../../../../assets/default-profile.avif";
import SubCard from "../../../../components/Card/SubCard";
import CampaignCard from "../../../../components/Card/CampaignCard";

const Feed = () => {
    // Mock Campaigns
    const campaigns = [
        {
            "id": "cmp_001",
            "brand": "EcoWear",
            "title": "Winter Sale Awareness",
            "budget": 50000,
            "cpm": 120,
            "requirements": "Create 1 reel + 2 story posts",
            "platform": "instagram",
            "doc_link": "https://example.com/docs/cmp_001",
            "status": 1,
            "created_at": "2025-01-12T10:30:00Z"
        }, {
            "id": "cmp_002",
            "brand": "Tech Gadgets",
            "title": "New Product Launch",
            "budget": 75000,
            "cpm": 150,
            "requirements": "Unboxing video + 1 carousel post",
            "platform": "youtube",
            "doc_link": "https://example.com/docs/cmp_002",
            "status": 0,
            "created_at": "2025-01-15T14:45:00Z"
        }, {
            "id": "cmp_003",
            "brand": "HealthPlus",
            "title": "App Install Campaign",
            "budget": 30000,
            "cpm": 90,
            "requirements": "2 reels demonstrating app usage",
            "platform": "instagram",
            "doc_link": "https://example.com/docs/cmp_003",
            "status": 1,
            "created_at": "2025-01-18T09:20:00Z"
        }, {
            "id": "cmp_004",
            "brand": "TechReview Pro",
            "title": "Tech Gadget Review",
            "budget": 100000,
            "cpm": 200,
            "requirements": "Full YouTube review + 1 short",
            "platform": "youtube",
            "doc_link": "https://example.com/docs/cmp_004",
            "status": 2,
            "created_at": "2025-01-20T17:10:00Z"
        }, {
            "id": "cmp_005",
            "brand": "FestiveVibes",
            "title": "Festival Campaign",
            "budget": 45000,
            "cpm": 110,
            "requirements": "Festive-themed reel + 1 story",
            "platform": "instagram",
            "doc_link": "https://example.com/docs/cmp_005",
            "status": 1,
            "created_at": "2025-01-22T12:00:00Z"
        }
    ];

    return (
        <div>
            <div className="submissions-table-container">
                <div className="campaign-table-wrapper">
                    {campaigns.slice(0, 10).map((camp, i) => (
                        <CampaignCard key={`camp-${i}`} campaign={camp}/>
                    ))}
                </div>
            </div>
        </div>
    )
};

export default Feed;