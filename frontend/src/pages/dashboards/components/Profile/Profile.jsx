import "./profile.css";
import { useEffect, useState } from "react";
import defaultProfile from "../../../../assets/default-profile.avif";
import { useAuth } from "../../../../AuthContext";
import { api } from "../../../../api";

export function validateImage(file) {
  return file.size < 2 * 1024 * 1024; // 2 MB
}

const Profile = ({ entity }) => {
    const { user } = useAuth();
    const [isLoading, setIsLoading] = useState(true);
    const [profile, setProfile] = useState(null);

    useEffect(() => {
        const fetchData = async () => {
            if (user?.id) { // Ensure user exists
                try {
                    let url;
                    if (entity === "users") {
                        url = "/users/me";
                    } else if (entity === "brands") {
                        url = `/brands/${user.id}`;
                    }
                    const response = await api.get(url);
                    setProfile(response.data.data); // Update state
                    setIsLoading(false);
                } catch (err) {
                    console.error(err);
                }
            }
        };

        // Call it immediately
        fetchData();

    }, [user]);

    const [ selectedFile, setSelectedFile ] = useState(null);
    const [ preview, setPreview ] = useState(defaultProfile);


    if (isLoading) {
        return <div>Loading...</div>
    }
    
    function handlePhotoChange(e) {
        const file = e.target.files[0];
        if (!validateImage(file)) {
            return
        }
        const newPreview = URL.createObjectURL(file);
        setPreview(newPreview);
        console.log(newPreview);
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
                    {(entity === "users") && (
                        <div className="profile-form-section">
                            <div className="profile-form-group">
                            <label className="profile-form-label">First Name</label>
                            <input type="text" defaultValue="Jane" className="profile-form-input" value={profile.first_name} />
                            </div>
                            <div className="profile-form-group">
                            <label className="profile-form-label">Last Name</label>
                            <input type="text" defaultValue="Doe" className="profile-form-input" value={profile.last_name} />
                            </div>
                            <div className="profile-form-group">
                            <label className="profile-form-label">Email</label>
                            <input type="email" defaultValue="jane@example.com" className="profile-form-input" value={profile.email}/>
                            </div>
                            <div className="profile-form-group">
                            <label className="profile-form-label">Age</label>
                            <input type="text" defaultValue="0" className="profile-form-input" value={profile.age} />
                            </div>
                            <button className="profile-save-button">Save Changes</button>
                        </div>
                    )}
                    {(entity === "brands") && (
                        <div className="profile-form-section">
                            <div className="profile-form-group">
                            <label className="profile-form-label">Name</label>
                            <input type="text" placeholder="XYZ" readOnly={true} className="profile-form-input" value={profile.name} />
                            </div>
                            <div className="profile-form-group">
                            <label className="profile-form-label">Sector</label>
                            <input type="text" placeholder="Beauty" readOnly={true} className="profile-form-input" value={profile.sector} />
                            </div>
                            <div className="profile-form-group">
                            <label className="profile-form-label">Email</label>
                            <input type="email" defaultValue="jane@example.com" className="profile-form-input" value={profile.email}/>
                            </div>
                            <div className="profile-form-group">
                            <label className="profile-form-label">Address</label>
                            <input type="text-area" defaultValue="Where you are" className="profile-form-input" value={profile.address} />
                            </div>
                            <div className="profile-form-group">
                            <label className="profile-form-label">Website</label>
                            <input type="text" placeholder="https://" className="profile-form-input" value={profile.website} />
                            </div>
                            <button className="profile-save-button">Save Changes</button>
                        </div>
                    )}
                </div>
            </div>
            {(entity === "users") && (
                <div className="profile-social-card">
                    <div className="profile-social-header">
                        <h3 className="profile-social-title">Social Media Accounts</h3>
                        <button className="social-add-button">Add Link</button>
                    </div>
                    <div className="profile-social-list">
                    {profile.links.map((account, i) => (
                        <a key={i} className="profile-social-item" href={account.url}>
                            <div className="profile-social-item-left">
                                <div className="profile-social-avatar">
                                    {account.platform.slice(0, 2)}
                                </div>
                                <div className="profile-social-info">
                                    <p className="profile-social-platform">{account.platform}</p>
                                </div>
                            </div>
                            <button className={`profile-social-button profile-social-button-connected`}>
                                Connected
                            </button>
                        </a>
                    ))}
                    </div>
                </div>
            )}
        </div>
    )
};

export default Profile;