import { API_BASE_URL } from '@/config';
const initData = window.Telegram.WebApp.initData;

export async function crashCashout() {
    try {
        const response = await fetch(`https://${API_BASE_URL}/games/crashgame/cashout`, {
            method: 'POST',
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
