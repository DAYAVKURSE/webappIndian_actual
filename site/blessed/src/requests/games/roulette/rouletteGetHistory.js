import { API_BASE_URL, WS_PROTOCOL, API_PROTOCOL } from '@/config';

const initData = window.Telegram.WebApp.initData;

export async function rouletteGetHistory() {
    try {
        const response = await fetch(`${API_PROTOCOL}://${API_BASE_URL}/games/roulettex14/history`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                'X-Telegram-Init-Data': initData,
            },
        });

        return await response.json();
    } catch (error) {
        console.error('Error fetching game history:', error);
        throw error;
    }
}