import "./messages.css";

const Messages = ({ messages }) => {
    return (
        <div className="messages-container">
            {messages.map(message => (
                <div key={message.id} className={`message-item ${message.unread ? 'message-item-unread' : ''}`}>
                <div className="message-item-wrapper">
                    <div className="message-item-left">
                    <div className="message-avatar">
                        {message.brand.slice(0, 2)}
                    </div>
                    <div className="message-content">
                        <div className="message-header">
                            <p className="message-brand-name">{message.brand}</p>
                            {message.unread && <span className="message-unread-indicator"></span>}
                        </div>
                        <p className="message-preview">{message.preview}</p>
                    </div>
                    </div>
                    <span className="message-time">{message.time}</span>
                </div>
                </div>
            ))}
        </div>
    )
};

export default Messages;