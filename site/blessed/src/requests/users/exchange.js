import { API_BASE_URL } from '@/config';
const initData = window.Telegram.WebApp.initData;

export async function exchange(AmountBcoins) {
    try {
        const response = await fetch(`https://${API_BASE_URL}/users/exchange`, {
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
