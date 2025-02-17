import { API_BASE_URL } from '@/config';
const initData = window.Telegram.WebApp.initData;

export async function getWheelSpins() {
    try {
        const response = await fetch(`https://${API_BASE_URL}/games/fortunewheel/spins`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                'X-Telegram-Init-Data': initData,
            },
        });

        const data = await response.json();
        return data;
    } catch (error) {
        console.error('Error registering user:', error);
        throw error;
    }
}
