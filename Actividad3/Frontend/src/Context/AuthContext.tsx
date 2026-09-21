import { createContext, useContext, useState, type ReactNode } from "react";

interface AuthContextType {
  isAuthenticated: boolean;
  userEmail: string | null;
  token: string | null;
  login: (email: string, token: string) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth debe ser usado dentro de un AuthProvider");
  }
  return context;
};

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider = ({ children }: AuthProviderProps) => {
  const [token, setToken] = useState<string | null>(() =>
    localStorage.getItem("token")
  );
  const [userEmail, setUserEmail] = useState<string | null>(() =>
    localStorage.getItem("userEmail")
  );
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(() =>
    Boolean(localStorage.getItem("token"))
  );

  const login = (email: string, authToken: string) => {
    setIsAuthenticated(true);
    setUserEmail(email);
    setToken(authToken);
    localStorage.setItem("token", authToken);
    localStorage.setItem("userEmail", email);
  };

  const logout = () => {
    setIsAuthenticated(false);
    setUserEmail(null);
    setToken(null);
    localStorage.removeItem("token");
    localStorage.removeItem("userEmail");
  };

  return (
    <AuthContext.Provider
      value={{ isAuthenticated, userEmail, token, login, logout }}
    >
      {children}
    </AuthContext.Provider>
  );
};