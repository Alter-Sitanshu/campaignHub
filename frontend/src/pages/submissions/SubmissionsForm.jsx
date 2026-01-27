import { useEffect, useState } from 'react';
import { X, Link, Video, CheckCircle, AlertCircle } from 'lucide-react';
import './SubmissionsForm.css';
import { api } from '../../api';
import { useAuth } from '../../AuthContext';

const SubmissionsForm = ({ isOpen, onClose, campaignId, onSuccess }) => {
  const [formData, setFormData] = useState({
    url: '',
    platform: 'youtube',
    status: 1,
  });

  const [errors, setErrors] = useState({});
  const [isValid, setIsValid] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [submitError, setSubmitError] = useState('');
  const [submitSuccess, setSubmitSuccess] = useState('');
  const { user } = useAuth();

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
    // Clear any previous messages
    if (submitError) setSubmitError('');
    if (submitSuccess) setSubmitSuccess('');
  };

  const validateUrl = (url, platform) => {
    if (!url.trim()) return 'Video URL is required';

    try {
      const urlObj = new URL(url);
      const hostname = urlObj.hostname.toLowerCase();

      if (platform === 'youtube') {
        if (!hostname.includes('youtube.com') && !hostname.includes('youtu.be')) {
          return 'Please enter a valid YouTube URL';
        }
      } else if (platform === 'instagram') {
        if (!hostname.includes('instagram.com')) {
          return 'Please enter a valid Instagram URL';
        }
      }
    } catch {
      return 'Please enter a valid URL';
    }

    return '';
  };

  const getErrors = () => {
    const newErrors = {};
    const urlError = validateUrl(formData.url, formData.platform);
    if (urlError) newErrors.url = urlError;
    return newErrors;
  };

  const handleSubmit = async () => {
    setIsLoading(true);
    setSubmitError('');
    setSubmitSuccess('');

    try {
      const payload = {
        creator_id: user?.id,
        campaign_id: campaignId,
        url: formData.url,
        status: 1, // Always send as number 1
      };

      const response = await api.post("/submissions", payload);

      if (response.status === 201) {
        setSubmitSuccess('Submission created successfully!');
        setFormData({
          url: '',
          platform: 'youtube',
          status: 1,
        });

        // Call onSuccess callback if provided (for refreshing parent component)
        if (onSuccess) {
          onSuccess();
        }

        // Close modal after a short delay to show success message
        setTimeout(() => {
          onClose();
          setSubmitSuccess('');
        }, 2000);
      } else {
        throw new Error('Failed to create submission');
      }
    } catch (error) {
      console.error('Submission error:', error);
      if (error.response?.data?.message) {
        setSubmitError(error.response.data.message);
      } else if (error.response?.status === 400) {
        setSubmitError('Invalid submission data. Please check your inputs.');
      } else if (error.response?.status === 401) {
        setSubmitError('You are not authorized to make this submission.');
      } else if (error.response?.status === 403) {
        setSubmitError('Access forbidden. Please check your permissions.');
      } else {
        setSubmitError('Failed to create submission. Please try again.');
      }
    } finally {
      setIsLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="modal-overlay">
      <div className="modal-container submissions-form">
        <div className="modal-header">
          <div className="modal-header-content">
            <div>
              <h2 className="modal-title">Submit Content</h2>
              <p className="modal-subtitle">Share your video content for the campaign</p>
            </div>
            <button onClick={onClose} className="modal-close-btn">
              <X className="modal-close-icon" />
            </button>
          </div>
        </div>

        <div className="modal-body">
          {/* Platform Selection */}
          <div className="platform-selection">
            <label className="form-label">Platform *</label>
            <div className="platform-buttons">
              <button
                type="button"
                onClick={() => setFormData(prev => ({ ...prev, platform: 'youtube' }))}
                className={`platform-btn ${formData.platform === 'youtube' ? 'active' : ''}`}
              >
                <Video className="platform-icon" />
                <span>YouTube</span>
              </button>
              <button
                type="button"
                onClick={() => setFormData(prev => ({ ...prev, platform: 'instagram' }))}
                className={`platform-btn ${formData.platform === 'instagram' ? 'active' : ''}`}
              >
                <Video className="platform-icon" />
                <span>Instagram</span>
              </button>
            </div>
          </div>

          {/* Video URL */}
          <div className="url-input-group">
            <label className="form-label">
              <div className="form-label-with-icon">
                <Link className="label-icon" />
                <span>Video URL *</span>
              </div>
            </label>
            <input
              type="url"
              name="url"
              value={formData.url}
              onChange={handleChange}
              placeholder={
                formData.platform === 'youtube'
                  ? "https://youtube.com/watch?v=... or https://youtu.be/..."
                  : "https://instagram.com/p/... or https://instagram.com/reel/..."
              }
              className={`url-input ${errors.url ? 'error' : ''}`}
            />
            {errors.url && (
              <div className="error-message">
                <AlertCircle size={16} />
                {errors.url}
              </div>
            )}
            <p className="helper-text">
              Enter the URL of your {formData.platform} video content
            </p>
          </div>
          <div className="form-group">
            <label className="form-label">Submission Status</label>
            <select
              name="status"
              value={formData.status}
              onChange={handleChange}
              className="form-select"
              disabled={true}
            >
              <option value={1}>Active</option>
            </select>
          </div>

          {/* Error/Success Messages */}
          {submitError && (
            <div className="error-message">
              <AlertCircle size={16} />
              {submitError}
            </div>
          )}
          {submitSuccess && (
            <div className="success-message">
              <CheckCircle size={16} />
              {submitSuccess}
            </div>
          )}

          {/* Footer Buttons */}
          <div className="form-footer">
            <button
              type="button"
              disabled={isLoading}
              onClick={onClose}
              className="btn-cancel"
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={handleSubmit}
              className="btn-submit"
              disabled={!isValid || isLoading}
            >
              {isLoading ? (
                <div className="loading-container">
                    <div className="loading-spinner"></div>
                    <p>Submitting...</p>
                </div>
              ) : (
                "Submit Content"
              )}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SubmissionsForm;
