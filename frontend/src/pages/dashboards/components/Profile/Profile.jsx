import "./profile.css";
import { useState } from "react";
import defaultProfile from "../../../../assets/default-profile.avif";

export function validateImage(file) {
  return file.size < 2 * 1024 * 1024; // 2 MB
}

const Profile = () => {

    const [ selectedFile, setSelectedFile ] = useState(null);
    const [ preview, setPreview ] = useState(defaultProfile)
    
    function handlePhotoChange(e) {
        const file = e.target.files[0];
        if (!validateImage(file)) {
            return
        }
        const newPreview = URL.createObjectURL(selectedFile);
        setPreview(newPreview);

        setSelectedFile(file);
    }


    return (
        <div>
            <div className="profile-main-card">
                <div className="profile-content">
                    <div className="profile-avatar-section">
                        <div className="profile-avatar-large">
                            <img className="profile-avatar" src={preview}></img>
                        </div>
                        <label className="profile-change-photo-button">
                            Change Photo
                            <input 
                                type="file"
                                accept="image/*"
                                hidden
                                onChange={handlePhotoChange}
                            />
                        </label>
                    </div>

                    <div className="profile-form-section">
                        <div className="profile-form-group">
                        <label className="profile-form-label">Full Name</label>
                        <input type="text" defaultValue="Jane Doe" className="profile-form-input" />
                        </div>
                        <div className="profile-form-group">
                        <label className="profile-form-label">Username</label>
                        <input type="text" defaultValue="@janedoe" className="profile-form-input" />
                        </div>
                        <div className="profile-form-group">
                        <label className="profile-form-label">Email</label>
                        <input type="email" defaultValue="jane@example.com" className="profile-form-input" />
                        </div>
                        <div className="profile-form-group">
                        <label className="profile-form-label">Bio</label>
                        <textarea rows="4" className="profile-form-textarea" placeholder="Tell brands about yourself..."></textarea>
                        </div>
                        <button className="profile-save-button">Save Changes</button>
                    </div>
                </div>
            </div>

            <div className="profile-social-card">
                <div className="profile-social-header">
                    <h3 className="profile-social-title">Social Media Accounts</h3>
                    <button className="social-add-button">Add Link</button>
                </div>
                <div className="profile-social-list">
                {[
                    { platform: 'Instagram', handle: '@janedoe', followers: '125K', connected: true },
                    { platform: 'YouTube', handle: 'Jane Doe', followers: '89K', connected: true },
                    { platform: 'TikTok', handle: '@janedoecreates', followers: '210K', connected: true },
                ].map((account, i) => (
                    <div key={i} className="profile-social-item">
                        <div className="profile-social-item-left">
                            <div className="profile-social-avatar">
                                {account.platform.slice(0, 2)}
                            </div>
                            <div className="profile-social-info">
                                <p className="profile-social-platform">{account.platform}</p>
                                <p className="profile-social-details">{account.handle} • {account.followers} followers</p>
                            </div>
                        </div>
                        <button className={`profile-social-button profile-social-button-connected`}>
                            Connected
                        </button>
                    </div>
                ))}
                </div>
            </div>
        </div>
    )
};

export default Profile;