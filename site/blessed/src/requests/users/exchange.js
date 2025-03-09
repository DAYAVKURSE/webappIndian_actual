import { API_BASE_URL, WS_PROTOCOL, API_PROTOCOL } from '@/config';
const initData = window.Telegram.WebApp.initData;

export async function exchange(AmountBcoins) {
    try {
        const response = await fetch(`${API_PROTOCOL}://${API_BASE_URL}/users/exchange`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Telegram-Init-Data': initData,
            },
            body: JSON.stringify({ 
                AmountBcoins: AmountBcoins
            }),
        });
        
        return await response;
    } catch (error) {
        console.error('Error registering user:', error);
        throw error;
    }
}
