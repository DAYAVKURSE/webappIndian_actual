import styles from "./Loading.module.scss";
import { useNavigate } from "react-router-dom";
import { useEffect } from "react";
import { auth } from "@/requests";
import useStore from "@/store";

export const Loading = () => {
    const navigate = useNavigate();
    const { setReferredBy } = useStore();

    useEffect(() => {
        const getReferralFromURL = () => {
            const params = new URLSearchParams(window.location.search);
            return params.get("referral");
        };

        async function checkAuth() {
            const referral = getReferralFromURL();
            
            if (referral) {
                setReferredBy(referral);
            }

            const response = await auth();

            if (response.status === 200) {
                navigate("/clicker");
            } else {
                navigate("/onboarding");
            }
        }

        checkAuth();
    }, [navigate, setReferredBy]);

    return (
        <div className={styles.loading}>
        </div>
    );
};
