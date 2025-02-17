import { API_BASE_URL } from '@/config';
const initData = window.Telegram.WebApp.initData;

export async function getWheelWins() {
    try {
        const response = await fetch(`https://${API_BASE_URL}/games/fortunewheel/wins`, {
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
