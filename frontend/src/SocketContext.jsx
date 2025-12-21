import { createContext, useContext, useEffect, useState } from "react";
import { socket } from "./sockets";
import { useAuth } from "./AuthContext";

const SocketContext = createContext(null);

export const SocketProvider = ({ children }) => {
  const { user, loading } = useAuth();
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    if (loading) return;

    if (user) {
      socket.connect();

      socket.on("connect", () => {
        console.log("Socket connected");
        setConnected(true);
      });

      socket.on("disconnect", () => {
        console.log("Socket disconnected");
        setConnected(false);
      });
    }

    return () => {
      socket.off("connect", () => {});
      socket.off("disconnect", () => {});

      if (socket && socket.connected) {
        socket.close();
      }
    };
  }, [user, loading]);

  return (
    <SocketContext.Provider value={{ socket, connected }}>
      {children}
    </SocketContext.Provider>
  );
};

export const useSocket = () => {
  const ctx = useContext(SocketContext);
  if (!ctx) {
    throw new Error("useSocket must be used within SocketProvider");
  }
  return ctx;
};
