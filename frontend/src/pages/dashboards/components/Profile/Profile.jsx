import "./profile.css";
import { useEffect, useState } from "react";
import defaultProfile from "../../../../assets/default-profile.avif";
import { useAuth } from "../../../../AuthContext";
import { api } from "../../../../api";
import { useNavigate } from "react-router-dom";

function validateImage(file) {
    return file.size < 2 * 1024 * 1024; // 2 MB
}

const Profile = ({ entity }) => {
    const { user } = useAuth();
    const [isLoading, setIsLoading] = useState(true);
    const [selectedFile, setSelectedFile] = useState(null);
    const [preview, setPreview] = useState(defaultProfile);

    const [profile, setProfile] = useState(null);
    const [form, setForm] = useState(null);

    const [Link, setLink] = useState({
        platform: "",
        url: "",
    });
    const navigate = useNavigate();
    const [popup, setPopup] = useState(false);

    // --------------------------
    //  Fetch Data
    // --------------------------
    useEffect(() => {
        if (user === null) {
            navigate("/");
        }
        const fetchData = async () => {
            if (!user?.id) return;

            try {
                let url =
                    entity === "users"
                        ? "/users/me"
                        : `/brands/${user.id}`;

                const response = await api.get(url);
                const data = response.data.data;

                // Set profile
                setProfile(data);

                // Initialize form based on entity
                if (entity === "users") {
                    setForm({
                        first_name: data.first_name,
                        last_name: data.last_name,
                        email: data.email,
                        age: data.age,
                    });
                } else {
                    setForm({
                        email: data.email,
                        address: data.address,
                        website: data.website,
                    });
                }

                setIsLoading(false);
            } catch (err) {
                console.error(err);
            }
        };

        fetchData();
    }, [user, entity]);

    if (isLoading || !profile || !form) {
        return <div>Loading...</div>;
    }

    // --------------------------
    //  Photo Change
    // --------------------------
    function handlePhotoChange(e) {
        const file = e.target.files[0];
        if (!file || !validateImage(file)) return;

        setSelectedFile(file);
        setPreview(URL.createObjectURL(file));
    }

    // --------------------------
    // Add social link
    // --------------------------
    function handleAddLink() {
        //TODO: handle later
        console.log(Link);
    }

    function handleOnChange(e) {
        let newVal = e.target.value;
        let targetKey = e.target.name;
        setForm({...form, [targetKey]: newVal});
    }

    // --------------------------
    // Upload photo + update profile
    // --------------------------
    async function handleProfileChange() {
        try {
            // 1) Upload Photo (if selected)
            if (selectedFile) {
                const ext = selectedFile.name.split(".").pop();
                const response = await api.get(`/${entity}/profile_picture/?ext=${ext}`);

                const uploadUrl = response.data.data.uploadUrl;
                const objKey = response.data.data.objKey;

                // Upload file to signed URL
                await fetch(uploadUrl, {
                    method: "PUT",
                    headers: {
                        "Content-Type": selectedFile.type,
                    },
                    body: selectedFile,
                });

                // Inform backend that upload succeeded
                await api.post(`/${entity}/profile_picture/confirm`, {
                    objectKey: objKey,
                });
            }

            // 2) Find changed fields
            const updates = {};
            for (const key in form) {
                if (form[key] !== profile[key]) {
                    updates[key] = form[key];
                }
            }

            // 3) Send profile updates ONLY if something changed
            if (Object.keys(updates).length > 0) {
                await api.patch(`/${entity}/${user.id}`, updates);
            }

            alert("Profile updated!");
        } catch (err) {
            console.error(err);
            alert("Error updating profile!");
        }
    }
    return (
        <div>
            {popup && (
                <div className="addlink-popup-overlay">
                    <div className="addlink-popup-card">

                        <h3 className="popup-title">Add Social Link</h3>

                        <div className="link-group">
                            <select
                                id="platform-select"
                                value={Link.platform}
                                onChange={(e) => setLink({ ...Link, platform: e.target.value })}
                                className="link-select"
                            >
                                <option value="">Select platform</option>
                                <option value="instagram">Instagram</option>
                                <option value="youtube">YouTube</option>
                                <option value="twitter">Twitter</option>
                                <option value="tiktok">TikTok</option>
                                <option value="linkedin">LinkedIn</option>
                            </select>

                            <input
                                type="url"
                                placeholder="https://..."
                                value={Link.url}
                                onChange={(e) => setLink({ ...Link, url: e.target.value })}
                                className="link-input"
                            />
                        </div>

                        <div className="popup-actions">
                            <button className="popup-cancel" onClick={() => setPopup(false)}>
                                Cancel
                            </button>
                            <button className="popup-save" onClick={handleAddLink}>
                                Save
                            </button>
                        </div>
                    </div>
                </div>
            )}
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
                            <input type="text" placeholder="Jane" onChange={handleOnChange} className="profile-form-input" value={profile.first_name} />
                            </div>
                            <div className="profile-form-group">
                            <label className="profile-form-label">Last Name</label>
                            <input type="text" placeholder="Doe" onChange={handleOnChange} className="profile-form-input" value={profile.last_name} />
                            </div>
                            <div className="profile-form-group">
                            <label className="profile-form-label">Email</label>
                            <input type="email" placeholder="jane@example.com" onChange={handleOnChange} className="profile-form-input" value={profile.email}/>
                            </div>
                            <div className="profile-form-group">
                            <label className="profile-form-label">Age</label>
                            <input type="text" placeholder="0" onChange={handleOnChange} className="profile-form-input" value={profile.age} />
                            </div>
                            <button className="profile-save-button">Save Changes</button>
                        </div>
                    )}
                    {(entity === "brands") && (
                        <div className="profile-form-section">
                            <div className="profile-form-group">
                            <label className="profile-form-label">Name</label>
                            <input id="name" type="text" placeholder="XYZ" readOnly={true} className="profile-form-input" value={profile.name} />
                            </div>
                            <div className="profile-form-group">
                            <label className="profile-form-label">Sector</label>
                            <input id="sector" type="text" placeholder="Beauty" readOnly={true} className="profile-form-input" value={profile.sector} />
                            </div>
                            <div className="profile-form-group">
                            <label className="profile-form-label">Email</label>
                            <input type="email" placeholder="jane@example.com" className="profile-form-input" name="email"  onChange={handleOnChange} value={profile.email}/>
                            </div>
                            <div className="profile-form-group">
                            <label className="profile-form-label">Address</label>
                            <input type="text-area" placeholder="Where you are" className="profile-form-input" name="address" onChange={handleOnChange} value={profile.address} />
                            </div>
                            <div className="profile-form-group">
                            <label className="profile-form-label">Website</label>
                            <input type="text" placeholder="https://" className="profile-form-input" name="website" onChange={handleOnChange} value={profile.website} />
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
                        <button className="social-add-button" onClick={() => setPopup(true)}>Add Link</button>
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