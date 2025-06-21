import { Navigate, Outlet } from 'react-router-dom';
import { useEffect, useState } from 'react';
import { getMe } from '../requests/users/getMe';
import { getAccessToken, removeAccessToken } from '../utils/token-storage';

export const PrivateRoute = () => {
  const [auth, setAuth] = useState(null);

  useEffect(() => {
    if (getAccessToken()) {
      setAuth(true);
      return;
    }

    getMe()
      .then(() => setAuth(true))
      .catch(() => {
        removeAccessToken();
        setAuth(false);
      });
  }, []);

  if (auth === null) {
    return <div>Loading...</div>;
  }

  if (!auth) {
    return <Navigate to="/login" replace />;
  }

  return <Outlet/>;
};
