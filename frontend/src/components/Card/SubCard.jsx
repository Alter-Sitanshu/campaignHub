import "./SubCard.css";
import fallbackThumbnail from "../../assets/samplethumb.jpg";

const SubCard = ({ sub }) => {
  // tolerate both naming conventions for counts
  const views = sub.view_count ?? sub.views ?? 0;
  const likes = sub.like_count ?? sub.likeCount ?? 0;
  const title = sub.title ?? sub.video_title ?? "Untitled";
  const platform = (sub.platform ?? sub.video_platform ?? "unknown").toLowerCase();
  const videoStatus = (sub.video_status ?? sub.videoStatus ?? "unknown").toLowerCase();

  const formatCount = (n) => {
    if (n >= 1_000_000) return (n / 1_000_000).toFixed(1).replace(/\.0$/, "") + "M";
    if (n >= 1_000) return (n / 1_000).toFixed(1).replace(/\.0$/, "") + "K";
    return n.toLocaleString();
  };

  const timeAgo = (iso) => {
    if (!iso) return "-";
    const diff = Math.floor((Date.now() - new Date(iso)) / 1000);
    if (diff < 60) return `${diff}s ago`;
    if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
    if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
    if (diff < 604800) return `${Math.floor(diff / 86400)}d ago`;
    if (diff < 2419200) return `${Math.floor(diff / 604800)}w ago`;
    // fallback to a readable date
    return new Date(iso).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });
  };

  const capitalize = (s) => (s ? s.charAt(0).toUpperCase() + s.slice(1) : s);

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
            <p className="sub-title" title={title}>{title}</p>
            <div className="sub-meta">
              <span className="sub-platform">{capitalize(platform)}</span>
              <span className="sub-sep">•</span>
              <span className="sub-date">Submitted {timeAgo(sub.uploaded_at)}</span>
            </div>
          </div>

          <div className="sub-badges">
            <span className={`badge badge-${videoStatus}`}>{capitalize(videoStatus)}</span>
            <span className={`platform-badge platform-${platform}`}>{capitalize(platform)}</span>
          </div>
        </div>

        <div className="sub-body">
          <div className="metrics">
            <div className="metric">
              <svg viewBox="0 0 24 24" className="metric-icon" aria-hidden>
                <path fill="currentColor" d="M12 5a7 7 0 100 14 7 7 0 000-14zm0 2a5 5 0 110 10 5 5 0 010-10z" />
                <path fill="currentColor" d="M2 12s4-7 10-7 10 7 10 7-4 7-10 7S2 12 2 12z" opacity="0.15" />
              </svg>
              <div>
                <div className="metric-number">{formatCount(views)}</div>
                <div className="metric-label">views</div>
              </div>
            </div>

            <div className="metric">
              <svg viewBox="0 0 24 24" className="metric-icon" aria-hidden>
                <path fill="currentColor" d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 6 4 4 6.5 4c1.74 0 3.41 1 4.5 2.09C12.09 5 13.76 4 15.5 4 18 4 20 6 20 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z" />
              </svg>
              <div>
                <div className="metric-number">{formatCount(likes)}</div>
                <div className="metric-label">likes</div>
              </div>
            </div>

            <div className="metric">
              <svg viewBox="0 0 24 24" className="metric-icon" aria-hidden>
                <path fill="currentColor" d="M12 1v2M12 21v2M4.2 4.2l1.4 1.4M18.4 18.4l1.4 1.4M1 12h2M21 12h2M4.2 19.8l1.4-1.4M18.4 5.6l1.4-1.4" opacity="0.15" />
                <path fill="currentColor" d="M12 8c-2.2 0-4 1.3-4 3v1a2 2 0 002 2h4a2 2 0 002-2v-1c0-1.7-1.8-3-4-3z" />
              </svg>
              <div>
                <div className="metric-number">&#8377;{Number(sub.earnings ?? 0).toFixed(2)}</div>
                <div className="metric-label">earnings</div>
              </div>
            </div>
          </div>

          <div className="details-wrapper">
            <table>
              <tbody className="ids-grid">
                <tr className="id-row">
                  <td className="id-label">Submission</td>
                  <td className="id-val">{sub.submission_id ?? sub.id ?? "-"}</td>
                </tr>
                <tr className="id-row">
                  <td className="id-label">Campaign</td>
                  <td className="id-val">{sub.campaign_id ?? sub.campaign ?? "-"}</td>
                </tr>
                <tr className="id-row">
                  <td className="id-label">Video ID</td>
                  <td className="id-val">{sub.video_id ?? "-"}</td>
                </tr>
              </tbody>
            </table>
            <table>
              <tbody className="ids-grid">
                <tr className="id-row">
                  <td className="id-label">Applied</td>
                  <td className="id-val">{sub.created_at ? new Date(sub.created_at).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' }) : "-"}</td>
                </tr>
                <tr className="id-row">
                  <td className="id-label">Submitted</td>
                  <td className="id-val">{sub.uploaded_at ? timeAgo(sub.uploaded_at) : "-"}</td>
                </tr>
                <tr className="id-row">
                  <td className="id-label">Last Synced</td>
                  <td className="id-val">{sub.last_synced_at ? timeAgo(sub.last_synced_at) : "-"}</td>
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
