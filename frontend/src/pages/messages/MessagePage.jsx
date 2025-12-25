import { useState, useRef, useEffect } from "react";
import { ArrowLeft, Paperclip, Send } from "lucide-react";
import "./messaging.css";
import { api } from "../../api";
import { useAuth } from "../../AuthContext";
import { data, useNavigate } from "react-router-dom";
import { useSocket } from "../../SocketContext";

const MessagePage = () => {
  const [ conversations, setConversations ] = useState(null);
  const { user } = useAuth();
  const { socket, connected } = useSocket();
  const [ messages, setMessages ] = useState(null);
  const navigate = useNavigate();
  const [ isLoading, setIsLoading ] = useState(true);
  const [ loadingMessages, setLoadingMessages ] = useState(true);
  const [ activeConv, setactiveConv ] = useState(null);
  const [input, setInput] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const messagesEndRef = useRef(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  const fetchConversations = async () => {
    const response = await api.get("/private/conversations");
    if (response.status != 200 ) {
      navigate(`/errors/${response.status}`);
    }
    setIsLoading(false);
    const data = response.data.data;
    setConversations(data);
  }

  function groupMessagesByDate(msgArr) {
    const groups = {};
    if (!msgArr || msgArr.length == 0) return [];
    msgArr.forEach((msg) => {
      const dateKey = new Date(msg.created_at).toDateString();
      if (!groups[dateKey]) groups[dateKey] = [];
      groups[dateKey].push(msg);
    });

    return Object.entries(groups).map(([date, msgs]) => ({
      date,
      messages: msgs.sort(
        (a, b) => new Date(a.created_at) - new Date(b.created_at)
      ),
    }));
  }

  const loadMessages = async (conversationId, timestamp = "", cursor = "") => {
    if (activeConv?.id === conversationId & messages !== null) return;
    setLoadingMessages(true);
    let endpoint = `/private/conversations/${conversationId}/messages`;
    if (timestamp !== "") {
      endpoint += `?timestamp=${timestamp},cursor=${cursor}`;
    }
    const response = await api.get(endpoint);
    if (response.status != 200 && response.status != 204) {
      navigate(`/errors/${response.status}`);
    } 
    if (response.status == 204) {
      setMessages({
        messages: [],
        meta: {},
      });
    } else {
      setMessages(response.data.data);
    }
    setLoadingMessages(false);
  };

  useEffect(() => {
    fetchConversations();
  }, [])

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  useEffect(() => {
    if (!connected || !activeConv) return;

    // We keep a lightweight "join" event locally (backend doesn't require it, but
    // frontend can use it for presence or to manage UI state server-side later)
    socket.sendEvent("conversation:join", {
      conversation_id: activeConv.id,
    });

    loadMessages(activeConv.id);

    const handler = (msg) => {
      // backend will send payload.message where appropriate – normalize here
      const conversationId = msg.conversation_id || msg.conversationID || (msg.message && msg.message.conversation_id) || (msg.message && msg.message.conversationID);
      if (conversationId === activeConv?.id) {
        setMessages((prev) => {
          return {
            ...prev,
            messages: [...(prev?.messages ?? []), msg.message ? msg.message : msg]
          }
        })
      }
    };

    const ackHandler = (serverMsg) => {
      // serverMsg is the saved message object (contains client_id which is the temp id)
      const tempId = serverMsg.client_id || serverMsg.clientID;
      if (!tempId) return;
      setMessages((prev) => {
        if (!prev) return prev;
        return {
          ...prev,
          messages: prev.messages?.map((m) => {
            if (m.id === tempId) {
              return {
                ...m,
                id: serverMsg.id || m.id,
                created_at: serverMsg.created_at || m.created_at,
                sending: false,
                // reflect any other persisted fields
                is_read: serverMsg.is_read ?? m.is_read,
              };
            }
            return m;
          })}
        });
    };

    socket.on("message:new", handler);
    socket.on("message:ack", ackHandler);

    return () => {
      socket.off("message:new", handler);
      socket.off("message:ack", ackHandler);
      socket.sendEvent("conversation:leave", {
        conversation_id: activeConv.id,
      });
    };
  }, [activeConv, connected]);

  const sendMessage = () => {
    if (!input.trim()) return;
    const client_id = "temp-" + new Date().toISOString();
    const incoming = {
      client_id: client_id,
      type: "chat_message", // backend expects this type
      conversation_id: activeConv?.id,
      content: input,
      message_type: "txt",
    };

    // Optimistic UI update
    setMessages((prev) => {
      return {
        ...prev,
        messages: [...(prev?.messages ?? []), {
          id: client_id, // client-generated temp id
          conversation_id: activeConv?.id,
          sender_id: user?.id,
          message_type: "txt",
          content: input,
          is_read: false,
          created_at: new Date().toISOString(),
          sending: true
        }],
      };
    });

    setInput("");
    socket.sendEvent("message", incoming);
  };

  const formatDate = (dateString) => {
      const date = new Date(dateString);
      return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  };

  function formatRelativeTime(isoString) {
    const now = new Date();
    const past = new Date(isoString);

    const diffMs = now - past;
    const diffSec = Math.floor(diffMs / 1000);

    if (diffSec < 60) return "just now";

    const diffMin = Math.floor(diffSec / 60);
    if (diffMin < 60) return `${diffMin}m ago`;

    const diffHr = Math.floor(diffMin / 60);
    if (diffHr < 24) return `${diffHr}h ago`;

    const diffDay = Math.floor(diffHr / 24);
    if (diffDay === 1) return "yesterday";
    if (diffDay < 7) return `${diffDay}d ago`;

    const diffWeek = Math.floor(diffDay / 7);
    if (diffWeek < 4) return `${diffWeek}w ago`;

    return past.toLocaleDateString(); // fallback for old messages
  }

  if (isLoading) {
    return (
      <main className="messaging-container empty-state">
        <div className="empty-message">
          <h3>Loading conversations...</h3>
        </div>
      </main>
    )
  }

  return (
    <div className="messaging-page">
      {/* Sidebar - Conversation List */}
      <aside className="messaging-sidebar">
        <div className="sidebar-header">
          <h2>Messages</h2>
        </div>

        <div className="sidebar-search">
          <input
            type="text"
            placeholder="Search conversations..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
          <a href={`/${user.entity}/dashboard/${user.id}`} className="home-button">Home</a>
        </div>

        <div className="conversation-list">
          {conversations?.map((conv) => (
            <div
              key={conv.id}
              className={`conversation-item ${
                activeConv?.id === conv.id ? "active" : ""
              }`}
              onClick={() => {setactiveConv(conv)}}
            >
              <div className="conversation-avatar">{conv.participant_name.slice(0,2)}</div>
              <div className="conversation-info">
                <h4 className="conversation-name">{conv.participant_name}</h4>
                <p className="conversation-preview">{conv.last_message}</p>
              </div>
              <div className="conversation-meta">
                <span className="conversation-time">{formatRelativeTime(conv.last_message_at)}</span>
              </div>
            </div>
          ))}
        </div>
      </aside>

      {/* Main Chat Area */}
      {!activeConv ? (
        <main className="messaging-container empty-state">
          <div className="empty-message">
            <h3>Select a conversation</h3>
            <p>Choose a conversation from the left to start messaging.</p>
          </div>
        </main>
      ) : (
        <main className="messaging-container">
          {/* Header */}
          <div className="messaging-header">
            <div className="header-left">
              <a className="back-button" href={`${user.entity}/dashboard/${user.id}`}>
                <ArrowLeft />
              </a>
              <div className="header-avatar">
                {activeConv?.participant_name.slice(0,2)}
              </div>
              <div className="header-info">
                <h3>{activeConv?.participant_name}</h3>
                <p className="header-status">{formatRelativeTime(activeConv?.last_message_at)}</p>
              </div>
            </div>
          </div>

          {/* Messages */}
          <div className="messaging-body">
            {loadingMessages ? <div className="empty-message">
              <h3>Loading your messages...</h3>
              <p>Please wait while we load your messages</p>
            </div>
            : groupMessagesByDate(messages?.messages).map((group) => (
              <div key={group.date}>
                <div className="date-divider">
                  <span>{formatDate(group.date)}</span>
                </div>

                {group.messages.map((msg) => (
                  <div
                    key={msg.id}
                    className={`message-bubble ${
                      msg.sender_id === user?.id ? "own" : "other"
                    } ${msg.sending ? "sending" : ""}`}
                  >
                    {msg.content}
                    <span className={`chat-time ${msg.own ? "own" : "other"}`}>{
                      new Date(msg.created_at).toLocaleTimeString(
                        "en-US",
                        { hour: "numeric", minute: "numeric"}
                    )}</span>
                  </div>
                ))}
              </div>
            ))}
            <div ref={messagesEndRef} />
          </div>

          {/* Input */}
          <div className="messaging-input">
            <div className="input-wrapper">
              <button className="attachment-btn">
                <Paperclip />
              </button>
              <input
                type="text"
                placeholder="Type a message..."
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && sendMessage()}
              />
            </div>
            <button
              className="send-button"
              onClick={sendMessage}
              disabled={!input.trim()}
            >
              <Send />
              Send
            </button>
          </div>
        </main>
      )}
    </div>
  );
};

export default MessagePage;