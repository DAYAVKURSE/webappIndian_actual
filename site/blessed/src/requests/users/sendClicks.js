import { API_BASE_URL, WS_PROTOCOL, API_PROTOCOL } from '@/config';
import useStore from '@/store';
const initData = window.Telegram.WebApp.initData;

export async function sendClicks(Clicks, BPC) {
    try {
        const response = await fetch(`${API_PROTOCOL}://${API_BASE_URL}/clicker`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Telegram-Init-Data': initData,
            },
            body: JSON.stringify({ 
                ClicksCount: Clicks,
                BiPerClick: BPC
            }),
        });

        const contentLength = response.headers.get('Content-Length');
        if (contentLength && parseInt(contentLength) > 0) {
            const data = await response.json();
            if (data && data.BiPerClick) {
                useStore.setState({ BiPerClick: data.BiPerClick * data.BonusMultiplier });
            }
            return data;
        } else {
            return null;
        }
    } catch (error) {
        console.error('Error registering user:', error);
        throw error;
    }
}
