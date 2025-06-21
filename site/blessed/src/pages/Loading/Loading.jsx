import styles from './Loading.module.scss';
import { useNavigate } from 'react-router-dom';
import { useEffect } from 'react';
import useStore from '@/store';
import { logout } from '../../requests';
import { getAccessToken } from '../../utils/token-storage';

export const Loading = () => {
  const navigate = useNavigate();
  const { setReferredBy } = useStore();

  useEffect(() => {
    const getReferralFromURL = () => {
      const params = new URLSearchParams(window.location.search);
      return params.get('referral');
    };

    async function checkAuth() {
      const referral = getReferralFromURL();

      if (referral) {
        setReferredBy(referral);
      }

      const token = getAccessToken();

      if (token) {
        navigate('/clicker');
      } else {
        navigate('/onboarding');
      }
    }

    checkAuth();
  }, [navigate, setReferredBy]);

  return <div className={styles.loading}></div>;
};
