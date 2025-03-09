import { API_BASE_URL, WS_PROTOCOL, API_PROTOCOL } from '@/config';
const initData = window.Telegram.WebApp.initData;

export async function getWheelWins() {
    try {
        const response = await fetch(`${API_PROTOCOL}://${API_BASE_URL}/games/fortunewheel/wins`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                'X-Telegram-Init-Data': initData,
            },
        });

        
        return await response;
    } catch (error) {
        console.error('Error registering user:', error);
        throw error;
    }
}
