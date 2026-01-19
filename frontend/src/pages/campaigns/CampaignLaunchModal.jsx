import { useEffect, useState } from 'react';
import { X, FileText, IndianRupee, Target, Video } from 'lucide-react';
import './campaignLaunch.css';
import { api } from '../../api';
import { useNavigate } from 'react-router-dom';
import { useQueryClient } from "@tanstack/react-query";


const CampaignLaunchModal = ({ isOpen, onClose, brandId, fillForm }) => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [formData, setFormData] = useState({
    title: '',
    requirements: '',
    platform: 'youtube',
    doc_link: '',
    status: 0,
    ...fillForm,
    budget: fillForm?.budget?.toString() ?? '',
    cpm: fillForm?.cpm?.toString() ?? ''
  });

  const [errors, setErrors] = useState({});
  const [isValid, setIsValid] = useState(false || fillForm !== null);
  const [ isLoading, setIsLoading ] = useState(false);

  useEffect(() => {
    let errs = getErrors();
    setErrors(errs);
    setIsValid(Object.keys(errs).length === 0);
  }, [formData]);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData({
      ...formData,
      [name]: value
    });
    if (errors[name]) {
      setErrors({ ...errors, [name]: '' });
    }
  };

  const getErrors = () => {
    const newErrors = {};
    
    if (!formData.title.trim()) newErrors.title = 'campaign title is required';
    if (!formData.budget || parseFloat(formData.budget) <= 0) newErrors.budget = 'valid budget is required';
    if (!formData.cpm || parseFloat(formData.cpm) <= 0) newErrors.cpm = 'valid CPM is required';
    if (!formData.requirements.trim()) newErrors.requirements = 'requirements are required';
    if (!formData.doc_link.trim()) newErrors.doc_link = 'document link is required';
    
    return newErrors;
  };


  const handleSubmit = async () => {
    setIsLoading(true);
    if (fillForm === null) {
      // This is not an edit and a upload request form
      // upload to the backend
      const payload = {
        brand_id: brandId,
        title: formData.title,
        budget: parseFloat(formData.budget),
        cpm: parseFloat(formData.cpm),
        requirements: formData.requirements,
        platform: formData.platform,
        doc_link: formData.doc_link,
        status: parseInt(formData.status),
      };
      const response = await api.post("/campaigns", payload);
      if(response.status != 200) {
        navigate(`/errors/${response.status}`);
      }
      setFormData({
        title: '',
        requirements: '',
        platform: 'youtube',
        doc_link: '',
        status: 0,
        budget: '',
        cpm: '',
      });
    }
    queryClient.invalidateQueries({ queryKey: ['brandCampaignFeed'] });
    setIsLoading(false);
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="modal-overlay">
      <div className="modal-container">
        <div className="modal-header">
          <div className="modal-header-content">
            <div>
              <h2 className="modal-title">Launch New Campaign</h2>
              <p className="modal-subtitle">Create and customize your campaign details</p>
            </div>
            <button onClick={onClose} className="modal-close-btn">
              <X className="modal-close-icon" />
            </button>
          </div>
        </div>

        <div className="modal-body">
          {/* Campaign Title */}
          <div className="form-group">
            <label className="form-label">Campaign Title *</label>
            <input
              type="text"
              name="title"
              value={formData.title}
              onChange={handleChange}
              placeholder="e.g., Winter Sale Awareness Campaign"
              className={`form-input campaign ${errors.title ? 'error' : ''}`}
            />
            {errors.title && <p className="error-message">{errors.title}</p>}
          </div>

          {/* Platform Selection */}
          <div className="form-group">
            <label className="form-label">Platform *</label>
            <div className="platform-grid">
              <button
                type="button"
                onClick={() => setFormData(prev => ({ ...prev, platform: 'instagram' }))}
                className={`platform-button ${formData.platform === 'instagram' ? 'active' : ''}`}
              >
                <Video className="platform-icon" />
                <span>Instagram</span>
              </button>
              <button
                type="button"
                onClick={() => setFormData(prev => ({ ...prev, platform: 'youtube' }))}
                className={`platform-button ${formData.platform === 'youtube' ? 'active' : ''}`}
              >
                <Video className="platform-icon" />
                <span>YouTube</span>
              </button>
            </div>
          </div>

          {/* Budget and CPM */}
          <div className="input-grid">
            <div className="form-group">
              <label className="form-label">
                <div className="form-label-with-icon">
                  <IndianRupee className="label-icon" />
                  <span>Budget (₹) *</span>
                </div>
              </label>
              <input
                type="number"
                name="budget"
                value={formData.budget}
                onChange={handleChange}
                placeholder="50000"
                min="0"
                step="1000"
                className={`form-input campaign ${errors.budget ? 'error' : ''}`}
              />
              {errors.budget && <p className="error-message">{errors.budget}</p>}
            </div>

            <div className="form-group">
              <label className="form-label">
                <div className="form-label-with-icon">
                  <Target className="label-icon" />
                  <span>CPM (₹) *</span>
                </div>
              </label>
              <input
                type="number"
                name="cpm"
                value={formData.cpm}
                onChange={handleChange}
                placeholder="120"
                min="0"
                step="10"
                className={`form-input campaign ${errors.cpm ? 'error' : ''}`}
              />
              {errors.cpm && <p className="error-message">{errors.cpm}</p>}
            </div>
          </div>

          {/* Requirements */}
          <div className="form-group">
            <label className="form-label">Campaign Requirements *</label>
            <textarea
              name="requirements"
              value={formData.requirements}
              onChange={handleChange}
              placeholder="e.g., Create 1 reel + 2 story posts showcasing product features..."
              rows="4"
              className={`form-textarea ${errors.requirements ? 'error' : ''}`}
            />
            {errors.requirements && <p className="error-message">{errors.requirements}</p>}
          </div>

          {/* Document Link */}
          <div className="form-group">
            <label className="form-label">
              <div className="form-label-with-icon">
                <FileText className="label-icon" />
                <span>Campaign Document Link *</span>
              </div>
            </label>
            <input
              type="url"
              name="doc_link"
              value={formData.doc_link}
              onChange={handleChange}
              placeholder="https://docs.google.com/..."
              className={`form-input campaign ${errors.doc_link ? 'error' : ''}`}
            />
            {errors.doc_link && <p className="error-message">{errors.doc_link}</p>}
            <p className="helper-text">Provide a link to detailed campaign brief</p>
          </div>

          {/* Status */}
          <div className="form-group">
            <label className="form-label">Campaign Status *</label>
            <select
              name="status"
              value={formData.status}
              onChange={handleChange}
              className="form-select"
              disabled={fillForm !== null}
            >
              <option value={0}>Draft</option>
              <option value={1}>Active</option>
            </select>
          </div>

          {/* Footer Buttons */}
          <div className="modal-footer">
            <button type="button" disabled={isLoading} onClick={onClose} className="btn btn-cancel">
              Cancel
            </button>
            <button
              type="button"
              onClick={handleSubmit}
              className="btn btn-submit"
              disabled={!isValid || isLoading}
            >
              {isLoading ? "Submitting..." : fillForm == null ? "Launch Campaign" : "Save"}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default CampaignLaunchModal;

// const Demo = () => {
//   const [isModalOpen, setIsModalOpen] = useState(false);

//   return (
//     <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 p-8">
//       <div className="max-w-4xl mx-auto">
//         <h1 className="text-3xl font-bold text-gray-800 mb-6">Brand Dashboard</h1>
//         <button
//           onClick={() => setIsModalOpen(true)}
//           className="px-8 py-4 bg-gradient-to-r from-emerald-600 to-teal-600 text-white font-semibold rounded-xl hover:from-emerald-700 hover:to-teal-700 transition-all shadow-lg hover:shadow-xl transform hover:-translate-y-0.5"
//         >
//           + Launch New Campaign
//         </button>

//         <CampaignLaunchModal
//           isOpen={isModalOpen}
//           onClose={() => setIsModalOpen(false)}
//           brandId="brand_123"
//         />
//       </div>
//     </div>
//   );
// };

// export default Demo;