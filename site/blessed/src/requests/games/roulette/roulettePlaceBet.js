import { API_BASE_URL, WS_PROTOCOL, API_PROTOCOL } from '@/config';
const initData = window.Telegram.WebApp.initData;

export async function roulettePlaceBet(Amount, Color) {
    try {
        const response = await fetch(`${API_PROTOCOL}://${API_BASE_URL}/games/roulettex14/place`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Telegram-Init-Data': initData,
            },
            body: JSON.stringify({ 
                Amount: Amount,
                Color: Color,
            }),
        });

        return await response;
    } catch (error) {
        console.error('Error registering user:', error);
        throw error;
    }
}
