import { API_BASE_URL } from '@/config';
const initData = window.Telegram.WebApp.initData;

export async function roulettePlaceBet(Amount, Color) {
    try {
        const response = await fetch(`https://${API_BASE_URL}/games/roulettex14/place`, {
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
