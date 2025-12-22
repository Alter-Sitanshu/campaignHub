import "./SubCard.css";
import fallbackThumbnail from "../../assets/sampleThumb.jpg";

const SubCard = ({ sub }) => {
  // tolerate both naming conventions for counts
  const views = sub.view_count ?? sub.views ?? 0;
  const likes = sub.like_count ?? sub.likeCount ?? 0;
  const title = sub.title ?? sub.video_title ?? "Untitled";
  const platform = (sub.platform ?? sub.video_platform ?? "unknown").toLowerCase();
  const videoStatus = (sub.video_status ?? sub.videoStatus ?? "unknown").toLowerCase();

  return (
    <a className="sub-card detailed" role="article" aria-label={`${title} submission`} href={sub.url} target="_blank" rel="noopener noreferrer">
      <div className="thumbnail-wrapper">
        <img
          src={sub.thumbnail?.url !== "" ? sub.thumbnail.url : fallbackThumbnail}
          alt={title}
          className="thumbnail"
          loading="lazy"
        />
      </div>

      <div className="sub-content">
        <div className="sub-header">
          <div className="sub-title-wrap">
            <p className="sub-title">{title}</p>
            <p className="sub-platform">{platform.toUpperCase()}</p>
          </div>

          <div className="sub-badges">
            <span className={`badge badge-${videoStatus}`}>{videoStatus}</span>
            <span className={`platform-badge platform-${platform}`}>{platform}</span>
          </div>
        </div>

        <div className="sub-body">
          <div className="metrics">
            <div className="metric">
              <div className="metric-number">{views.toLocaleString()}</div>
              <div className="metric-label">views</div>
            </div>

            <div className="metric">
              <div className="metric-number">{likes.toLocaleString()}</div>
              <div className="metric-label">likes</div>
            </div>

            <div className="metric">
              <div className="metric-number">₹{Number(sub.earnings ?? 0).toFixed(2)}</div>
              <div className="metric-label">earnings</div>
            </div>
          </div>

          <div className="details-wrapper">
            <table>
              <tbody className="ids-grid">
                <tr className="id-row">
                  <td className="id-label">Submisssion</td>
                  <td className="id-val">{sub.submission_id ?? sub.id}</td>
                </tr>
                <tr className="id-row">
                  <td className="id-label">Campaign</td>
                  <td className="id-val">{sub.campaign_id ?? sub.campaign}</td>
                </tr>
                <tr className="id-row">
                  <td className="id-label">VideoID</td>
                  <td className="id-val">{sub.video_id}</td>
                </tr>
              </tbody>
            </table>
            <table>
              <tbody className="ids-grid">
                <tr className="id-row">
                  <td className="id-label">Applied</td>
                  <td className="id-val">{sub.created_at ? new Date(sub.created_at).toLocaleString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) : "-"}</td>
                </tr>
                <tr className="id-row">
                  <td className="id-label">Submitted</td>
                  <td className="id-val">{sub.uploaded_at ? new Date(sub.uploaded_at).toLocaleString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) : "-"}</td>
                </tr>
                <tr className="id-row">
                  <td className="id-label">Last Synced</td>
                  <td className="id-val">{sub.last_synced_at ? new Date(sub.last_synced_at).toLocaleString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) : "-"}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </a>
  );
};

export default SubCard;
