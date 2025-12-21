import "./messages.css";
import { useState, useEffect } from "react";

const Messages = () => {
    const [ conversations, setConversations ] = useState([
        {
          id: 1,
          partner: "EcoWear Brand",
          lastMessage: "Thanks for your interest in the campaign!",
          time: "2m ago",
          unread: true,
          active: true,
        },
        {
          id: 2,
          partner: "Tech Gadgets",
          lastMessage: "Can you send me your media kit?",
          time: "1h ago",
          unread: false,
          active: false,
        },
        {
          id: 3,
          partner: "HealthPlus",
          lastMessage: "The payment has been processed",
          time: "3h ago",
          unread: true,
          active: false,
        },
    ]);
    return (
        <div className="messages-container">
            {conversations.map(message => (
                <div key={message.id} className={`message-item ${message.unread ? 'message-item-unread' : ''}`}>
                <div className="message-item-wrapper">
                    <div className="message-item-left">
                    <div className="message-avatar">
                        {message.partner.slice(0, 2)}
                    </div>
                    <div className="message-content">
                        <div className="message-header">
                            <p className="message-brand-name">{message.partner}</p>
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