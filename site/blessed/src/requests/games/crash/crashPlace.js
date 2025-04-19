import { API_BASE_URL } from '@/config';
const initData = window.Telegram?.WebApp?.initData || '';

export async function crashPlace(Amount, CashOutMultiplier) {
    if (!initData) {
        console.error('Telegram WebApp initData is missing');
        return { 
            ok: false, 
            status: 401,
            json: async () => ({ error: 'Authorization error. Telegram WebApp initData is missing.' })
        };
    }

    try {
        const requestBody = { 
            Amount: Number(Amount),
            CashOutMultiplier: CashOutMultiplier !== undefined ? Number(CashOutMultiplier) : undefined
        };

        console.log('Sending bet request:', requestBody);
        console.log('URL:', `https://${API_BASE_URL}/games/crashgame/place`);
        
        const response = await fetch(`https://${API_BASE_URL}/games/crashgame/place`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Telegram-Init-Data': initData,
            },
            body: JSON.stringify(requestBody),
        });

        console.log('Received response with status:', response.status);
        
        return response;
    } catch (error) {
        console.error('Error placing bet:', error);
        return { 
            ok: false, 
            status: 500,
            json: async () => ({ error: 'Network error while placing bet.' })
        };
    }
}
