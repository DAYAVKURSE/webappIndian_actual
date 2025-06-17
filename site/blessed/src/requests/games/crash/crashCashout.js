import { API_BASE_URL } from '@/config';
const initData = window.Telegram?.WebApp?.initData || '';

export async function crashCashout() {
    if (!initData) {
        console.error('Telegram WebApp initData is missing');
        return { 
            ok: false, 
            status: 401,
            json: async () => ({ error: 'Authorization error. Telegram WebApp initData is missing.' })
        };
    }

    try {
        console.log('Sending cashout request');
        console.log('URL:', `https://${API_BASE_URL}/games/crashgame/cashout`);
        
        const response = await fetch(`https://${API_BASE_URL}/games/crashgame/cashout`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Telegram-Init-Data': initData,
            },
        });

        console.log('Received response with status:', response.status);
        
        return response;
    } catch (error) {
        console.error('Error during cashout:', error);
        return { 
            ok: false, 
            status: 500,
            json: async () => ({ error: 'Network error during cashout.' })
        };
    }
}
